# ─── Build stage ──────────────────────────────────────────────────────────────
FROM node:18-alpine AS builder
WORKDIR /app

# Install deps
COPY package*.json ./
RUN npm ci

# Copy source & build
COPY . .
RUN npm run build

# ─── Production stage ─────────────────────────────────────────────────────────
FROM nginx:stable-alpine
# Replace default nginx config to serve index.html for SPA & proxy /api
COPY nginx.conf /etc/nginx/conf.d/default.conf
# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
