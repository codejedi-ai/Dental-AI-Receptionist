#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SUPABASE_SCRIPT="${ROOT_DIR}/scripts/vapi/deploy-supabase.sh"
VAPI_SCRIPT="${ROOT_DIR}/scripts/vapi/deploy-vapi.sh"
ASSISTANT_CONFIG="${ROOT_DIR}/vapi/riley-assistant.json"

if [[ ! -f "${ASSISTANT_CONFIG}" ]]; then
  echo "Missing ${ASSISTANT_CONFIG}" >&2
  exit 1
fi

"${SUPABASE_SCRIPT}"

PROJECT_REF="${SUPABASE_PROJECT_REF:-}"
if [[ -z "${PROJECT_REF}" && -f "${ROOT_DIR}/supabase/.temp/project-ref" ]]; then
  PROJECT_REF="$(tr -d '[:space:]' < "${ROOT_DIR}/supabase/.temp/project-ref")"
fi

if [[ -z "${PROJECT_REF}" ]]; then
  echo "Missing SUPABASE_PROJECT_REF after Supabase deploy" >&2
  exit 1
fi

TMP_CONFIG="$(mktemp /tmp/riley-assistant.with-server.XXXXXX.json)"
cleanup() {
  rm -f "${TMP_CONFIG}"
}
trap cleanup EXIT

python3 <<PY
import json
from pathlib import Path

src = Path("${ASSISTANT_CONFIG}")
dst = Path("${TMP_CONFIG}")
payload = json.loads(src.read_text(encoding="utf-8"))
payload["serverUrl"] = "https://${PROJECT_REF}.supabase.co/functions/v1/webhook"
dst.write_text(json.dumps(payload, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
PY

cp "${TMP_CONFIG}" "${ASSISTANT_CONFIG}"
echo "Updated serverUrl in vapi/riley-assistant.json"

"${VAPI_SCRIPT}"

echo "Both deploys complete."
