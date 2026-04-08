# ─────────────────────────────────────────────────────────────
# Stage 1 — Build Rust binary
# ─────────────────────────────────────────────────────────────
FROM rust:1-slim AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Build the real project sources (no dummy main cache trick)
COPY Cargo.toml Cargo.lock* ./
COPY .cargo ./.cargo
COPY src ./src
RUN cargo build --release

# ─────────────────────────────────────────────────────────────
# Stage 2 — Minimal runtime image
# ─────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates libssl3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/target/release/dental-ai-browser-backend /usr/local/bin/dental-ai-browser-backend

RUN mkdir -p /app/data

EXPOSE 3001

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD curl -f http://localhost:3001/health || exit 1

CMD ["dental-ai-browser-backend"]
