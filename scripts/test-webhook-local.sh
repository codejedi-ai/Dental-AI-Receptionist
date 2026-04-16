#!/usr/bin/env bash
# Run the webhook Edge Function with Deno and send mock Vapi-style requests.
# Does not require `supabase start` (Docker). Install Deno once: https://deno.land
# Tool responses use { "results": [ { "toolCallId", "result" } ] } per Vapi server-url docs.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WEBHOOK_DIR="$ROOT/supabase/functions/webhook"
IMPORT_MAP="$ROOT/supabase/functions/import_maps.json"
DENO="${DENO:-$HOME/.deno/bin/deno}"
PORT="${PORT:-8000}"
BASE="http://127.0.0.1:${PORT}/"

if [[ ! -x "$DENO" ]]; then
  echo "Deno not found at $DENO. Install: curl -fsSL https://deno.land/install.sh | sh"
  exit 1
fi

echo "Starting webhook (Deno) on ${BASE} ..."
cd "$WEBHOOK_DIR"
"$DENO" run --allow-net --allow-env --import-map="$IMPORT_MAP" index.ts &
PID=$!
cleanup() { kill "$PID" 2>/dev/null || true; }
trap cleanup EXIT

for _ in $(seq 1 45); do
  if curl -sf "$BASE" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

if ! curl -sf "$BASE" >/dev/null 2>&1; then
  echo "Server did not become ready on $BASE"
  exit 1
fi

echo ""
echo "=== GET (health) ==="
curl -sS "$BASE" | jq .

echo ""
echo "=== POST tool_calls OpenAI shape (id + function.name + function.arguments string) ==="
curl -sS -X POST "$BASE" -H 'Content-Type: application/json' \
  -d '{"message":{"type":"tool-calls","tool_calls":[{"id":"call_Qivp6mhnqrhdclur3DNGn1eW","type":"function","function":{"name":"dispatch_dental_action","arguments":"{\"operation\":\"get_current_date\",\"payload\":{}}"}}]}}' | jq .

echo ""
echo "=== POST toolCallList at ROOT (no message wrapper — Chat/custom-tool style) ==="
curl -sS -X POST "$BASE" -H 'Content-Type: application/json' \
  -d '{"toolCallList":[{"id":"call_Qivp6mhnqrhdclur3DNGn1eW","name":"dispatch_dental_action","parameters":{"operation":"get_current_date"}}]}' | jq .

echo ""
echo "=== POST tool-calls + toolCallList (Vapi Chat shape) dispatch_dental_action, id matches ==="
curl -sS -X POST "$BASE" -H 'Content-Type: application/json' \
  -d '{"message":{"type":"tool-calls","toolCallList":[{"id":"call_TdR1g9nfcdjQcjap0J9LDuM4","name":"dispatch_dental_action","parameters":{"operation":"get_current_date"}}]}}' | jq .

echo ""
echo "=== POST toolCallList only: dispatch_dental_action (no payload key) ==="
curl -sS -X POST "$BASE" -H 'Content-Type: application/json' \
  -d '{"message":{"toolCallList":[{"id":"tc2","name":"dispatch_dental_action","parameters":{"operation":"get_current_date"}}]}}' | jq .

echo ""
echo "=== POST function-call + OpenAI-style arguments string + toolCall id ==="
curl -sS -X POST "$BASE" -H 'Content-Type: application/json' \
  -d '{"message":{"type":"function-call","functionCall":{"id":"call_abc123","name":"dispatch_dental_action","arguments":"{\"operation\":\"get_current_date\"}"}}}' | jq .

echo ""
echo "=== POST function-call: dispatch_dental_action -> get_clinic_info ==="
curl -sS -X POST "$BASE" -H 'Content-Type: application/json' \
  -d '{"message":{"type":"function-call","functionCall":{"id":"call_x","name":"dispatch_dental_action","parameters":{"operation":"get_clinic_info","payload":{"topic":"hours"}}}}}' | jq .

echo ""
echo "=== POST assistant-request (empty body; assistant lives in Vapi) ==="
curl -sS -X POST "$BASE" -H 'Content-Type: application/json' \
  -d '{"message":{"type":"assistant-request","call":{}}}' | jq .

echo ""
echo "Done. (Tools that hit Supabase REST need SUPABASE_URL + key in env when testing DB calls.)"
