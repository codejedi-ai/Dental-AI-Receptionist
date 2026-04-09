#!/usr/bin/env bash
set -euo pipefail

# Probe the host-run Go Vapi server: /health and GET /api/tools.
# Use for local smoke tests or to verify a public tunnel (same paths Vapi will hit).
#
# Usage:
#   ./vapi-healthcheck.sh                    # http://127.0.0.1:<port>/... from HTTP_LISTEN_ADDR or 127.0.0.1:8080
#   ./vapi-healthcheck.sh --public           # uses PUBLIC_BASE_URL from .env
#   ./vapi-healthcheck.sh --url https://host # explicit base (no trailing slash)
# Env: VAPI_HEALTH_BASE overrides default local URL; loads .env if present.

ROOT="$(cd "$(dirname "$0")" && pwd)"
load_env_file() {
  local f="$1"
  [[ -f "$f" ]] || return 0
  set -a
  # shellcheck disable=SC1090
  source "$f" || exit 1
  set +a
}
load_env_file "$ROOT/.env"
load_env_file "$ROOT/vapi-backend/.env"

MODE="local"
EXPLICIT_URL=""

usage() {
  cat >&2 << 'EOF'
Usage: ./vapi-healthcheck.sh [--public | --url https://host]
  (default)   GET /health and GET /api/tools on loopback (HTTP_LISTEN_ADDR or VAPI_HEALTH_BASE)
  --public    use PUBLIC_BASE_URL from .env (tunnel smoke test)
  --url U     use explicit base URL (no trailing slash)
EOF
  exit 2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --public)
      MODE="public"
      shift
      ;;
    --url)
      EXPLICIT_URL="${2:-}"
      [[ -n "$EXPLICIT_URL" ]] || usage
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      ;;
  esac
done

local_base_from_listen_addr() {
  local addr="${HTTP_LISTEN_ADDR:-127.0.0.1:8080}"
  local port="${addr##*:}"
  echo "http://127.0.0.1:${port}"
}

if [[ -n "$EXPLICIT_URL" ]]; then
  BASE="${EXPLICIT_URL}"
elif [[ "$MODE" == "public" ]]; then
  BASE="${PUBLIC_BASE_URL:-}"
  if [[ -z "$BASE" ]]; then
    echo "PUBLIC_BASE_URL is empty; set it in .env or use --url https://..." >&2
    exit 2
  fi
else
  BASE="${VAPI_HEALTH_BASE:-$(local_base_from_listen_addr)}"
fi
BASE="${BASE%/}"

if ! command -v curl &>/dev/null; then
  echo "curl is required" >&2
  exit 2
fi

fail() {
  echo "❌ $*" >&2
  exit 1
}

ok() {
  echo "✅ $*"
}

echo "Checking base: ${BASE}"

# /health — liveness + JSON shape
body_health=""
if ! body_health=$(curl -fsS --max-time 15 "${BASE}/health"); then
  fail "${BASE}/health — connection failed or non-2xx (is ./run-vapi-local.sh running?)"
fi
if ! grep -qE '"status"[[:space:]]*:[[:space:]]*"ok"' <<<"$body_health"; then
  fail "${BASE}/health — expected JSON with status ok, got: ${body_health:0:200}"
fi
ok "${BASE}/health"

# GET /api/tools — reachability (browser / cellular); Vapi still uses POST
body_tools=""
if ! body_tools=$(curl -fsS --max-time 15 "${BASE}/api/tools"); then
  fail "${BASE}/api/tools (GET) — connection failed or non-2xx"
fi
if ! grep -qE '"ok"[[:space:]]*:[[:space:]]*true' <<<"$body_tools"; then
  fail "${BASE}/api/tools (GET) — expected ok:true, got: ${body_tools:0:200}"
fi
ok "${BASE}/api/tools (GET probe — Vapi uses POST with Bearer)"

echo ""
echo "All checks passed."
