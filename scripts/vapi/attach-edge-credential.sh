#!/usr/bin/env bash
set -euo pipefail

# Attach a Vapi *Custom Credential* (Bearer token = Supabase publishable key) to
# the dispatch_dental_action tool. Run AFTER creating the credential in the
# Vapi Dashboard (Credentials → Bearer Token). See vapi/CREDENTIALS.txt
#
# Usage:
#   export VAPI_EDGE_CREDENTIAL_ID='<uuid from dashboard>'
#   ./scripts/vapi/attach-edge-credential.sh
#
# Optional:
#   VAPI_DISPATCH_TOOL_ID  (default: 72cb70d0-b15b-4e5b-8d45-013de0cbcffd)

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
TOOL_ID="${VAPI_DISPATCH_TOOL_ID:-72cb70d0-b15b-4e5b-8d45-013de0cbcffd}"
CRED_ID="${VAPI_EDGE_CREDENTIAL_ID:-}"

API_KEY="${VAPI_API_KEY:-}"
if [[ -z "${API_KEY}" && -f "${HOME}/.vapi-cli.yaml" ]]; then
  API_KEY="$(awk -F': ' '/^api_key:/ {print $2; exit}' "${HOME}/.vapi-cli.yaml")"
fi

if [[ -z "${API_KEY}" ]]; then
  echo "Set VAPI_API_KEY or run: vapi login" >&2
  exit 1
fi

if [[ -z "${CRED_ID}" ]]; then
  echo "Set VAPI_EDGE_CREDENTIAL_ID to your Bearer credential UUID from the Vapi dashboard." >&2
  echo "See ${ROOT}/vapi/CREDENTIALS.txt" >&2
  exit 1
fi

REF="$(tr -d '[:space:]' < "${ROOT}/supabase/.temp/project-ref" 2>/dev/null || true)"
if [[ -z "${REF}" ]]; then
  echo "Missing ${ROOT}/supabase/.temp/project-ref (project ref)" >&2
  exit 1
fi

WEBHOOK="https://${REF}.supabase.co/functions/v1/webhook"

BODY="$(python3 <<PY
import json
print(json.dumps({
  "server": {
    "url": "${WEBHOOK}",
    "timeoutSeconds": 20,
    "credentialId": "${CRED_ID}",
  }
}))
PY
)"

echo "PATCH tool ${TOOL_ID} → server.credentialId=${CRED_ID}"
code=$(curl -sS -o /tmp/vapi_tool_patch_body -w "%{http_code}" -X PATCH \
  "https://api.vapi.ai/tool/${TOOL_ID}" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d "${BODY}")

if [[ "${code}" != "200" ]]; then
  echo "HTTP ${code}" >&2
  cat /tmp/vapi_tool_patch_body >&2 || true
  exit 1
fi

echo "OK (HTTP 200)"
python3 -m json.tool /tmp/vapi_tool_patch_body 2>/dev/null | head -35
