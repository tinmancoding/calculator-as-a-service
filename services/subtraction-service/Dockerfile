FROM node:20-slim

WORKDIR /app

# Install dependencies
COPY package.json package-lock.json* ./
RUN npm install

# Copy source code and build
COPY tsconfig.json ./
COPY src ./src
RUN npm run build

# Expose the service port
EXPOSE 8083

# Run the compiled JavaScript
CMD ["node", "dist/index.js"]
