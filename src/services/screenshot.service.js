const puppeteer = require('puppeteer');
const path = require('path');
const fs = require('fs/promises');
const { ApiError } = require('../middleware/error.middleware');
const { convertWhatsAppToHTML } = require('../utils/whatsapp-html');

class ScreenshotService {
  constructor() {
    this.templatePath = path.join(__dirname, '../templates/whatsapp-chat.html');
    this.browser = null;
    this.chatTemplate = null; // Initialize chatTemplate property
    this.initializeBrowser().catch(err => {
      console.error("Failed to initialize ScreenshotService on startup:", err);
      // Depending on the application's needs, this might be a fatal error.
      // For now, we log it. The service might be in a non-operational state.
    });
  }

  async initializeBrowser() {
    // Load HTML template if not already loaded
    if (!this.chatTemplate) {
      try {
        console.log('Loading HTML template...');
        this.chatTemplate = await fs.readFile(this.templatePath, 'utf-8');
        console.log('HTML template loaded successfully.');
      } catch (error) {
        console.error('Failed to load HTML template:', error);
        // This is a critical error for the service's operation.
        // Rethrow to be caught by the constructor's catch or calling context.
        throw error; 
      }
    }

    if (this.browser && this.browser.isConnected()) {
      console.log('Browser already initialized.');
      return;
    }
    console.log('Initializing browser...');
    try {
      this.browser = await puppeteer.launch({
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
      console.log('Browser initialized successfully.');
    } catch (error) {
      console.error('Error initializing browser:', error);
      // We'll let subsequent calls to generateWhatsAppScreenshot handle the error
      // by attempting to re-initialize. If it fails there, it will throw.
      this.browser = null; // Ensure browser is null if initialization failed
      throw error; // Rethrow to allow handling by the caller if needed immediately
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
      const { width = 400, format = 'png', quality = 'high' } = options;
      
      // Generate HTML content
      const htmlContent = await this.generateChatHTML(messages, { width });

      // Ensure browser is initialized
      if (!this.browser || !this.browser.isConnected()) {
        await this.initializeBrowser();
      }
      
      const page = await this.browser.newPage();
      
      // Set content first. For local content, 'domcontentloaded' is usually sufficient.
      // A minimal default viewport is active before this, which is fine for rendering.
      await page.setContent(htmlContent, { waitUntil: 'domcontentloaded' });
      
      // Calculate the height of the content
      const bodyHandle = await page.$('body');
      if (!bodyHandle) {
        await page.close(); // Close the page to free up resources
        throw new ApiError(500, 'Failed to get body handle for height calculation');
      }
      const boundingBox = await bodyHandle.boundingBox();
      await bodyHandle.dispose();

      if (!boundingBox) {
        await page.close(); // Close the page to free up resources
        throw new ApiError(500, 'Failed to get bounding box for height calculation');
      }
      const contentHeight = Math.ceil(boundingBox.height);
      
      // Set the viewport to the full height of the content and desired width
      await page.setViewport({
        width: parseInt(width, 10),
        height: contentHeight > 0 ? contentHeight : 800, // Fallback height if calculation is zero
        deviceScaleFactor: 2 // For better quality
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
      
      // Do not close the browser here; it's reused.
      // await browser.close(); 
      
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

  /**
   * Closes the Puppeteer browser instance.
   * This should be called on application shutdown.
   */
  async closeBrowser() {
    if (this.browser && this.browser.isConnected()) {
      console.log('Closing browser...');
      await this.browser.close();
      this.browser = null;
      console.log('Browser closed.');
    } else {
      console.log('Browser not open or already closed.');
    }
  }
}

const screenshotServiceInstance = new ScreenshotService();

// To ensure the browser is closed gracefully on application shutdown,
// you would typically call screenshotServiceInstance.closeBrowser() in your main server file (e.g., server.js or app.js)
// Example for server.js:
//
// const screenshotService = require('./services/screenshot.service'); // Adjust path as needed
//
// process.on('SIGINT', async () => {
//   console.log('SIGINT signal received. Closing browser...');
//   await screenshotService.closeBrowser();
//   process.exit(0);
// });
//
// process.on('SIGTERM', async () => {
//   console.log('SIGTERM signal received. Closing browser...');
//   await screenshotService.closeBrowser();
//   process.exit(0);
// });
//
// // Handle unhandled rejections and uncaught exceptions to also close browser
// process.on('unhandledRejection', async (reason, promise) => {
//   console.error('Unhandled Rejection at:', promise, 'reason:', reason);
//   await screenshotService.closeBrowser();
//   process.exit(1);
// });
//
// process.on('uncaughtException', async (error) => {
//   console.error('Uncaught Exception:', error);
//   await screenshotService.closeBrowser();
//   process.exit(1);
// });

module.exports = screenshotServiceInstance;
