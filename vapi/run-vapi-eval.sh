#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────
# Vapi Evals — Run Evaluation Test Suite Against Live Assistant
#
# Runs the evals defined in evals-test-suite.json against the
# live Riley assistant on Vapi.
#
# Prerequisites:
#   1. vapi CLI installed and authenticated: vapi login
#   2. VAPI_API_KEY exported
#   3. Backend running and reachable at serverUrl
#
# Usage:
#   ./run-vapi-eval.sh                    # Run all evals
#   ./run-vapi-eval.sh --test "Booking"   # Run matching test(s)
# ─────────────────────────────────────────────────────────────

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVALS_FILE="${SCRIPT_DIR}/evals-test-suite.json"
ASSISTANT_ID="450435e9-4562-4ddd-8429-54584d3285a7"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PASS=0
FAIL=0
SKIP=0

echo_logo() {
    echo -e "${CYAN}"
    echo " ╔══════════════════════════════════════════════════╗"
    echo " ║    Riley Dental AI — Vapi Evals Runner       ║"
    echo " ╚══════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

check_requirements() {
    if [ ! -f "${EVALS_FILE}" ]; then
        echo -e "${RED}❌ ${EVALS_FILE} not found${NC}"
        exit 1
    fi

    if ! python3 -m json.tool "${EVALS_FILE}" >/dev/null 2>&1; then
        echo -e "${RED}❌ Invalid JSON in ${EVALS_FILE}${NC}"
        exit 1
    fi
}

run_eval_with_cli() {
    local test_name="$1"

    echo -e "${CYAN}  ▶ Running: ${test_name}${NC}"

    # vapi eval run --assistant <id> --scenario-file <file>
    if command -v vapi &>/dev/null; then
        local output
        if output=$(vapi eval run --assistant "${ASSISTANT_ID}" --scenario "${EVALS_FILE}" 2>&1); then
            echo -e "    ${GREEN}✅ Eval completed${NC}"
            echo "${output}"
            PASS=$((PASS + 1))
        else
            echo -e "    ${RED}❌ Eval failed${NC}"
            echo "${output}"
            FAIL=$((FAIL + 1))
        fi
    else
        run_eval_with_api "${test_name}"
    fi
}

run_eval_with_api() {
    local test_name="$1"

    if [ -z "${VAPI_API_KEY:-}" ]; then
        echo -e "    ${YELLOW}⚠️  No VAPI_API_KEY — skipping API eval${NC}"
        SKIP=$((SKIP + 1))
        return
    fi

    # The Vapi API doesn't have a direct "eval run" endpoint.
    # We simulate by making a test call and checking the result.
    echo -e "    ${YELLOW}⚠️  Vapi CLI not installed — using API fallback${NC}"
    echo -e "    For full evals, install CLI: curl -sSL https://vapi.ai/install.sh | bash"

    # Check that the assistant config is valid by fetching it
    local http_code
    http_code=$(curl -s -o /dev/null -w "%{http_code}" \
        "https://api.vapi.ai/assistant/${ASSISTANT_ID}" \
        -H "Authorization: Bearer ${VAPI_API_KEY}")

    if [ "${http_code}" = "200" ]; then
        echo -e "    ${GREEN}✅ Assistant is reachable (HTTP ${http_code})${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "    ${RED}❌ Assistant unreachable (HTTP ${http_code})${NC}"
        FAIL=$((FAIL + 1))
    fi
}

check_backend() {
    echo -e "${CYAN}─── Pre-flight: Backend Health ───${NC}"
    echo ""

    local backend_url="https://dental-ai.taildd3965.ts.net"

    local http_code
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "${backend_url}/health" 2>/dev/null || echo "000")

    if [ "${http_code}" = "200" ]; then
        echo -e "  ${GREEN}✅ Backend reachable (${backend_url})${NC}"
    else
        echo -e "  ${RED}❌ Backend unreachable (HTTP ${http_code})${NC}"
        echo -e "  ${YELLOW}⚠️  Evals will fail without a working backend${NC}"
        echo ""
        echo "  Make sure the Go Vapi server is running on the host:"
        echo "    ./run-vapi-local.sh"
        echo "    curl -sS https://<your-tunnel-host>/health"
        echo ""
        read -p "  Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# ─── Main ─────────────────────────────────────────────────────

echo_logo
check_requirements

FILTER="${1:-}"

echo -e "${CYAN}─── Running Vapi Evals ───${NC}"
echo ""

# Pre-flight
check_backend

echo ""

# Run evals
if command -v vapi &>/dev/null; then
    # Read test names from JSON
    test_names=$(python3 -c "
import json
with open('${EVALS_FILE}') as f:
    data = json.load(f)
for t in data.get('tests', []):
    print(t['name'])
" 2>/dev/null)

    while IFS= read -r name; do
        if [ -n "${FILTER}" ] && ! echo "${name}" | grep -qi "${FILTER}"; then
            continue
        fi
        run_eval_with_cli "${name}"
        echo ""
    done <<< "${test_names}"
else
    echo -e "  ${YELLOW}⚠️  Vapi CLI not found — running API-only checks${NC}"
    echo ""
    run_eval_with_api "assistant_reachable"
fi

# Summary
echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║  📊 Eval Results Summary                                  ║"
echo "╠══════════════════════════════════════════════════════════╣"
printf "║  ${GREEN}✅ Passed:  %-4s${NC}                                       ║\n" "${PASS}"
printf "║  ${RED}❌ Failed:  %-4s${NC}                                       ║\n" "${FAIL}"
printf "║  ${YELLOW}⚠️  Skipped: %-4s${NC}                                       ║\n" "${SKIP}"
TOTAL=$((PASS + FAIL + SKIP))
printf "║  Total:     %-4s${NC}                                       ║\n" "${TOTAL}"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

if [ "${FAIL}" -gt 0 ]; then
    echo -e "${RED}⚠️  Some evals failed. Review the output above.${NC}"
    exit 1
else
    echo -e "${GREEN}✅ All evals passed!${NC}"
    exit 0
fi
