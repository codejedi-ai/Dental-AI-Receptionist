#!/usr/bin/env bash
set -euo pipefail

# Run the Supabase webhook function locally for Vapi tool/function callbacks.
# This supersedes the legacy Go backend path.

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if command -v supabase &>/dev/null; then
  SUPABASE_CMD=(supabase)
elif command -v npx &>/dev/null; then
  SUPABASE_CMD=(npx supabase)
else
  echo "supabase CLI is required on PATH (or install npx)." >&2
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

# Repo root .env (optional)
load_env_file "$ROOT/.env"
export PUBLIC_BASE_URL="${PUBLIC_BASE_URL:-}"

if [[ -n "$PUBLIC_BASE_URL" ]]; then
  echo "[run-vapi-local] PUBLIC_BASE_URL=$PUBLIC_BASE_URL"
  echo "[run-vapi-local] Vapi serverUrl: ${PUBLIC_BASE_URL%/}/functions/v1/webhook"
else
  echo "[run-vapi-local] Set PUBLIC_BASE_URL to your tunnel HTTPS origin, then set Vapi serverUrl to .../functions/v1/webhook"
fi

echo ""
echo "Starting local Supabase stack if needed..."
"${SUPABASE_CMD[@]}" start >/dev/null

echo "Serving Supabase Edge Function: webhook"
echo "Local endpoint: http://127.0.0.1:54321/functions/v1/webhook"
echo ""
echo "Health: ${ROOT}/scripts/vapi-healthcheck.sh  |  Funnel + Vapi CLI: see ${ROOT}/README.md"
echo ""

if [[ -f "$ROOT/.env" ]]; then
  exec "${SUPABASE_CMD[@]}" functions serve webhook --no-verify-jwt --env-file "$ROOT/.env"
else
  exec "${SUPABASE_CMD[@]}" functions serve webhook --no-verify-jwt
fi
