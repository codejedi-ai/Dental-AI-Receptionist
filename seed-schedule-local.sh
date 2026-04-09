#!/usr/bin/env bash
set -euo pipefail

# Seed Excel schedule into local PostgreSQL — no Docker seeder container.
# Prereq: Python 3 + openpyxl + psycopg2-binary; DB from setup-local-databases.sh or equivalent.

ROOT="$(cd "$(dirname "$0")" && pwd)"
EXCEL="${1:-$ROOT/dentist_schedule_6_months.xlsx}"

if [[ ! -f "$EXCEL" ]]; then
  echo "Excel not found: $EXCEL" >&2
  exit 1
fi

if ! command -v python3 &>/dev/null; then
  echo "python3 required" >&2
  exit 1
fi

if ! python3 -c "import openpyxl, psycopg2" 2>/dev/null; then
  echo "Install:  python3 -m pip install --user openpyxl psycopg2-binary" >&2
  exit 1
fi

export DB_HOST="${DB_HOST:-localhost}"
echo "Seeding into PostgreSQL host=$DB_HOST (user dental, db dental)..."
python3 "$ROOT/seed_schedule.py" "$EXCEL"
echo "Done."
