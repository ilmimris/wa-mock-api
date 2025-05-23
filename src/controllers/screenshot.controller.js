const screenshotService = require('../services/screenshot.service');
const { ApiError } = require('../middleware/error.middleware');

/**
 * Generate a WhatsApp chat screenshot
 * @route POST /api/whatsapp-screenshot
 * @param {Object} req - Express request object
 * @param {Object} res - Express response object
 * @param {Function} next - Next middleware function
 */
const generateScreenshot = async (req, res, next) => {
  try {
    const { messages, options = {} } = req.body;

    if (!messages || !Array.isArray(messages) || messages.length === 0) {
      throw new ApiError(400, 'At least one message is required');
    }

    // Generate the screenshot
    const imageData = await screenshotService.generateWhatsAppScreenshot(messages, options);
    
    // Get the first message for metadata
    const firstMessage = messages[0];
    const lastMessage = messages[messages.length - 1];

    // Prepare response
    const response = {
      success: true,
      data: {
        image: imageData,
        metadata: {
          width: options.width || 400,
          format: options.format || 'png',
          quality: options.quality || 'high',
          message_count: messages.length,
          first_message_timestamp: firstMessage.timestamp,
          last_message_timestamp: lastMessage.timestamp,
          generated_at: new Date().toISOString(),
          session_id: firstMessage.session_id,
          awb_number: firstMessage.awb_number
        }
      }
    };

    res.status(200).json(response);
  } catch (error) {
    next(error);
  }
};

module.exports = {
  generateScreenshot
};
