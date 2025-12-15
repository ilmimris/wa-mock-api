FROM node:18-slim

# Install wkhtmltopdf, xvfb (required for unpatched wkhtmltopdf), and fonts
# We also install dumb-init to handle signals and zombies properly
RUN apt-get update && apt-get install -y \
    wkhtmltopdf \
    xvfb \
    dumb-init \
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

# Setup Xvfb as a background process script
RUN echo '#!/bin/bash\nXvfb :99 -screen 0 1024x768x24 > /dev/null 2>&1 &\nexec "$@"' > /entrypoint.sh && \
    chmod +x /entrypoint.sh

ENV NODE_ENV=production
ENV DISPLAY=:99

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE 3000
ENTRYPOINT ["/usr/bin/dumb-init", "--", "/entrypoint.sh"]
CMD ["node", "server.js"]
