#!/usr/bin/env bash
set -euo pipefail

# Deploy Supabase webhook function, then point your Vapi assistant at it.
#
# Required: VAPI_ASSISTANT_ID (from: vapi assistant list)
#
# CONNECT_ASSISTANT_MODE:
#   full   (default) — PATCH assistant from vapi/riley-assistant.json (replaces prompt/tools copy)
#   server — only set serverUrl on the assistant (keeps your existing agent prompt/model)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CONNECT_ASSISTANT_MODE="${CONNECT_ASSISTANT_MODE:-full}"
VAPI_DIR="${ROOT_DIR}/vapi"
ASSISTANT_JSON="${VAPI_DIR}/riley-assistant.json"

load_env_file() {
  local f="$1"
  [[ -f "$f" ]] || return 0
  set -a
  # shellcheck disable=SC1090
  source "$f"
  set +a
}

load_env_file "${ROOT_DIR}/.env"

if command -v supabase >/dev/null 2>&1; then
  SUPABASE_CMD=(supabase)
elif command -v npx >/dev/null 2>&1; then
  SUPABASE_CMD=(npx supabase)
else
  echo "supabase CLI not found. Install and run 'supabase login' first." >&2
  exit 1
fi

if [[ ! -f "${ASSISTANT_JSON}" ]]; then
  echo "Missing assistant config: ${ASSISTANT_JSON}" >&2
  exit 1
fi

SUPABASE_PROJECT_REF="${SUPABASE_PROJECT_REF:-}"
if [[ -z "${SUPABASE_PROJECT_REF}" && -f "${ROOT_DIR}/supabase/.temp/project-ref" ]]; then
  SUPABASE_PROJECT_REF="$(tr -d '[:space:]' < "${ROOT_DIR}/supabase/.temp/project-ref")"
fi

if [[ -z "${SUPABASE_PROJECT_REF}" ]]; then
  echo "Missing SUPABASE_PROJECT_REF." >&2
  echo "Set SUPABASE_PROJECT_REF in .env or run 'supabase link --project-ref <ref>'." >&2
  exit 1
fi

if [[ -z "${VAPI_API_KEY:-}" && -f "${HOME}/.vapi-cli.yaml" ]]; then
  export VAPI_API_KEY="$(awk -F': ' '/^api_key:/ {print $2; exit}' "${HOME}/.vapi-cli.yaml")"
fi
if [[ -z "${VAPI_API_KEY:-}" ]]; then
  echo "Missing VAPI_API_KEY (set in .env or run vapi login)" >&2
  exit 1
fi

if [[ -z "${VAPI_ASSISTANT_ID:-}" ]]; then
  echo "Missing VAPI_ASSISTANT_ID. Run: vapi assistant list" >&2
  echo "Then set VAPI_ASSISTANT_ID to the pre-existing agent you want to connect." >&2
  exit 1
fi

SERVER_URL="https://${SUPABASE_PROJECT_REF}.supabase.co/functions/v1/webhook"
echo "Using Supabase project ref: ${SUPABASE_PROJECT_REF}"
echo "Deploying function: webhook"
"${SUPABASE_CMD[@]}" functions deploy webhook --project-ref "${SUPABASE_PROJECT_REF}" --no-verify-jwt

TMP_ASSISTANT="$(mktemp /tmp/riley.assistant.XXXXXX.json)"
cleanup() {
  rm -f "${TMP_ASSISTANT}"
}
trap cleanup EXIT

python3 <<PY
import json
from pathlib import Path

src = Path("${ASSISTANT_JSON}")
dst = Path("${TMP_ASSISTANT}")
server_url = "${SERVER_URL}"

assistant = json.loads(src.read_text(encoding="utf-8"))
assistant["serverUrl"] = server_url
dst.write_text(json.dumps(assistant, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
print(f"Prepared assistant config with serverUrl={server_url}")
PY

cp "${TMP_ASSISTANT}" "${ASSISTANT_JSON}"
echo "Updated ${ASSISTANT_JSON}"

API_KEY="${VAPI_API_KEY}"
if [[ -z "${API_KEY}" && -f "${HOME}/.vapi-cli.yaml" ]]; then
  API_KEY="$(awk -F': ' '/^api_key:/ {print $2; exit}' "${HOME}/.vapi-cli.yaml")"
fi

if [[ "${CONNECT_ASSISTANT_MODE}" == "server" ]]; then
  echo "CONNECT_ASSISTANT_MODE=server — PATCH only serverUrl on assistant ${VAPI_ASSISTANT_ID}..."
  code=$(curl -sS -o /tmp/vapi_connect_patch -w "%{http_code}" -X PATCH \
    "https://api.vapi.ai/assistant/${VAPI_ASSISTANT_ID}" \
    -H "Authorization: Bearer ${API_KEY}" \
    -H "Content-Type: application/json" \
    -d "$(python3 -c "import json; print(json.dumps({'serverUrl': '${SERVER_URL}'}))")")
  if [[ "${code}" != "200" ]]; then
    echo "HTTP ${code}" >&2
    cat /tmp/vapi_connect_patch >&2 || true
    exit 1
  fi
  echo "Assistant serverUrl updated (existing prompt/tools unchanged)."
else
  echo "CONNECT_ASSISTANT_MODE=full — pushing vapi/riley-assistant.json to assistant ${VAPI_ASSISTANT_ID}..."
  # No webhook auth: skip Vapi tool header sync unless you set SKIP_TOOL_SYNC=0 and TOOL_API_KEY
  SKIP_TOOL_SYNC="${SKIP_TOOL_SYNC:-1}" bash "${ROOT_DIR}/scripts/vapi/push-assistant-and-sync-tools.sh"
fi

echo "Done."
echo "Webhook URL: ${SERVER_URL}"
