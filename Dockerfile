# Stage 1: Build the Go application
FROM golang:1.23-bookworm AS builder

# Install system dependencies for chromedp (Chromium)
# Based on https://github.com/chromedp/chromedp/issues/1003#issuecomment-1071831884
# and general recommendations for running headless Chrome in Docker.
RUN apt-get update && apt-get install -y \
    chromium \
    # Below are commonly needed for headless execution and font rendering
    xvfb \
    fonts-ipafont-gothic \
    fonts-wqy-zenhei \
    fonts-thai-tlwg \
    fonts-kacst \
    fonts-freefont-ttf \
    libfontconfig1 \
    libxrender1 \
    libxss1 \
    libxtst6 \
    libnss3 \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libatspi2.0-0 \
    libcairo2 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libgbm1 \
    libgdk-pixbuf-2.0-0 \
    libglib2.0-0 \
    libgtk-3-0 \
    libpango-1.0-0 \
    libpangocairo-1.0-0 \
    libx11-6 \
    libx11-xcb1 \
    libxcb1 \
    libxcomposite1 \
    libxcursor1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxi6 \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the templates directory
COPY templates ./templates/

# Copy the rest of the application source code
COPY . .

# Build the Go application
# CGO_ENABLED=0 can sometimes help create more portable binaries,
# but for chromedp, CGO might be needed or it might link dynamically.
# Defaulting to CGO_ENABLED=1 (or unset) as chromedp often needs it.
RUN go build -o /app/server ./cmd/server/main.go

# Stage 2: Create a smaller final image (optional, but good practice)
# For this exercise, we'll stick to a single-stage build for simplicity,
# as managing Chromium dependencies across stages can be complex.
# The 'builder' stage will effectively be our final image.

# Set environment variables (optional, as the app defaults PORT if not set)
ENV PORT=8080

# Expose the port the server listens on
EXPOSE 8080

# Set the entrypoint to run the application
ENTRYPOINT ["/app/server"]

# Healthcheck (optional, but good practice)
# To use this, ensure 'curl' is installed (RUN apt-get install -y curl)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
#   CMD curl -f http://localhost:8080/health || exit 1
```
