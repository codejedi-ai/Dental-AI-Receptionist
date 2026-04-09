#!/usr/bin/env bash
set -euo pipefail

# Docker is used ONLY for PostgreSQL + MongoDB. Go Vapi runs on the host: ./run-vapi-local.sh
# Seeds Excel → Postgres via local Python (./seed-schedule-local.sh), not a Docker image.

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE="$ROOT/docker-compose.yml"
EXCEL="$ROOT/dentist_schedule_6_months.xlsx"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log()  { echo -e "${GREEN}[✔]${NC} $*"; }
info() { echo -e "${BLUE}[ℹ]${NC} $*"; }
fail() { echo -e "${RED}[✘]${NC} $*"; }

if command -v docker-compose &>/dev/null; then
    DC="docker-compose"
elif docker compose version &>/dev/null 2>&1; then
    DC="docker compose"
else
    fail "docker compose required (for postgres + mongo only)"; exit 1
fi
C() { $DC -f "$COMPOSE" "$@"; }

[[ -f "$EXCEL" ]] || { fail "Excel not found: $EXCEL"; exit 1; }

info "Stopping existing DB containers for this project..."
C down --remove-orphans 2>/dev/null || true

info "Starting PostgreSQL + MongoDB..."
C up -d postgres mongo

info "Waiting for PostgreSQL..."
for i in $(seq 1 30); do
    docker exec dental-postgres pg_isready -U dental &>/dev/null && { log "PostgreSQL ready ($i)"; break; }
    [ "$i" -eq 30 ] && { fail "PostgreSQL timeout"; C logs postgres; exit 1; }
    sleep 2
done

info "Waiting for MongoDB..."
for i in $(seq 1 30); do
    docker exec dental-mongo mongosh --quiet --eval "db.adminCommand('ping')" &>/dev/null && { log "MongoDB ready ($i)"; break; }
    [ "$i" -eq 30 ] && { fail "MongoDB timeout"; C logs mongo; exit 1; }
    sleep 2
done

info "Seeding dentist schedule from Excel (local Python)..."
if ! "$ROOT/seed-schedule-local.sh" "$EXCEL"; then
    fail "Seeding failed — install: python3 -m pip install --user openpyxl psycopg2-binary"
    exit 1
fi

echo ""
echo -e "${BLUE}══════════════════════════════════════════════${NC}"
echo -e "  ${GREEN}DATABASES READY (Docker)${NC}"
echo -e "${BLUE}══════════════════════════════════════════════${NC}"
echo ""

docker exec dental-postgres psql -U dental -d dental -c "
SELECT service, COUNT(*) as slots, COUNT(DISTINCT appointment_date) as days
FROM appointments WHERE status = 'confirmed'
GROUP BY service ORDER BY service;
" 2>/dev/null || true

echo -e "${BLUE}───────────────────────────────────────────────${NC}"
C ps

echo ""
log "Databases are ready."
info "Start Vapi on the host:  $ROOT/run-vapi-local.sh"
info "Tunnel on this PC: ./tailscale-expose-vapi.sh — PUBLIC_BASE_URL + Vapi serverUrl https://.../api/tools"
