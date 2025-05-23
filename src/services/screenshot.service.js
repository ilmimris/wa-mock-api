const puppeteer = require('puppeteer');
const path = require('path');
const fs = require('fs/promises');
const { ApiError } = require('../middleware/error.middleware');
const { convertWhatsAppToHTML } = require('../utils/whatsapp-html');

class ScreenshotService {
  constructor() {
    this.templatePath = path.join(__dirname, '../templates/whatsapp-chat.html');
  }

  /**
   * Generate a WhatsApp-style chat screenshot from messages
   * @param {Array} messages - Array of message objects
   * @param {Object} options - Screenshot options
   * @returns {Promise<string>} Base64 encoded image
   */
  async generateWhatsAppScreenshot(messages, options = {}) {
    try {
      const { width = 400, format = 'png', quality = 'high' } = options;
      
      // Generate HTML content
      const htmlContent = await this.generateChatHTML(messages, { width });
      
      // Launch browser
      const browser = await puppeteer.launch({
        headless: 'new',
        args: [
          '--no-sandbox',
          '--disable-setuid-sandbox',
          '--disable-dev-shm-usage',
          '--disable-accelerated-2d-canvas',
          '--no-first-run',
          '--no-zygote',
          '--single-process',
          '--disable-gpu'
        ]
      });

      const page = await browser.newPage();
      
      // Set viewport
      await page.setViewport({
        width: parseInt(width, 10),
        height: 800,
        deviceScaleFactor: 2 // For better quality
      });

      // Set content and wait for rendering
      await page.setContent(htmlContent, { waitUntil: 'networkidle0' });
      
      // Calculate the height of the content
      const bodyHandle = await page.$('body');
      const { height } = await bodyHandle.boundingBox();
      await bodyHandle.dispose();
      
      // Set the viewport to the full height of the content
      await page.setViewport({
        width: parseInt(width, 10),
        height: Math.ceil(height),
        deviceScaleFactor: 2
      });

      // Take screenshot
      const screenshotOptions = {
        type: format,
        fullPage: true,
        omitBackground: true
      };

      // Add quality for formats that support it
      if (format === 'jpeg' || format === 'webp') {
        screenshotOptions.quality = quality === 'high' ? 90 : quality === 'medium' ? 70 : 50;
      }

      const screenshot = await page.screenshot(screenshotOptions);
      
      await browser.close();
      
      // Convert to base64
      const base64Image = screenshot.toString('base64');
      return `data:image/${format};base64,${base64Image}`;
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
      // Read the template file
      const templatePath = path.join(__dirname, '../templates/whatsapp-chat.html');
      let template = await fs.readFile(templatePath, 'utf-8');
      
      // Extract recipient info from the first message
      const firstMessage = messages[0] || {};
      const recipientName = firstMessage.recipient_name || 'Customer';
      let recipientPhone = firstMessage.recipient_phone || 'Unknown';

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
        .replace('{{recipientPhone}}', recipientPhone)
        .replace('{{lastSeen}}', lastSeen)
        .replace('{{messages}}', messagesHTML)
        .replace('{{width}}', options.width || '400px');
    } catch (error) {
      console.error('Error generating chat HTML:', error);
      throw new ApiError(500, 'Failed to generate chat HTML');
    }
  }
}

module.exports = new ScreenshotService();
