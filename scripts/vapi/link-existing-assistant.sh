#!/usr/bin/env bash
# Connect a pre-existing Vapi agent to Supabase (see configure-riley-supabase.sh).
exec "$(cd "$(dirname "$0")" && pwd)/configure-riley-supabase.sh" "$@"
