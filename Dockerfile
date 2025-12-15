# Use the official Node.js 18 image as the base image
FROM node:18-slim

# Set the working directory in the container
WORKDIR /usr/src/app

# Install required system dependencies for Puppeteer
RUN apt-get update && apt-get install -y \
    chromium \
    fonts-ipafont-gothic \
    fonts-wqy-zenhei \
    fonts-thai-tlwg \
    fonts-kacst \
    fonts-symbola \
    fonts-noto \
    fonts-freefont-ttf \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Set environment variables for Puppeteer
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium \
    PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true

# Copy package.json and package-lock.json
COPY package*.json ./


# Install application dependencies
RUN npm install --production

# Copy the rest of the application code
COPY . .

# Expose the port the app runs on
EXPOSE 3000

# Set environment variables
ENV NODE_ENV=production

# Set the command to run the application
CMD ["node", "server.js"]
