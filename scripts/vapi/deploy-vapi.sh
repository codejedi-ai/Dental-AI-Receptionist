#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ASSISTANT_CONFIG="${ROOT_DIR}/vapi/riley-assistant.json"

if [[ ! -f "${ASSISTANT_CONFIG}" ]]; then
  echo "Missing ${ASSISTANT_CONFIG}" >&2
  exit 1
fi

if [[ -z "${VAPI_ASSISTANT_ID:-}" ]]; then
  echo "Missing VAPI_ASSISTANT_ID" >&2
  exit 1
fi

if [[ -z "${VAPI_API_KEY:-}" && -f "${HOME}/.vapi-cli.yaml" ]]; then
  VAPI_API_KEY="$(awk -F': ' '/^api_key:/ {print $2; exit}' "${HOME}/.vapi-cli.yaml")"
fi

if [[ -z "${VAPI_API_KEY:-}" ]]; then
  echo "Missing VAPI_API_KEY (or run vapi login)" >&2
  exit 1
fi

TMP_CONFIG="$(mktemp /tmp/riley-assistant.deploy.XXXXXX.json)"
cleanup() {
  rm -f "${TMP_CONFIG}"
}
trap cleanup EXIT

python3 <<PY
import json
from pathlib import Path

config_path = Path("${ASSISTANT_CONFIG}")
tmp_path = Path("${TMP_CONFIG}")
payload = json.loads(config_path.read_text(encoding="utf-8"))

num = ""
env_path = Path("${ROOT_DIR}") / ".env"
if env_path.exists():
    for line in env_path.read_text(encoding="utf-8").splitlines():
        s = line.strip()
        if s.startswith("VAPI_SMS_FROM_NUMBER="):
            num = s.split("=", 1)[1].strip().strip('"').strip("'")
            break

for tool in payload.get("model", {}).get("tools", []):
    if tool.get("type") == "sms":
        metadata = tool.setdefault("metadata", {})
        metadata["from"] = num if num else metadata.get("from", "+10000000000")
        break

tmp_path.write_text(json.dumps(payload, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
PY

echo "Deploying assistant ${VAPI_ASSISTANT_ID} to Vapi..."
HTTP_CODE="$(curl -sS -o /tmp/vapi_deploy_response -w "%{http_code}" \
  -X PATCH "https://api.vapi.ai/assistant/${VAPI_ASSISTANT_ID}" \
  -H "Authorization: Bearer ${VAPI_API_KEY}" \
  -H "Content-Type: application/json" \
  -d @"${TMP_CONFIG}")"

if [[ "${HTTP_CODE}" != "200" ]]; then
  echo "Vapi deploy failed (HTTP ${HTTP_CODE})" >&2
  python3 -m json.tool /tmp/vapi_deploy_response 2>/dev/null || true
  exit 1
fi

echo "Vapi deploy complete."
