# Technical Specification Document
## WhatsApp Chat Mockup API

### 1. Project Overview

**Application Name:** WhatsApp Chat Mockup API  
**Version:** 1.0  
**Purpose:** REST API that converts JSON message history to WhatsApp-style chat screenshot (base64 encoded)  
**Architecture:** RESTful API with server-side rendering

### 2. API Endpoint Specification

**2.1 Base Endpoint**
```
POST /api/whatsapp-screenshot
```

**2.2 Request Format**
```http
POST /api/whatsapp-screenshot
Content-Type: application/json

{
  "messages": [
    {
      "session_id": 6762016005514153,
      "timestamp": "2025-05-22T16:48:26.858",
      "sender": "Bot",
      "content": "Message content here...",
      "awb_number": "016005514153",
      "recipient_name": "Mila Palastri ( Tukang Jahit )",
      "recipient_phone": "+6285642856762"
    }
  ],
  "options": {
    "width": 400,
    "quality": "high",
    "format": "png"
  }
}
```

**2.3 Response Format**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "data": {
    "image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
    "metadata": {
      "width": 400,
      "height": 800,
      "format": "png",
      "message_count": 8,
      "generated_at": "2025-05-22T16:51:00.000Z"
    }
  }
}
```

### 3. Technical Stack

**3.1 Backend Framework**
- Node.js with Express.js
- Alternative: Python with Flask/FastAPI

**3.2 Image Generation**
- wkhtmltopdf (wkhtmltoimage) for screenshot generation
- Sharp library for image optimization
- Canvas API for direct rendering (alternative)

**3.3 Template Engine**
- HTML/CSS templates for WhatsApp UI
- Handlebars or EJS for dynamic content

### 4. System Requirements

**4.1 Server Requirements**
- Node.js 18+ or Python 3.8+
- Headless Chrome/Chromium
- 2GB RAM minimum (4GB recommended)
- Linux/Ubuntu server environment

**4.2 Dependencies**
```json
{
  "express": "^4.18.0",
  "wkhtmltoimage": "^0.1.5",
  "sharp": "^0.32.0",
  "joi": "^17.9.0",
  "helmet": "^7.0.0",
  "cors": "^2.8.5"
}
```

### 5. Implementation Architecture

**5.1 Project Structure**
```
whatsapp-api/
├── src/
│   ├── controllers/
│   │   └── screenshot.controller.js
│   ├── services/
│   │   ├── template.service.js
│   │   └── screenshot.service.js
│   ├── middleware/
│   │   └── validation.middleware.js
│   ├── templates/
│   │   ├── whatsapp-chat.html
│   │   └── styles.css
│   └── utils/
│       └── helpers.js
├── package.json
├── server.js
└── Dockerfile
```

**5.2 Core Service Implementation**

```javascript
// screenshot.service.js
const wkhtmltoimage = require('wkhtmltoimage');
const sharp = require('sharp');

class ScreenshotService {
  async generateWhatsAppScreenshot(messages, options = {}) {
     // Generate HTML content
    const htmlContent = this.generateChatHTML(messages);
    
    const screenshotOptions = {
        width: options.width || 400,
        format: options.format || 'png',
        quality: options.format === 'jpeg' ? 90 : undefined
    };

    return new Promise((resolve, reject) => {
        const stream = wkhtmltoimage.generate(htmlContent, screenshotOptions);

        const chunks = [];
        stream.on('data', (chunk) => chunks.push(chunk));
        stream.on('end', () => {
          const buffer = Buffer.concat(chunks);
          const base64Image = buffer.toString('base64');
          resolve(`data:image/${options.format || 'png'};base64,${base64Image}`);
        });
        stream.on('error', reject);
    });
  }
  
  generateChatHTML(messages) {
    // HTML template generation logic
    return `
      <!DOCTYPE html>
      <html>
      <head>
        <meta charset="utf-8">
        <style>${this.getCSSStyles()}</style>
      </head>
      <body>
        <div class="whatsapp-chat">
          ${this.renderHeader(messages[0])}
          <div class="chat-messages">
            ${messages.map(msg => this.renderMessage(msg)).join('')}
          </div>
        </div>
      </body>
      </html>
    `;
  }
}
```

### 6. Request Validation Schema

```javascript
const Joi = require('joi');

const messageSchema = Joi.object({
  session_id: Joi.number().required(),
  timestamp: Joi.string().isoDate().required(),
  sender: Joi.string().valid('Bot', 'Customer').required(),
  content: Joi.string().required(),
  awb_number: Joi.string().optional(),
  recipient_name: Joi.string().optional(),
  recipient_phone: Joi.string().optional()
});

