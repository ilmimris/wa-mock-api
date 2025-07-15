const path = require('path');
const fs = require('fs/promises');
const { toPng } = require('html-to-image');
const { ApiError } = require('../middleware/error.middleware');
const { convertWhatsAppToHTML } = require('../utils/whatsapp-html');

class ScreenshotService {
  constructor() {
    this.templatePath = path.join(__dirname, '../templates/whatsapp-chat.html');
    this.chatTemplate = null; // Initialize chatTemplate property
    this.loadTemplate().catch(err => {
      console.error("Failed to load HTML template on startup:", err);
    });
  }

  async loadTemplate() {
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
      const { width = 400, format = 'png', headerDisplay = 'phone' } = options;

      // Generate HTML content
      const htmlContent = await this.generateChatHTML(messages, { width, headerDisplay });

      // Create a dummy element to render the HTML
      const dataUrl = await toPng(htmlContent, { width, height: 10, skipAutoScale: true });

      return dataUrl;

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
        // This case should ideally not be reached if initializeBrowser was successful.
        // However, as a fallback, or if generateChatHTML could be called before full initialization.
        console.error('Chat template not loaded. Attempting to load now...');
        try {
          this.chatTemplate = await fs.readFile(this.templatePath, 'utf-8');
          console.log('HTML template loaded on demand.');
        } catch (error) {
          console.error('Failed to load HTML template on demand:', error);
          throw new ApiError(500, 'Failed to load chat template');
        }
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
          recipientPhone = `+62 ${recipientPhone}`;
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
        hour: '2-digit',
        minute: '2-digit',
        hour12: true
      });

      // Generate messages HTML
      const messagesHTML = messages.map(msg => {
        const isBot = msg.sender === 'Bot';
        const time = new Date(msg.timestamp).toLocaleTimeString('id-ID', {
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

}

module.exports = new ScreenshotService();
