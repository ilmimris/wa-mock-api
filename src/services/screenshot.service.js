const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs/promises');
const os = require('os');
const { ApiError } = require('../middleware/error.middleware');
const { convertWhatsAppToHTML } = require('../utils/whatsapp-html');

class ScreenshotService {
  constructor() {
    this.templatePath = path.join(__dirname, '../templates/whatsapp-chat.html');
    this.chatTemplate = null;
    this.initializeTemplate().catch(err => {
      console.error("Failed to initialize ScreenshotService on startup:", err);
    });
  }

  async initializeTemplate() {
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
   * Generate a WhatsApp-style chat screenshot from messages using wkhtmltoimage
   * @param {Array} messages - Array of message objects
   * @param {Object} options - Screenshot options
   * @returns {Promise<string>} Base64 encoded image
   */
  async generateWhatsAppScreenshot(messages, options = {}) {
    let inputPath = null;
    let outputPath = null;

    try {
      const { width = 400, format = 'png', quality = 'high' } = options;

      // Generate HTML content
      const htmlContent = await this.generateChatHTML(messages, options);

      // Create temp files for input HTML and output image
      const tempDir = os.tmpdir();
      const timestamp = Date.now();
      inputPath = path.join(tempDir, `wa-chat-${timestamp}.html`);
      outputPath = path.join(tempDir, `wa-chat-${timestamp}.${format}`);

      // Write HTML to temp file
      await fs.writeFile(inputPath, htmlContent, 'utf-8');

      // Build wkhtmltoimage arguments
      const args = [
        '--width', String(width),
        '--format', format,
        '--enable-local-file-access',
        '--disable-smart-width',
        '--quiet'
      ];

      // Add quality for formats that support it
      if (format === 'jpeg' || format === 'jpg') {
        const qualityValue = quality === 'high' ? 90 : quality === 'medium' ? 70 : 50;
        args.push('--quality', String(qualityValue));
      }

      // Add input and output paths
      args.push(inputPath, outputPath);

      // Execute wkhtmltoimage
      await this.runWkhtmltoimage(args);

      // Read the output image
      const imageBuffer = await fs.readFile(outputPath);

      // Convert to base64
      const base64Image = imageBuffer.toString('base64');
      const mimeType = format === 'jpg' ? 'jpeg' : format;
      return `data:image/${mimeType};base64,${base64Image}`;

    } catch (error) {
      console.error('Error generating screenshot:', error);
      throw new ApiError(500, `Failed to generate screenshot: ${error.message}`);
    } finally {
      // Clean up temp files
      try {
        if (inputPath) await fs.unlink(inputPath).catch(() => {});
        if (outputPath) await fs.unlink(outputPath).catch(() => {});
      } catch (cleanupError) {
        console.warn('Failed to clean up temp files:', cleanupError);
      }
    }
  }

  /**
   * Execute wkhtmltoimage command
   * @private
   */
  runWkhtmltoimage(args) {
    return new Promise((resolve, reject) => {
      const wkhtmltoimage = process.env.WKHTMLTOIMAGE_PATH || 'wkhtmltoimage';
      console.log(`Running: ${wkhtmltoimage} ${args.join(' ')}`);

      const child = spawn(wkhtmltoimage, args);

      let stderr = '';

      child.stderr.on('data', (data) => {
        stderr += data.toString();
      });

      child.on('close', (code) => {
        if (code === 0) {
          resolve();
        } else {
          reject(new Error(`wkhtmltoimage exited with code ${code}: ${stderr}`));
        }
      });

      child.on('error', (error) => {
        reject(new Error(`Failed to spawn wkhtmltoimage: ${error.message}`));
      });
    });
  }

  /**
   * Generate HTML content for the chat
   * @private
   */
  async generateChatHTML(messages, options = {}) {
    try {
      const { width, headerDisplay } = options;

      if (!this.chatTemplate) {
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
   * No-op for backwards compatibility.
   * wkhtmltoimage doesn't need cleanup like Puppeteer's browser.
   */
  async closeBrowser() {
    console.log('closeBrowser() called - no cleanup needed for wkhtmltoimage.');
  }
}

const screenshotServiceInstance = new ScreenshotService();

module.exports = screenshotServiceInstance;
