const express = require('express');
const router = express.Router();
const { validateScreenshotRequest } = require('../middleware/validation.middleware');
const { generateScreenshot } = require('../controllers/screenshot.controller');

/**
 * @swagger
 * /api/whatsapp-screenshot:
 *   post:
 *     summary: Generate a WhatsApp-style chat screenshot
 *     description: Converts a JSON message history into a WhatsApp-style chat screenshot
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             required:
 *               - messages
 *             properties:
 *               messages:
 *                 type: array
 *                 items:
 *                   type: object
 *                   required:
 *                     - session_id
 *                     - timestamp
 *                     - sender
 *                     - content
 *                   properties:
 *                     session_id:
 *                       type: number
 *                       example: 6762016005514153
 *                     timestamp:
 *                       type: string
 *                       format: date-time
 *                       example: "2025-05-22T16:48:26.858Z"
 *                     sender:
 *                       type: string
 *                       enum: [Bot, Customer]
 *                       example: "Bot"
 *                     content:
 *                       type: string
 *                       example: "Hello, how can I help you today?"
 *                     awb_number:
 *                       type: string
 *                       example: "016005514153"
 *                     recipient_name:
 *                       type: string
 *                       example: "John Doe"
 *                     recipient_phone:
 *                       type: string
 *                       example: "+6281234567890"
 *               options:
 *                 type: object
 *                 properties:
 *                   width:
 *                     type: number
 *                     minimum: 300
 *                     maximum: 1200
 *                     default: 400
 *                     example: 400
 *                   quality:
 *                     type: string
 *                     enum: [low, medium, high]
 *                     default: high
 *                     example: "high"
 *                   format:
 *                     type: string
 *                     enum: [png, jpeg, webp]
 *                     default: png
 *                     example: "png"
 *     responses:
 *       200:
 *         description: Successful operation
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 data:
 *                   type: object
 *                   properties:
 *                     image:
 *                       type: string
 *                       format: byte
 *                       description: Base64 encoded image with data URL
 *                       example: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."
 *                     metadata:
 *                       type: object
 *                       properties:
 *                         width:
 *                           type: number
 *                           example: 400
 *                         height:
 *                           type: number
 *                           example: 800
 *                         format:
 *                           type: string
 *                           example: "png"
 *                         message_count:
 *                           type: number
 *                           example: 5
 *                         generated_at:
 *                           type: string
 *                           format: date-time
 *                           example: "2025-05-22T16:51:00.000Z"
 *       400:
 *         description: Invalid input
 *       500:
 *         description: Server error
 */
router.post('/whatsapp-screenshot', validateScreenshotRequest, generateScreenshot);

// Health check endpoint
router.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok', timestamp: new Date().toISOString() });
});

module.exports = router;
