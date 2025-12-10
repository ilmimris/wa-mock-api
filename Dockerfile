FROM node:18-slim

# Install wkhtmltopdf, xvfb (required for unpatched wkhtmltopdf), and fonts
RUN apt-get update && apt-get install -y \
    wkhtmltopdf \
    xvfb \
    fonts-freefont-ttf \
    fonts-dejavu-core \
    fonts-liberation \
    fonts-ubuntu \
    fonts-ipafont-gothic \
    fonts-wqy-zenhei \
    fonts-thai-tlwg \
    fonts-kacst \
    fonts-noto-color-emoji \
    fonts-symbola \
    && rm -rf /var/lib/apt/lists/*

# Create a shim for wkhtmltoimage that runs it with xvfb-run
# This is necessary because the npm wrapper calls 'wkhtmltoimage' directly
RUN echo '#!/bin/bash\nxvfb-run -a /usr/bin/wkhtmltoimage "$@"' > /usr/local/bin/wkhtmltoimage-shim && \
    chmod +x /usr/local/bin/wkhtmltoimage-shim

ENV NODE_ENV=production

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE 3000
CMD ["node", "server.js"]
