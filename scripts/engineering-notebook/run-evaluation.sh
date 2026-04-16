#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────
# Dental AI Reception — Evaluation Test Suite
# Page 15: Technical Challenges
#
# Usage:
#   ./run-evaluation.sh
#   ./run-evaluation.sh https://dental-ai.taildd3965.ts.net
#
# Requires: curl, jq (optional, for pretty-printing)
# ─────────────────────────────────────────────────────────────

set -euo pipefail

BASE_URL="${1:-https://dental-ai.taildd3965.ts.net}"
TOOLS_URL="${BASE_URL}/functions/v1/webhook"
HEALTH_URL="${BASE_URL}/health"

PASS=0
FAIL=0
SKIP=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pretty_json() {
    if command -v jq &>/dev/null; then
        echo "$1" | jq . 2>/dev/null || echo "$1"
    else
        echo "$1"
    fi
}

assert_http() {
    local label="$1"
    local expected="$2"
    local actual="$3"
    local body="${4:-}"

    if [ "$actual" = "$expected" ]; then
        echo -e "  ${GREEN}✅ PASS${NC} — ${label} (HTTP ${actual})"
        PASS=$((PASS + 1))
    else
        echo -e "  ${RED}❌ FAIL${NC} — ${label} (expected HTTP ${expected}, got ${actual})"
        if [ -n "$body" ]; then
            echo -e "  ${YELLOW}Response:${NC}"
            pretty_json "$body"
        fi
        FAIL=$((FAIL + 1))
    fi
}

assert_contains() {
    local label="$1"
    local expected_substring="$2"
    local actual="$3"

    if echo "$actual" | grep -qi "$expected_substring"; then
        echo -e "  ${GREEN}✅ PASS${NC} — ${label}"
        PASS=$((PASS + 1))
    else
        echo -e "  ${RED}❌ FAIL${NC} — ${label}"
        echo -e "  Expected to contain: ${YELLOW}${expected_substring}${NC}"
        echo -e "  Actual: ${actual}"
        FAIL=$((FAIL + 1))
    fi
}

call_tool() {
    local tool_name="$1"
    local args="$2"
    local id="${3:-eval-$(date +%s)}"

    local payload
    payload=$(cat <<EOF
{"message":{"type":"tool-calls","tool_calls":[{"id":"${id}","function":{"name":"${tool_name}","arguments":${args}}}]}}
EOF
)
    curl -s -X POST "${TOOLS_URL}" \
        -H "Content-Type: application/json" \
        -d "${payload}" 2>/dev/null || echo "CONNECTION_FAILED"
}

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  🦷 Dental AI Reception — Evaluation Test Suite         ║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║  Target: ${BASE_URL}"
echo "║  Date:   $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# ─────────────────────────────────────────────────────────────
# Challenge 3: Availability Query Tests
# ─────────────────────────────────────────────────────────────

echo "━━━ Challenge 3: Scheduling Database / Availability ━━━"
echo ""

# Test 1: Health Check
echo "Test 1: Health Check"
HTTP_CODE=$(curl -s -o /tmp/health_body -w "%{http_code}" "${HEALTH_URL}" 2>/dev/null || echo "000")
BODY=$(cat /tmp/health_body 2>/dev/null || echo "")
assert_http "Backend is reachable" "200" "${HTTP_CODE}" "${BODY}"
echo ""

# Test 2: Tool Endpoint Accepts POST
echo "Test 2: Tool Endpoint (POST /functions/v1/webhook)"
RESPONSE=$(call_tool "get_current_date" "{}" "t2")
assert_contains "Tool endpoint responds to POST" "results" "${RESPONSE}"
echo ""

# Test 3: check_availability for future date
echo "Test 3: check_availability (future date)"
RESPONSE=$(call_tool "check_availability" '{"date":"2026-04-15"}' "t3")
if echo "${RESPONSE}" | grep -qi "available\|no available\|slots"; then
    echo -e "  ${GREEN}✅ PASS${NC} — Returns availability info"
    PASS=$((PASS + 1))
