#!/usr/bin/env bash
# Reference only — run these commands ON the backend host (machine in your tailnet).
# Default local port matches Supabase API in local dev.
#
# Docs: https://tailscale.com/kb/1242/tailscale-serve
#       https://tailscale.com/kb/1223/tailscale-funnel
#
# Funnel public HTTPS must use one of: 443, 8443, 10000 (TLS terminated by tailscaled).

set -euo pipefail
# Default local API port for Supabase.
PORT="${VAPI_LOCAL_PORT:-54321}"

cat << EOF
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Local Supabase webhook → Tailscale Serve / Funnel  (proxy to 127.0.0.1:${PORT})
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
All services stay on this machine; Funnel only exposes loopback to Vapi.

Why Funnel must run ON this same machine
────────────────────────────────────────
Vapi’s cloud sends HTTPS requests TO your serverUrl (tool webhooks). That traffic
enters the public internet, hits Tailscale’s Funnel relay, then comes IN to this
host’s tailscaled. So “outbound from Vapi” = inbound to YOU — only works if:

  • This machine is online, tailscaled is up, and
  • Funnel is enabled here (below), proxying to your Supabase API on 127.0.0.1:${PORT}.

Use  tailscale funnel status  to confirm Funnel is active after you enable it.
--bg  keeps the config across restarts (see Tailscale docs).

1) Start local Supabase + webhook server:
     ./scripts/run-vapi-local.sh
  (Override this help with VAPI_LOCAL_PORT=...)

2a) Tailnet-only — Serve (HTTPS on this machine, MagicDNS):
     sudo tailscale serve --bg --https=443 http://127.0.0.1:${PORT}

   Path-only mount (optional):
     sudo tailscale serve --bg --https=443 --set-path=/functions/v1/webhook http://127.0.0.1:${PORT}

2b) Public internet — Funnel (required for Vapi; TLS on port 443):
     sudo tailscale funnel --bg --yes --https=443 http://127.0.0.1:${PORT}

   Shorthand (local target port ${PORT}; public URL still HTTPS :443):
     sudo tailscale funnel --bg --yes ${PORT}

   Leave this in place whenever you want Vapi tool calls to reach this PC.
   If Funnel is off, Vapi cannot open a connection → no tool-calls in the dashboard.

3) Verify from a non-tailnet device (cellular):
  https://<your-machine>.<tailnet>.ts.net/functions/v1/webhook
   Expect HTTP 200 + JSON on GET (probe). Vapi uses POST.

4) Status / off:
     tailscale serve status
     tailscale funnel status
     sudo tailscale funnel --https=443 http://127.0.0.1:${PORT} off
     sudo tailscale serve --https=443 http://127.0.0.1:${PORT} off

Note: "tailscale funnel 443 on" is not valid. Use funnel + target URL or port as above.
See also: ./scripts/tailscale-expose-vapi.sh (this file) — https://tailscale.com/kb/1223
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EOF
