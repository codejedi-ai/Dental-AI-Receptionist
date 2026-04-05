#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────
# Docker entrypoint — boots Tailscale, opens Funnel, starts server
# ─────────────────────────────────────────────────────────────────
set -e

SHARED_DIR="/shared"
URL_FILE="$SHARED_DIR/webhook-url.txt"
FULL_URL_FILE="$SHARED_DIR/webhook-full-url.txt"
STATUS_FILE="$SHARED_DIR/status.json"

mkdir -p "$SHARED_DIR"

# ── 1. Start Tailscale daemon ─────────────────────────────────────
echo "[entrypoint] Starting tailscaled..."
mkdir -p /var/run/tailscale /var/lib/tailscale
tailscaled \
  --state=/var/lib/tailscale/tailscaled.state \
  --socket=/var/run/tailscale/tailscaled.sock \
  &
sleep 3

# ── 2. Authenticate ───────────────────────────────────────────────
echo "[entrypoint] Authenticating Tailscale..."
tailscale up \
  --authkey="${TAILSCALE_AUTH_KEY}" \
  --hostname="${TAILSCALE_HOSTNAME:-smiledental-webhook}" \
  --accept-routes \
  --ssh=false
echo "[entrypoint] Tailscale connected"

# ── 3. Enable Funnel (non-fatal if not yet enabled on tailnet) ────
echo "[entrypoint] Enabling Tailscale Funnel on port 3000..."
FUNNEL_URL=""

if tailscale funnel --bg 3000 2>/tmp/funnel-err; then
  sleep 3

  # Try funnel status for the URL
  FUNNEL_URL=$(tailscale funnel status 2>/dev/null \
    | grep -oE 'https://[^[:space:]]+' \
    | head -1 || true)

  # Fallback: build from DNS name
  if [ -z "$FUNNEL_URL" ]; then
    DNS_NAME=$(tailscale status --json 2>/dev/null \
      | grep -o '"DNSName":"[^"]*"' \
      | head -1 \
      | sed 's/"DNSName":"//;s/"//g;s/\.$//' || true)
    [ -n "$DNS_NAME" ] && FUNNEL_URL="https://${DNS_NAME}"
  fi
else
  FUNNEL_ERR=$(cat /tmp/funnel-err)
  # Extract the enable URL from the error message
  ENABLE_URL=$(echo "$FUNNEL_ERR" | grep -oE 'https://login\.tailscale\.com/f/funnel[^[:space:]]*' || true)

  echo ""
  echo "┌──────────────────────────────────────────────────────────────┐"
  echo "│  ⚠️  Tailscale Funnel is not enabled on your tailnet yet     │"
  echo "│                                                              │"
  echo "│  Open this URL in your browser to enable it (one-time):     │"
  echo "│                                                              │"
  if [ -n "$ENABLE_URL" ]; then
  echo "│  $ENABLE_URL"
  else
  echo "│  https://login.tailscale.com/admin/acls  (enable Funnel)    │"
  fi
  echo "│                                                              │"
  echo "│  Then restart this container:  docker compose restart       │"
  echo "└──────────────────────────────────────────────────────────────┘"
  echo ""

  cat > "$STATUS_FILE" <<EOF
{
  "status": "funnel_not_enabled",
  "webhook_url": null,
  "enable_funnel_url": "${ENABLE_URL}",
  "message": "Visit the enable_funnel_url in your browser, then restart the container.",
  "started_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
fi

# ── 4. Write URL to shared volume ────────────────────────────────
if [ -n "$FUNNEL_URL" ]; then
  WEBHOOK_FULL="${FUNNEL_URL}/webhook"

  echo "$FUNNEL_URL"   > "$URL_FILE"
  echo "$WEBHOOK_FULL" > "$FULL_URL_FILE"

  cat > "$STATUS_FILE" <<EOF
{
  "status": "ready",
  "tunnel_url": "${FUNNEL_URL}",
  "webhook_url": "${WEBHOOK_FULL}",
  "port": ${PORT:-3000},
  "hostname": "${TAILSCALE_HOSTNAME:-smiledental-webhook}",
  "started_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

  echo ""
  echo "┌──────────────────────────────────────────────────────────┐"
  echo "│  ✅ Tailscale Funnel is live!                             │"
  echo "│                                                          │"
  echo "│  Tunnel  : ${FUNNEL_URL}"
  echo "│  Webhook : ${WEBHOOK_FULL}"
  echo "│                                                          │"
  echo "│  Paste Webhook URL into Vapi dashboard:                  │"
  echo "│    Phone Number → Server URL                             │"
  echo "│    Assistant → Server URL                                │"
  echo "│                                                          │"
  echo "│  Files written to /shared/                              │"
  echo "└──────────────────────────────────────────────────────────┘"
  echo ""

  export WEBHOOK_URL="$FUNNEL_URL"
fi

# ── 5. Start the webhook server (always, even if funnel failed) ───
echo "[entrypoint] Starting webhook server on port ${PORT:-3000}..."
exec node dist/server.js
