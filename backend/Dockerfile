# ─────────────────────────────────────────────────────────────
# Stage 1 — Build Rust binary
# ─────────────────────────────────────────────────────────────
FROM rust:1.75-slim AS builder

WORKDIR /build

# Cache dependencies (rebuild only on Cargo.toml change)
COPY Cargo.toml Cargo.lock* ./
COPY .cargo ./.cargo
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release 2>/dev/null || true

# Copy actual source and build
COPY . .
RUN rm -f src/main.rs
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
