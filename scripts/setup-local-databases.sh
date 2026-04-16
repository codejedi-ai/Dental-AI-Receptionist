#!/usr/bin/env bash
set -euo pipefail

# Apply PostgreSQL schema + MongoDB init on localhost — no Docker.
# Expects Postgres and MongoDB installed and listening on default ports.
# Same DB user/database names as docker-compose postgres service (dental / dental).

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export PGHOST="${PGHOST:-127.0.0.1}"
export PGPORT="${PGPORT:-5432}"
export PGUSER="${PGUSER:-dental}"
export PGPASSWORD="${PGPASSWORD:-internal_pg_2024}"
export PGDATABASE="${PGDATABASE:-dental}"

psql_ok() { PGPASSWORD="$PGPASSWORD" psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -v ON_ERROR_STOP=1 "$@"; }

echo "── PostgreSQL: $PGHOST:$PGPORT db=$PGDATABASE user=$PGUSER ──"
if ! command -v psql &>/dev/null; then
  echo "psql not found. Install PostgreSQL client." >&2
  exit 1
fi

if ! psql_ok -c "SELECT 1" &>/dev/null; then
  cat >&2 << EOF
Cannot connect to PostgreSQL. Create role/database once (as superuser), e.g.:

  createuser -h $PGHOST -U postgres -P dental
  createdb -h $PGHOST -U postgres -O dental dental

Or set PGHOST PGUSER PGPASSWORD PGDATABASE to match your local install.
EOF
  exit 1
fi

psql_ok -f "$ROOT/db/init.sql"
echo "✅ PostgreSQL init.sql applied"

echo "── MongoDB: ${MONGO_URL:-mongodb://127.0.0.1:27017/dental} ──"
if ! command -v mongosh &>/dev/null; then
  echo "mongosh not found. Install MongoDB shell / database." >&2
  exit 1
fi

MONGO_URI="${MONGO_URI:-mongodb://127.0.0.1:27017/dental}"
mongosh "$MONGO_URI" --quiet --file "$ROOT/db/mongo-init.js"
echo "✅ MongoDB mongo-init.js applied"

echo ""
echo "Next: ./scripts/seed-schedule-local.sh  (optional)  then  ./scripts/run-vapi-local.sh"