const requestSchema = Joi.object({
  messages: Joi.array().items(messageSchema).min(1).required(),
  options: Joi.object({
    width: Joi.number().min(300).max(800).default(400),
    quality: Joi.string().valid('low', 'medium', 'high').default('high'),
    format: Joi.string().valid('png', 'jpeg').default('png')
  }).default({})
});
```

### 7. WhatsApp UI Template

**7.1 HTML Structure**
```html
<div class="whatsapp-chat">
  <div class="chat-header">
    <div class="contact-info">
      <h3>{{recipient_name}}</h3>
      <span>{{recipient_phone}}</span>
    </div>
  </div>
  
  <div class="chat-messages">
    {{#each messages}}
    <div class="message {{#if (eq sender 'Bot')}}outgoing{{else}}incoming{{/if}}">
      <div class="message-bubble">
        <div class="message-content">{{content}}</div>
        <div class="message-time">{{formatTime timestamp}}</div>
      </div>
    </div>
    {{/each}}
  </div>
</div>
```

**7.2 CSS Styling**
```css
.whatsapp-chat {
  width: 400px;
  background: #e5ddd5;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  padding: 0;
}

.chat-header {
  background: #075e54;
  color: white;
  padding: 15px;
  display: flex;
  align-items: center;
}

.message {
  margin: 8px 12px;
  display: flex;
}

.message.outgoing {
  justify-content: flex-end;
}

.message.incoming {
  justify-content: flex-start;
}

.message-bubble {
  max-width: 280px;
  padding: 8px 12px;
  border-radius: 8px;
  position: relative;
}

.message.outgoing .message-bubble {
  background: #dcf8c6;
}

.message.incoming .message-bubble {
  background: white;
}

.message-time {
  font-size: 11px;
  color: #667781;
  text-align: right;
  margin-top: 4px;
}
```

### 8. API Controller Implementation

```javascript
// screenshot.controller.js
const ScreenshotService = require('../services/screenshot.service');

class ScreenshotController {
  async generateScreenshot(req, res) {
    try {
      const { messages, options } = req.body;
      
      const screenshotService = new ScreenshotService();
      const base64Image = await screenshotService.generateWhatsAppScreenshot(messages, options);
      
      const response = {
        success: true,
        data: {
          image: base64Image,
          metadata: {
            width: options.width || 400,
            height: 'auto',
            format: options.format || 'png',
            message_count: messages.length,
            generated_at: new Date().toISOString()
          }
        }
      };
      
      res.json(response);
    } catch (error) {
      res.status(500).json({
        success: false,
        error: {
          message: 'Screenshot generation failed',
          details: error.message
        }
      });
    }
  }
}
```

### 9. Server Configuration

```javascript
// server.js
const express = require('express');
const helmet = require('helmet');
const cors = require('cors');
const screenshotRoutes = require('./src/routes/screenshot.routes');

const app = express();

// Security middleware
app.use(helmet());
app.use(cors());

// Body parsing
app.use(express.json({ limit: '10mb' }));

// Routes
app.use('/api', screenshotRoutes);

// Error handling
app.use((error, req, res, next) => {
  res.status(500).json({
    success: false,
    error: {
      message: 'Internal server error',
      details: process.env.NODE_ENV === 'development' ? error.message : undefined
    }
  });
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`WhatsApp Screenshot API running on port ${PORT}`);
});
```

### 10. Docker Configuration

```dockerfile
FROM node:18-alpine

# Install wkhtmltopdf and fonts
RUN apk add --no-cache \
    wkhtmltopdf \
    ttf-freefont \
    ttf-dejavu \
    ttf-droid \
    ttf-liberation \
    ttf-ubuntu-font-family

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE 3000
CMD ["node", "server.js"]
```

### 11. Error Handling

**11.1 Error Response Format**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": [
      {
        "field": "messages",
        "message": "At least one message is required"
      }
    ]
  }
}
```

**11.2 HTTP Status Codes**
- `200`: Success
- `400`: Bad Request (validation errors)
- `422`: Unprocessable Entity (invalid data format)
- `500`: Internal Server Error
- `503`: Service Unavailable (Puppeteer issues)

### 12. Performance & Scaling

**12.1 Optimization Strategies**
- Browser instance pooling for Puppeteer
- Image compression with Sharp
- Request rate limiting
- Caching for template rendering

**12.2 Resource Limits**
- Maximum message count: 100 per request
- Request timeout: 30 seconds
- Image size limit: 5MB
- Memory limit per request: 512MB

This specification provides a complete foundation for building a WhatsApp screenshot generation API that processes your JSON message format and returns base64-encoded images.