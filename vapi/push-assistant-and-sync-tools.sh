#!/usr/bin/env bash
set -euo pipefail

# Push riley-assistant.json to Vapi (HTTP PATCH) + sync tool server URL + Bearer.
#
# Why not raw `vapi assistant update --file`?
#   Some CLI versions panic (nil client) on `assistant update`, and `vapi tool list`
#   may fail when the API returns new tool types. This matches the official API:
#   PATCH https://api.vapi.ai/assistant/:id  (see https://docs.vapi.ai/cli )
#
# "Swarm" is not a Vapi CLI command. Multi-agent flows use Dashboard (Squad) or
#   `vapi workflow` — see docs.

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ASSISTANT_ID="${VAPI_ASSISTANT_ID:-450435e9-4562-4ddd-8429-54584d3285a7}"

load_env_file() {
  local f="$1"
  [[ -f "$f" ]] || return 0
  set -a
  # shellcheck disable=SC1090
  source "$f"
  set +a
}
load_env_file "$ROOT/vapi-backend/.env"
load_env_file "$ROOT/.env"

if [[ -z "${VAPI_API_KEY:-}" ]]; then
  echo "VAPI_API_KEY missing (vapi-backend/.env or .env)" >&2
  exit 1
fi

TMP="$(mktemp /tmp/riley.push.XXXXXX.json)"
cleanup() { rm -f "$TMP"; }
trap cleanup EXIT

python3 <<PY
import json
from pathlib import Path
root = Path("$ROOT")
cfg_path = root / "vapi" / "riley-assistant.json"
d = json.loads(cfg_path.read_text(encoding="utf-8"))
num = ""
for envp in [root / ".env", root / "vapi-backend" / ".env"]:
    if not envp.exists():
        continue
    for line in envp.read_text(encoding="utf-8").splitlines():
        s = line.strip()
        if s.startswith("VAPI_SMS_FROM_NUMBER="):
            num = s.split("=", 1)[1].strip().strip('"').strip("'")
            break
    if num:
        break
for t in d.get("model", {}).get("tools", []):
    if t.get("type") == "sms":
        md = t.setdefault("metadata", {})
        md["from"] = num if num else md.get("from", "+10000000000")
        break
Path("$TMP").write_text(json.dumps(d, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
PY

echo "PATCH assistant $ASSISTANT_ID ..."
code=$(curl -sS -o /tmp/vapi_patch_body -w "%{http_code}" \
  -X PATCH "https://api.vapi.ai/assistant/${ASSISTANT_ID}" \
  -H "Authorization: Bearer ${VAPI_API_KEY}" \
  -H "Content-Type: application/json" \
  -d @"$TMP")
if [[ "$code" != "200" ]]; then
  echo "HTTP $code" >&2
  cat /tmp/vapi_patch_body >&2 || true
  exit 1
fi
echo "Assistant OK (HTTP 200)"

echo "Sync tool server + Bearer (sync-tool-auth.sh)..."
bash "$ROOT/vapi/sync-tool-auth.sh"

echo "Done."
