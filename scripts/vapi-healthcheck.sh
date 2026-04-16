#!/usr/bin/env bash
set -euo pipefail

# Probe Supabase webhook function endpoint.
# Uses GET for readiness and POST function-call shape for tool execution reachability.
#
# Usage:
#   ./scripts/vapi-healthcheck.sh                    # local Supabase endpoint
#   ./scripts/vapi-healthcheck.sh --public           # uses PUBLIC_BASE_URL from .env
#   ./scripts/vapi-healthcheck.sh --url https://host # explicit base (no trailing slash)
# Env: VAPI_HEALTH_BASE overrides default local URL; loads .env if present.

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
load_env_file() {
  local f="$1"
  [[ -f "$f" ]] || return 0
  set -a
  # shellcheck disable=SC1090
  source "$f" || exit 1
  set +a
}
load_env_file "$ROOT/.env"

MODE="local"
EXPLICIT_URL=""

usage() {
  cat >&2 << 'EOF2'
Usage: ./scripts/vapi-healthcheck.sh [--public | --url https://host]
  (default)   GET and POST checks on local Supabase webhook endpoint
  --public    use PUBLIC_BASE_URL from .env (tunnel smoke test)
  --url U     use explicit base URL (no trailing slash)
EOF2
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

default_base() {
  echo "http://127.0.0.1:54321"
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
  BASE="${VAPI_HEALTH_BASE:-$(default_base)}"
fi
BASE="${BASE%/}"
WEBHOOK_URL="${BASE}/functions/v1/webhook"

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

echo "Checking webhook base: ${WEBHOOK_URL}"

AUTH_HEADER=()
if [[ -n "${TOOL_API_KEY:-}" ]]; then
  AUTH_HEADER=(-H "Authorization: Bearer ${TOOL_API_KEY}")
fi

# GET readiness check
body_get=""
if ! body_get=$(curl -fsS --max-time 15 "${WEBHOOK_URL}"); then
  fail "${WEBHOOK_URL} (GET) — connection failed or non-2xx (is ./scripts/run-vapi-local.sh running?)"
fi
if ! grep -qE '"ok"[[:space:]]*:[[:space:]]*true' <<<"$body_get"; then
  fail "${WEBHOOK_URL} (GET) — expected ok:true, got: ${body_get:0:200}"
fi
ok "${WEBHOOK_URL} (GET readiness)"

# POST function-call shape check (single router tool)
body_post=""
if ! body_post=$(curl -fsS --max-time 15 "${AUTH_HEADER[@]}" \
  -H "Content-Type: application/json" \
  -X POST "${WEBHOOK_URL}" \
  --data '{"message":{"type":"function-call","functionCall":{"name":"dispatch_dental_action","parameters":{"operation":"get_current_date","payload":{}}}}}'); then
  fail "${WEBHOOK_URL} (POST) — connection failed or non-2xx"
fi
if ! grep -qE '"result"' <<<"$body_post"; then
  fail "${WEBHOOK_URL} (POST) — expected tool result payload, got: ${body_post:0:240}"
fi
ok "${WEBHOOK_URL} (POST function-call)"

echo ""
echo "All checks passed."