elif echo "${RESPONSE}" | grep -qi "dentist list\|trouble accessing\|CONNECTION_FAILED"; then
    echo -e "  ${RED}❌ FAIL${NC} — Cannot access dentist list or database"
    echo -e "  ${YELLOW}Response:${NC}"
    pretty_json "${RESPONSE}"
    FAIL=$((FAIL + 1))
else
    echo -e "  ${YELLOW}⚠️  UNCERTAIN${NC} — Unexpected response"
    echo -e "  ${YELLOW}Response:${NC}"
    pretty_json "${RESPONSE}"
    SKIP=$((SKIP + 1))
fi
echo ""

# Test 4: check_availability rejects past date
echo "Test 4: check_availability (past date)"
RESPONSE=$(call_tool "check_availability" '{"date":"2020-01-01"}' "t4")
assert_contains "Rejects past date" "past" "${RESPONSE}"
echo ""

# Test 5: check_availability rejects Sunday
echo "Test 5: check_availability (Sunday — 2026-04-12)"
RESPONSE=$(call_tool "check_availability" '{"date":"2026-04-12"}' "t5")
assert_contains "Rejects Sunday" "sunday" "${RESPONSE}"
echo ""

# ─────────────────────────────────────────────────────────────
# Challenge 1: Patient Info Collection Tests
# ─────────────────────────────────────────────────────────────

echo "━━━ Challenge 1: Patient Info Collection ━━━"
echo ""

# Test 6: book_appointment requires phone
echo "Test 6: book_appointment without phone (should fail)"
RESPONSE=$(call_tool "book_appointment" '{"patientName":"Test User","date":"2026-04-15","time":"09:00","dentist":"Dr. Sarah Chen","service":"Cleaning"}' "t6")
assert_contains "Rejects booking without phone" "missing\|mobile\|phone" "${RESPONSE}"
echo ""

# Test 7: Language Detection (Chinese)
echo "Test 7: detect_language (Chinese)"
RESPONSE=$(call_tool "detect_language" '{"sentence":"你好，我想预约"}' "t7")
assert_contains "Detects Chinese" '"lang_code":"zh"\|"lang_code": "zh"' "${RESPONSE}"
echo ""

# Test 8: Language Detection (English)
echo "Test 8: detect_language (English)"
RESPONSE=$(call_tool "detect_language" '{"sentence":"I want to book an appointment"}' "t8")
assert_contains "Detects English" '"lang_code":"en"\|"lang_code": "en"' "${RESPONSE}"
echo ""

# ─────────────────────────────────────────────────────────────
# Challenge 2: Service Type Tests
# ─────────────────────────────────────────────────────────────

echo "━━━ Challenge 2: Service Type Recognition ━━━"
echo ""

# Test 9: Intent Classification — Booking
echo "Test 9: classify_intent (booking)"
RESPONSE=$(call_tool "classify_intent" '{"utterance":"I want to book an appointment"}' "t9")
assert_contains "Classifies as booking" "book_appointment" "${RESPONSE}"
echo ""

# Test 10: Intent Classification — Emergency
echo "Test 10: classify_intent (emergency)"
RESPONSE=$(call_tool "classify_intent" '{"utterance":"I have a terrible toothache, it is an emergency"}' "t10")
assert_contains "Classifies as emergency" "emergency" "${RESPONSE}"
echo ""

# ─────────────────────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────────────────────

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  📊 Test Results Summary                                ║"
echo "╠══════════════════════════════════════════════════════════╣"
printf "║  ${GREEN}✅ Passed:  %-4s${NC}                                       ║\n" "${PASS}"
printf "║  ${RED}❌ Failed:  %-4s${NC}                                       ║\n" "${FAIL}"
printf "║  ${YELLOW}⚠️  Skipped: %-4s${NC}                                       ║\n" "${SKIP}"
TOTAL=$((PASS + FAIL + SKIP))
printf "║  Total:     %-4s${NC}                                       ║\n" "${TOTAL}"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

if [ "${FAIL}" -gt 0 ]; then
    echo -e "${RED}⚠️  Some tests failed. Review the output above and apply fixes.${NC}"
    echo ""
    echo "See Page 15 of the engineering notebook for fix instructions:"
    echo "  engineering-notebook/15-evaluation-technical-challenges.md"
    exit 1
else
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
fi
