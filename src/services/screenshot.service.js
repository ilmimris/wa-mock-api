const wkhtmltoimage = require('wkhtmltoimage');
const path = require('path');
const fs = require('fs/promises');
const { ApiError } = require('../middleware/error.middleware');
const { convertWhatsAppToHTML } = require('../utils/whatsapp-html');

class ScreenshotService {
  constructor() {
    // In Docker, we rely on the default PATH finding /usr/bin/wkhtmltoimage
    // and the DISPLAY environment variable being set for Xvfb support.
    // No explicit command setting is needed unless overriding.
    if (process.env.WKHTMLTOIMAGE_PATH) {
        wkhtmltoimage.setCommand(process.env.WKHTMLTOIMAGE_PATH);
    }

    this.templatePath = path.join(__dirname, '../templates/whatsapp-chat.html');
    this.chatTemplate = null; // Initialize chatTemplate property
    this.loadTemplate().catch(err => {
      console.error("Failed to load template on startup:", err);
    });
  }

  async loadTemplate() {
    // Load HTML template if not already loaded
    if (!this.chatTemplate) {
      try {
        console.log('Loading HTML template...');
        this.chatTemplate = await fs.readFile(this.templatePath, 'utf-8');
        console.log('HTML template loaded successfully.');
      } catch (error) {
        console.error('Failed to load HTML template:', error);
        throw error;
      }
    }
  }

  /**
   * Generate a WhatsApp-style chat screenshot from messages
   * @param {Array} messages - Array of message objects
   * @param {Object} options - Screenshot options
   * @returns {Promise<string>} Base64 encoded image
   */
  async generateWhatsAppScreenshot(messages, options = {}) {
    try {
      const { width = 400, format = 'png', quality = 'high', headerDisplay = 'phone' } = options;

      // Ensure template is loaded
      if (!this.chatTemplate) {
        await this.loadTemplate();
      }

      // Generate HTML content
      const htmlContent = await this.generateChatHTML(messages, { width, headerDisplay });

      // wkhtmltoimage options
      const screenshotOptions = {
        width: parseInt(width, 10),
        format: format === 'jpeg' ? 'jpg' : format, // wkhtmltoimage uses jpg
        quality: quality === 'high' ? 90 : quality === 'medium' ? 70 : 50,
        // Disable smart width to force the specified width
        'disable-smart-width': true,
        // Using - for stdout
      };

      // wkhtmltoimage wrapper returns a stream
      return new Promise((resolve, reject) => {
        const stream = wkhtmltoimage.generate(htmlContent, screenshotOptions);

        const chunks = [];
        stream.on('data', (chunk) => chunks.push(chunk));
        stream.on('end', () => {
          const buffer = Buffer.concat(chunks);
          const base64Image = buffer.toString('base64');
          resolve(`data:image/${format};base64,${base64Image}`);
        });
        stream.on('error', (err) => {
          console.error('Error generating screenshot with wkhtmltoimage:', err);
          reject(new ApiError(500, 'Failed to generate screenshot'));
        });
      });

    } catch (error) {
      console.error('Error generating screenshot:', error);
      throw new ApiError(500, 'Failed to generate screenshot');
    }
  }

  /**
   * Generate HTML content for the chat
   * @private
   */
  async generateChatHTML(messages, options = {}) {
    try {
      const { width, headerDisplay } = options;

      if (!this.chatTemplate) {
        await this.loadTemplate();
      }
      let template = this.chatTemplate;

      // Extract recipient info from the first message
      const firstMessage = messages[0] || {};
      const recipientName = firstMessage.recipient_name || 'Customer';
      let recipientPhone = firstMessage.recipient_phone || 'Unknown';
      let headerLineText;

      if (headerDisplay === 'name') {
        headerLineText = recipientName;
      } else {
        // Format recipient phone number to add +62 prefix if it's not already there
        if (!recipientPhone.startsWith('+62')) {
          if (!recipientPhone.startsWith('62')) {
            recipientPhone = `+62 ${recipientPhone}`;
          } else {
            recipientPhone = `+62 ${recipientPhone.slice(2)}`;
          }
        } else {
          // Add space after +62 if space is not already there
          if (!recipientPhone.includes(' ')) {
            recipientPhone = recipientPhone.replace('+62', '+62 ');
          }
        }
        // Format to add dash after every 4 digits
        recipientPhone = recipientPhone.replace(/(?=\d{4}(?:\d{4})*$)/g, '-');
        headerLineText = recipientPhone;
      }

      const lastSeen = new Date().toLocaleTimeString('id-ID', {
        timeZone: "Asia/Jakarta",
        hour: '2-digit',
        minute: '2-digit',
        hour12: true
      });

      // Generate messages HTML
      const messagesHTML = messages.map(msg => {
        const isBot = msg.sender === 'Bot';
        const time = new Date(msg.timestamp).toLocaleTimeString('id-ID', {
          // msg.timestamp is already in Asia/Jakarta
          hour: '2-digit',
          minute: '2-digit',
          hour12: true
        });

        // Format WhatsApp message formatting into html 
        const content = convertWhatsAppToHTML(msg.content);

        return `
          <div class="message ${isBot ? 'sent' : 'received'}">
            <div class="message-content">
              <p>${content}</p>
              <span class="message-time">
                ${time}
                ${isBot ? '<span class="message-status"></span>' : ''}
              </span>
            </div>
          </div>
        `;
      }).join('');

      // Replace placeholders in the template
      return template
        .replace('{{recipientName}}', recipientName.charAt(0).toUpperCase())
        .replace('{{headerLineText}}', headerLineText)
        .replace('{{lastSeen}}', lastSeen)
        .replace('{{messages}}', messagesHTML)
        .replace('{{width}}', width || '400px');
    } catch (error) {
      console.error('Error generating chat HTML:', error);
      throw new ApiError(500, 'Failed to generate chat HTML');
    }
  }

  /**
   * No browser instance to close for wkhtmltopdf
   */
  async closeBrowser() {
    // No-op
    return Promise.resolve();
  }
}

const screenshotServiceInstance = new ScreenshotService();
module.exports = screenshotServiceInstance;
