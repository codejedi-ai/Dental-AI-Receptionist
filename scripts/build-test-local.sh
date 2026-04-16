#!/usr/bin/env bash
set -euo pipefail

# Validate serverless local runtime prerequisites.

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

if command -v supabase &>/dev/null; then
  SUPABASE_CMD=(supabase)
elif command -v npx &>/dev/null; then
  SUPABASE_CMD=(npx supabase)
else
  echo "supabase CLI is required on PATH (or install npx)." >&2
  exit 1
fi

echo "── supabase: status ──"
"${SUPABASE_CMD[@]}" status >/dev/null

echo "── webhook function smoke check ──"
if [[ ! -f "$ROOT/supabase/functions/webhook/index.ts" ]]; then
  echo "missing supabase/functions/webhook/index.ts" >&2
  exit 1
fi

echo "Done. Serverless runtime prerequisites look good."
