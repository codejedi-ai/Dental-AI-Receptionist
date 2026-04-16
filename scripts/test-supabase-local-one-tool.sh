#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BASE_URL="${1:-http://127.0.0.1:54321}"
BASE_URL="${BASE_URL%/}"

echo "[1/3] Webhook one-tool healthcheck"
bash "$ROOT/scripts/vapi-healthcheck.sh" --url "$BASE_URL"

echo "[2/3] Custom tool endpoint check (docs-style tool-calls payload)"
RESP=$(curl -fsS -X POST "$BASE_URL/functions/v1/webhook"   -H "Content-Type: application/json"   --data '{"message":{"type":"tool-calls","toolCallList":[{"id":"toolcall_local_1","name":"get_current_date","arguments":{}}]}}')

if ! echo "$RESP" | grep -q '"results"'; then
  echo "Expected results array from webhook integrated tool-calls, got: $RESP" >&2
  exit 1
fi

if ! echo "$RESP" | grep -q '"toolCallId"'; then
  echo "Expected toolCallId in response, got: $RESP" >&2
  exit 1
fi

echo "Response: $RESP"

echo "[3/3] Done - local one-tool testing passed"
