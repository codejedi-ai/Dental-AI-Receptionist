#!/usr/bin/env bash
set -euo pipefail

# Run the Go Vapi backend locally on this machine only (plain HTTP on loopback by default).
# Do not run this process on a remote server — use a tunnel on this host if Vapi needs a public URL.
#
# Prereq: PostgreSQL + MongoDB reachable on localhost (e.g. docker compose + published ports).
# Public URL: tailscale funnel / similar → 127.0.0.1:<port>. See ./tailscale-expose-vapi.sh

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT/vapi-backend"

if ! command -v go &>/dev/null; then
  echo "go is required on PATH" >&2
  exit 1
fi

load_env_file() {
  local f="$1"
  [[ -f "$f" ]] || return 0
  set -a
  # shellcheck disable=SC1090
  source "$f" || {
    echo "Failed to source $f (check for unquoted spaces in values — use quotes around values with spaces)." >&2
    exit 1
  }
  set +a
}

# Repo root .env (optional), then vapi-backend/.env (optional) — latter overrides.
load_env_file "$ROOT/.env"
load_env_file "$ROOT/vapi-backend/.env"

export DATABASE_URL="${DATABASE_URL:-postgres://dental:internal_pg_2024@127.0.0.1:5432/dental}"
export MONGO_URL="${MONGO_URL:-mongodb://127.0.0.1:27017/dental}"
export MONGO_DB="${MONGO_DB:-dental}"
export HTTP_LISTEN_ADDR="${HTTP_LISTEN_ADDR:-127.0.0.1:8080}"
# Set to your public tunnel URL for logs / Vapi serverUrl, e.g. https://dental-ai.your-tailnet.ts.net
export PUBLIC_BASE_URL="${PUBLIC_BASE_URL:-}"

if [[ -z "${TOOL_API_KEY:-}" ]]; then
  echo "[run-vapi-local] ⚠️  TOOL_API_KEY unset — server accepts any POST /api/tools (dev only)." >&2
  echo "    For production: set TOOL_API_KEY in vapi-backend/.env and run vapi/sync-tool-auth.sh" >&2
fi

echo "[run-vapi-local] HTTP_LISTEN_ADDR=$HTTP_LISTEN_ADDR"
echo "[run-vapi-local] DATABASE_URL host should be 127.0.0.1/localhost (see enforceLocalDatastore in main.go)"
if [[ -n "$PUBLIC_BASE_URL" ]]; then
  echo "[run-vapi-local] PUBLIC_BASE_URL=$PUBLIC_BASE_URL  → Vapi serverUrl: ${PUBLIC_BASE_URL%/}/api/tools"
else
  echo "[run-vapi-local] Set PUBLIC_BASE_URL to your tunnel HTTPS origin, then point Vapi serverUrl at .../api/tools"
fi
LOCAL_PORT="${HTTP_LISTEN_ADDR##*:}"
echo ""
echo "Tailscale (on this host): see exact Serve/Funnel commands:"
echo "  ./tailscale-expose-vapi.sh   (or: VAPI_LOCAL_PORT=${LOCAL_PORT} ./tailscale-expose-vapi.sh)"
echo ""
echo "Health: ${ROOT}/vapi-healthcheck.sh  |  Funnel + Vapi CLI: see ${ROOT}/README.md"
echo ""

exec go run ./cmd/main.go
