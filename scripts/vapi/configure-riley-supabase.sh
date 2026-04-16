#!/usr/bin/env bash
set -euo pipefail

# Connect a *pre-existing* Vapi assistant to this repo's Supabase Edge webhook.
#
# Prerequisite: run ./scripts/vapi/connect-supabase-vapi.sh first (deploys webhook + updates
# vapi/riley-assistant.json serverUrl), or set serverUrl in riley-assistant.json yourself.
#
# Required:
#   export VAPI_ASSISTANT_ID='<uuid>'   # vapi assistant list
#
# Optional:
#   LINK_MODE=full|server   (default: full)
#     full   — merge vapi/riley-assistant.json (prompt, voice, toolIds, serverUrl, …)
#     server — only PATCH serverUrl (keeps your existing prompt/model/tools)
#   VAPI_DISPATCH_TOOL_ID   (full mode only; default: 72cb70d0-b15b-4e5b-8d45-013de0cbcffd)

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ASSISTANT_ID="${VAPI_ASSISTANT_ID:-}"
TOOL_ID="${VAPI_DISPATCH_TOOL_ID:-72cb70d0-b15b-4e5b-8d45-013de0cbcffd}"
LINK_MODE="${LINK_MODE:-full}"
ASSISTANT_JSON="${ROOT}/vapi/riley-assistant.json"
TMP_GET="$(mktemp)"
TMP_PATCH="$(mktemp)"
cleanup() {
  rm -f "${TMP_GET}" "${TMP_PATCH}"
}
trap cleanup EXIT

if [[ -z "${ASSISTANT_ID}" ]]; then
  echo "Set VAPI_ASSISTANT_ID to your existing assistant UUID." >&2
  echo "  vapi assistant list" >&2
  echo "  export VAPI_ASSISTANT_ID='...'" >&2
  exit 1
fi

if ! command -v curl &>/dev/null; then
  echo "curl is required" >&2
  exit 1
fi

if [[ ! -f "${ASSISTANT_JSON}" ]]; then
  echo "Missing ${ASSISTANT_JSON}" >&2
  exit 1
fi

API_KEY="${VAPI_API_KEY:-}"
if [[ -z "${API_KEY}" && -f "${HOME}/.vapi-cli.yaml" ]]; then
  API_KEY="$(awk -F': ' '/^api_key:/ {print $2; exit}' "${HOME}/.vapi-cli.yaml")"
fi
if [[ -z "${API_KEY}" ]]; then
  echo "Set VAPI_API_KEY or run: vapi login" >&2
  exit 1
fi

echo "Fetching assistant ${ASSISTANT_ID}..."
curl -sS "https://api.vapi.ai/assistant/${ASSISTANT_ID}" \
  -H "Authorization: Bearer ${API_KEY}" -o "${TMP_GET}"

if [[ "${LINK_MODE}" == "server" ]]; then
  echo "LINK_MODE=server — PATCH serverUrl only (from ${ASSISTANT_JSON})..."
  python3 <<PY
import json
with open("${ASSISTANT_JSON}", encoding="utf-8") as f:
    src = json.load(f)
url = src.get("serverUrl", "").strip()
if not url:
    raise SystemExit("serverUrl missing in riley-assistant.json — run connect-supabase-vapi.sh or set it")
with open("${TMP_PATCH}", "w", encoding="utf-8") as f:
    json.dump({"serverUrl": url}, f)
PY
else
  echo "LINK_MODE=full — merging dental config + toolIds -> ${TOOL_ID}..."
  python3 <<PY
import json
with open("${TMP_GET}", encoding="utf-8") as f:
    asst = json.load(f)
with open("${ASSISTANT_JSON}", encoding="utf-8") as f:
    src = json.load(f)

tool_id = "${TOOL_ID}"
model = asst.get("model", {})
model["provider"] = src["model"]["provider"]
model["model"] = src["model"]["model"]
model["temperature"] = src["model"]["temperature"]
model["messages"] = src["model"]["messages"]
model["toolIds"] = [tool_id]
model.pop("tools", None)

patch = {
    "name": src.get("name", asst.get("name")),
    "model": model,
    "voice": src.get("voice", asst.get("voice")),
    "transcriber": src.get("transcriber", asst.get("transcriber")),
    "firstMessage": src["firstMessage"],
    "voicemailMessage": src["voicemailMessage"],
    "endCallMessage": src["endCallMessage"],
    "endCallPhrases": src.get("endCallPhrases", []),
    "startSpeakingPlan": src.get("startSpeakingPlan", asst.get("startSpeakingPlan")),
    "backgroundDenoisingEnabled": src.get("backgroundDenoisingEnabled", True),
    "backchannelingEnabled": src.get("backchannelingEnabled", False),
    "hipaaEnabled": src.get("hipaaEnabled", False),
    "serverUrl": src["serverUrl"],
    "clientMessages": src.get("clientMessages", asst.get("clientMessages")),
    "serverMessages": src.get("serverMessages", asst.get("serverMessages")),
}
with open("${TMP_PATCH}", "w", encoding="utf-8") as f:
    json.dump(patch, f)
PY
fi

echo "PATCH assistant..."
code=$(curl -sS -o /tmp/vapi_asst_patch_body -w "%{http_code}" -X PATCH \
  "https://api.vapi.ai/assistant/${ASSISTANT_ID}" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d @"${TMP_PATCH}")
if [[ "${code}" != "200" ]]; then
  echo "HTTP ${code}" >&2
  cat /tmp/vapi_asst_patch_body >&2 || true
  exit 1
fi

echo "OK (HTTP 200)"
echo ""
echo "Verify: vapi assistant get ${ASSISTANT_ID}"
