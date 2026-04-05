#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# SmileDental — Start webhook server with Tailscale Funnel
#
# Usage:
#   bash start.sh
#
# What it does:
#   1. Authenticates Tailscale with the key in .env
#   2. Enables Tailscale Funnel on port 3000 (public HTTPS tunnel)
#   3. Reads the auto-assigned public URL
#   4. Updates WEBHOOK_URL in .env
#   5. Starts the Node.js webhook server
# ─────────────────────────────────────────────────────────────────────────────
set -e

# Load .env
if [ ! -f .env ]; then
  echo "❌ .env file not found. Copy .env.example and fill in your keys."
  exit 1
fi
set -a; source .env; set +a

# ── 1. Authenticate Tailscale ─────────────────────────────────────────────
echo ""
echo "🔑 Authenticating Tailscale..."
tailscale up \
  --authkey="$TAILSCALE_AUTH_KEY" \
  --hostname="${TAILSCALE_HOSTNAME:-smiledental-webhook}" \
  --accept-routes

echo "✅ Tailscale connected"

# ── 2. Enable Funnel on port 3000 ────────────────────────────────────────
echo ""
echo "🌐 Enabling Tailscale Funnel → port 3000..."
tailscale funnel --bg 3000

# Brief wait for funnel to register
sleep 2

# ── 3. Get public URL ─────────────────────────────────────────────────────
echo ""
echo "🔍 Reading tunnel URL..."

# Try 'tailscale funnel status' first (newer CLI)
FUNNEL_URL=$(tailscale funnel status 2>/dev/null \
  | grep -oE 'https://[^[:space:]]+' \
  | head -1)

# Fallback: build from machine DNS name
if [ -z "$FUNNEL_URL" ]; then
  DNS_NAME=$(tailscale status --json 2>/dev/null \
    | grep -o '"DNSName":"[^"]*"' \
    | head -1 \
    | cut -d'"' -f4 \
    | sed 's/\.$//')
  if [ -n "$DNS_NAME" ]; then
    FUNNEL_URL="https://${DNS_NAME}"
  fi
fi

if [ -z "$FUNNEL_URL" ]; then
  echo "⚠️  Could not auto-detect funnel URL."
  echo "   Run:  tailscale funnel status"
  echo "   Then copy the https:// URL and set it below."
else
  WEBHOOK_FULL="${FUNNEL_URL}/webhook"
  echo ""
  echo "┌─────────────────────────────────────────────────────┐"
  echo "│  Tailscale Funnel is live!                          │"
  echo "│                                                     │"
  echo "│  Public URL : ${FUNNEL_URL}"
  echo "│  Webhook    : ${WEBHOOK_FULL}"
  echo "│                                                     │"
  echo "│  Paste the Webhook URL into:                        │"
  echo "│  Vapi dashboard → Phone Number → Server URL         │"
  echo "└─────────────────────────────────────────────────────┘"
  echo ""

  # ── 4. Update .env ──────────────────────────────────────────────────────
  if grep -q '^WEBHOOK_URL=' .env; then
    # Replace existing line (portable sed for both GNU and BSD)
    sed -i.bak "s|^WEBHOOK_URL=.*|WEBHOOK_URL=${FUNNEL_URL}|" .env && rm -f .env.bak
  else
    echo "WEBHOOK_URL=${FUNNEL_URL}" >> .env
  fi
  echo "✅ .env updated: WEBHOOK_URL=${FUNNEL_URL}"
fi

# ── 5. Start the webhook server ───────────────────────────────────────────
echo ""
echo "🦷 Starting SmileDental webhook server on port ${PORT:-3000}..."
echo ""
npm run dev
