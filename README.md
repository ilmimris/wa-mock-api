# Go WhatsApp Chat Screenshot API

A RESTful API, written in Go, that converts JSON message history into WhatsApp-style chat screenshots. This API generates high-quality, realistic-looking WhatsApp chat screenshots from provided message data and returns them as raw image bytes.

## Features

- Convert JSON message history to WhatsApp-style chat screenshots.
- Customizable output options (width, height, quality, format, selector, full page capture).
- Realistic WhatsApp UI rendered from an HTML template.
- Support for different message senders to distinguish between outgoing ("sent") and incoming ("received") messages.
- Dockerized for easy deployment.

## Prerequisites

- **Go**: Version 1.21 or higher.
- **Chrome/Chromium**: Required by the `chromedp` library for headless browser operations. If not using Docker, ensure Chrome or Chromium is installed and accessible in your PATH.
- **Docker**: (Optional) For building and running the application in a containerized environment.

## Installation & Setup

1.  **Clone the repository**:
    ```bash
    git clone <repository-url>
    cd go-whatsapp-screenshot
    ```

2.  **Download Go module dependencies** (if developing locally):
    Go modules are typically downloaded automatically during the build process. You can explicitly download them using:
    ```bash
    go mod download
    ```

## Building the Application

### Local Build

To build the application locally, run the following command from the project root:
```bash
go build -o server ./cmd/server/main.go
```
This will create an executable file named `server` (or `server.exe` on Windows) in the project root.

### Docker Build

A `Dockerfile` is provided to build a container image for the application.
```bash
docker build -t go-whatsapp-screenshot .
```

## Running the Application

### Locally (Compiled Binary)

After building the application as described above, you can run it directly:
```bash
./server
```
The server will start, by default, on port `8080`.

**Configuration**:
-   `PORT`: The port for the server to listen on can be configured via the `PORT` environment variable. For example:
    ```bash
    PORT=8888 ./server
    ```

### Using Docker

Run the application using the Docker image built previously:
```bash
docker run -p 8080:8080 --rm go-whatsapp-screenshot
```
To run on a different host port or set the application's internal port:
```bash
docker run -p <host_port>:8080 -e PORT=8080 --rm go-whatsapp-screenshot
```
For example, to map host port 9000 to the container's port 8080:
```bash
docker run -p 9000:8080 -e PORT=8080 --rm go-whatsapp-screenshot
```
The `--rm` flag automatically removes the container when it exits.

## API Endpoint

### Generate WhatsApp Screenshot

**Endpoint:** `POST /screenshot`

**Request Body:**

The request body should be a JSON object with the following structure:

```json
{
  "messages": [
    {
      "timestamp": "2025-05-22T16:48:26.858Z",
      "sender": "Bot", // Or any identifier for one party
      "content": "Hello, how can I help you today?"
    },
    {
      "timestamp": "2025-05-22T16:49:15.123Z",
      "sender": "Customer", // Or any identifier for the other party
      "content": "Hi, I have a question about my order."
    }
    // ... more messages
  ],
  "chatName": "Support Chat", // Displayed in the chat header
  "lastSeen": "online",      // Displayed as the last seen status
  "outputFileName": "my-chat.png", // Optional: Suggested filename for download
  "screenshotOptions": { // Optional: Override default screenshot settings
    "width": 400,            // Width of the browser viewport for rendering
    "height": 700,           // Height of the browser viewport
    "selector": ".chat-container", // CSS selector of the element to capture (defaults to .chat-container)
    "isFullPage": false,     // If true, captures the full scrollable page
    "format": "png",         // "png" or "jpeg"
    "quality": 90,           // JPEG quality (1-100), only if format is "jpeg" and isFullPage is true
    "timeout": 30            // Timeout in seconds for the screenshot operation (Not yet fully implemented in options, uses server default)
  }
}
```
See `sample-messages.json` for a more detailed example of the `messages` array structure. The `sender` field is used to determine if a message is "sent" (e.g., if `sender` is "Bot" or "User", it's styled as an outgoing message without an author name in the bubble) or "received" (styled as an incoming message, potentially showing the sender's name).

**Success Response:**

-   **Status Code:** `200 OK`
-   **Content-Type:** `image/png` or `image/jpeg` (depending on `screenshotOptions.format`)
-   **Body:** Raw image bytes.
-   **Content-Disposition:** Header is set to suggest a filename (e.g., `inline; filename="whatsapp-chat-screenshot.png"` or `attachment; filename="your-name.png"` if `outputFileName` was provided).

**Error Response:**

-   **Status Code:**
    -   `400 Bad Request`: For invalid JSON or missing required fields.
    -   `405 Method Not Allowed`: If a non-POST request is made.
    -   `500 Internal Server Error`: If any error occurs during HTML generation or screenshot capture.
-   **Content-Type:** `application/json`
-   **Body:**
    ```json
    {
      "error": "Error message describing the issue",
      "details": "Optional additional details about the error"
    }
    ```

## Dependencies

-   **Go**: The core programming language.
-   **`chromedp`**: Go library for driving browsers using the Chrome DevTools Protocol. Used for taking screenshots.
-   **Chrome/Chromium**: A running instance of Chrome or Chromium is required for `chromedp` to connect to. The Dockerfile handles this by installing Chromium. If running locally without Docker, you must have it installed.

## Project Structure

```
go-whatsapp-screenshot/
├── cmd/server/             # Main application entry point (main.go)
├── internal/               # Internal application logic
│   ├── handlers/           # HTTP request handlers (screenshot_handler.go)
│   ├── middleware/         # HTTP middleware (error_middleware.go)
│   ├── services/           # Business logic (screenshot_service.go)
│   └── utils/              # Utility functions (html_generator.go)
├── templates/              # HTML templates (whatsapp-chat.html)
├── Dockerfile              # For building the Docker container
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── README.md               # This file
└── ...                     # Other configuration and sample files
```

## Development

### Running Tests

To run the unit tests for the Go application:
```bash
go test ./...
```
This command will run tests in all subdirectories.

## License

MIT
```
