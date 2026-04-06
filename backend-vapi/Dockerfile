# ─────────────────────────────────────────────────────────────────
# Stage 1 — Build TypeScript
# ─────────────────────────────────────────────────────────────────
FROM node:20-slim AS builder

WORKDIR /build

COPY package*.json ./
RUN npm install

COPY tsconfig.json ./
COPY src ./src
RUN npm run build

# ─────────────────────────────────────────────────────────────────
# Stage 2 — Runtime image with Tailscale
# ─────────────────────────────────────────────────────────────────
FROM node:20-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
      curl \
      ca-certificates \
    && curl -fsSL https://tailscale.com/install.sh | sh \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Production deps only
COPY package*.json ./
RUN npm install --omit=dev

# Compiled JS from builder
COPY --from=builder /build/dist ./dist

# Entrypoint script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Persistent data directories
RUN mkdir -p /app/data/appointments

# /shared — other containers or the host mount here to read the tunnel URL
VOLUME ["/shared", "/var/lib/tailscale", "/app/data"]

EXPOSE 3000

ENTRYPOINT ["/docker-entrypoint.sh"]
