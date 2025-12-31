# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@9 --activate

# Copy package files first for better caching
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy source code
COPY . .

# Build the app
RUN pnpm build

# Runtime stage - serve static files with nginx
FROM nginx:alpine

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config template (will be processed at runtime)
COPY nginx.conf /etc/nginx/templates/default.conf.template

# Default API URL for local development
ENV API_URL=http://localhost:8080

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1

# nginx:alpine uses envsubst automatically for .template files in /etc/nginx/templates/
CMD ["nginx", "-g", "daemon off;"]
