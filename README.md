# WhatsApp Chat Mockup API

A RESTful API that converts JSON message history into WhatsApp-style chat screenshots. This API generates high-quality, realistic-looking WhatsApp chat screenshots from provided message data and returns them as base64-encoded images.

## Features

- Convert JSON message history to WhatsApp-style chat screenshots
- Customizable output options (width, quality, format)
- Realistic WhatsApp UI with proper message bubbles and timestamps
- Support for both Bot and Customer messages
- Responsive design that works on different screen sizes

## Prerequisites

- Node.js 18 or higher
- npm or yarn
- Puppeteer (Chromium browser)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/whatsapp-chat-mockup-api.git
   cd whatsapp-chat-mockup-api
   ```

2. Install dependencies:
   ```bash
   npm install
   # or
   yarn install
   ```

3. Create a `.env` file in the root directory and configure the environment variables (see `.env.example` for reference).

## Usage

### Starting the Server

```bash
# Development mode with hot-reload
npm run dev

# Production mode
npm start
```

The API will be available at `http://localhost:3000` by default.

### API Endpoint

#### Generate WhatsApp Screenshot

**Endpoint:** `POST /api/whatsapp-screenshot`

**Request Body:**

```json
{
  "messages": [
    {
      "session_id": 6762016005514153,
      "timestamp": "2025-05-22T16:48:26.858Z",
      "sender": "Bot",
      "content": "Hello, how can I help you today?",
      "awb_number": "016005514153",
      "recipient_name": "John Doe",
      "recipient_phone": "+6281234567890"
    },
    {
      "session_id": 6762016005514153,
      "timestamp": "2025-05-22T16:49:15.123Z",
      "sender": "Customer",
      "content": "Hi, I have a question about my order"
    }
  ],
  "options": {
    "width": 400,
    "quality": "high",
    "format": "png"
  }
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
    "metadata": {
      "width": 400,
      "height": 800,
      "format": "png",
      "message_count": 2,
      "first_message_timestamp": "2025-05-22T16:48:26.858Z",
      "last_message_timestamp": "2025-05-22T16:49:15.123Z",
      "generated_at": "2025-05-22T16:51:00.000Z",
      "session_id": 6762016005514153,
      "awb_number": "016005514153"
    }
  }
}
```

### Request Parameters

#### Messages

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| session_id | number | Yes | Unique identifier for the chat session |
| timestamp | string | Yes | ISO 8601 timestamp of the message |
| sender | string | Yes | Either "Bot" or "Customer" |
| content | string | Yes | The message text content |
| awb_number | string | No | Air Waybill number (optional) |
| recipient_name | string | No | Name of the recipient (optional) |
| recipient_phone | string | No | Phone number of the recipient (optional) |

#### Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| width | number | 400 | Width of the output image (300-1200px) |
| quality | string | "high" | Image quality ("low", "medium", or "high") |
| format | string | "png" | Output format ("png", "jpeg", or "webp") |

## Development

### Project Structure

```
whatsapp-chat-mockup-api/
├── src/
│   ├── controllers/         # Request handlers
│   ├── middleware/          # Express middleware
│   ├── routes/              # API routes
│   ├── services/            # Business logic
│   └── templates/           # HTML/CSS templates
├── .env                     # Environment variables
├── .gitignore
├── package.json
├── README.md
└── server.js                # Application entry point
```

### Running Tests

```bash
npm test
```

### Linting

```bash
npm run lint
```

## Deployment

### Docker

1. Build the Docker image:
   ```bash
   docker build -t whatsapp-chat-mockup .
   ```

2. Run the container:
   ```bash
   docker run -p 3000:3000 -d whatsapp-chat-mockup
   ```

### PM2 (Production)

```bash
# Install PM2 globally
npm install -g pm2

# Start the application
pm2 start server.js --name "whatsapp-chat-mockup"

# Save the process list
pm2 save

# Generate startup script
pm2 startup
```

## License

MIT
