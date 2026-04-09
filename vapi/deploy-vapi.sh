#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────
# Vapi CLI — Deploy & Evaluate Riley Dental Assistant
#
# Prerequisites:
#   1. Install Vapi CLI: curl -sSL https://vapi.ai/install.sh | bash
#   2. Authenticate:       vapi login
#   3. Set env var:        export VAPI_API_KEY="your-key-here"
#
# Usage:
#   ./deploy-vapi.sh              # Deploy assistant config
#   ./deploy-vapi.sh --evals      # Deploy config + analysis plan (evals)
#   ./deploy-vapi.sh --dry-run    # Show diff without deploying
# ─────────────────────────────────────────────────────────────

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ASSISTANT_CONFIG="${SCRIPT_DIR}/riley-assistant.json"
ANALYSIS_CONFIG="${SCRIPT_DIR}/analysis-plan.json"
ASSISTANT_ID="450435e9-4562-4ddd-8429-54584d3285a7"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

usage() {
    echo "Usage: $0 [--dry-run] [--evals] [--list] [--get]"
    echo ""
    echo "Options:"
    echo "  (none)     Deploy assistant config to Vapi"
    echo "  --evals    Deploy config + analysis plan (evaluation criteria)"
    echo "  --dry-run  Show what would be deployed without pushing"
    echo "  --list     List all assistants"
    echo "  --get      Get current live config for diff"
    echo "  --test     Run post-deploy validation tests"
    exit 0
}

echo_logo() {
    echo -e "${CYAN}"
    echo " ╔══════════════════════════════════════════════════╗"
    echo " ║    Riley Dental AI — Vapi Deploy & Evaluate  ║"
    echo " ╚══════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

check_cli() {
    if ! command -v vapi &>/dev/null; then
        echo -e "${RED}❌ Vapi CLI not found.${NC}"
        echo ""
        echo "Install it:"
        echo "  curl -sSL https://vapi.ai/install.sh | bash"
        echo ""
        echo "Or use curl fallback (no CLI needed):"
        echo "  curl -X PATCH \"https://api.vapi.ai/assistant/${ASSISTANT_ID}\" \\"
        echo "    -H \"Authorization: Bearer \\\$VAPI_API_KEY\" \\"
        echo "    -H \"Content-Type: application/json\" \\"
        echo "    -d @${ASSISTANT_CONFIG}"
        exit 1
    fi
}

# Merge .env VAPI_SMS_FROM_NUMBER into model.tools[] type=sms (native Send Text).
prepare_assistant_deploy_file() {
    local out="$1"
    python3 <<PY
import json
from pathlib import Path

cfg_path = Path("${ASSISTANT_CONFIG}")
root = Path("${ROOT_DIR}")
d = json.loads(cfg_path.read_text(encoding="utf-8"))
num = ""
envp = root / ".env"
if envp.exists():
    for line in envp.read_text(encoding="utf-8").splitlines():
        s = line.strip()
        if not s or s.startswith("#"):
            continue
        if s.startswith("VAPI_SMS_FROM_NUMBER="):
            num = s.split("=", 1)[1].strip().strip('"').strip("'")
            break
for t in d.get("model", {}).get("tools", []):
    if t.get("type") == "sms":
        md = t.setdefault("metadata", {})
        md["from"] = num if num else md.get("from", "+10000000000")
        break
Path("${out}").write_text(json.dumps(d, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
PY
}

deploy_config() {
    echo -e "${CYAN}─── Deploying Assistant Config ───${NC}"
    echo ""

    local deploy_file
    deploy_file="$(mktemp /tmp/riley-assistant.deploy.XXXXXX.json)"
    prepare_assistant_deploy_file "${deploy_file}"
    trap 'rm -f "${deploy_file}"' RETURN

    if ! python3 -m json.tool "${deploy_file}" >/dev/null 2>&1; then
        echo -e "${RED}❌ Invalid JSON in prepared deploy config${NC}"
        rm -f "${deploy_file}"
        exit 1
    fi
    echo -e "  ${GREEN}✅${NC} Config JSON is valid"

    if grep -q '"type": "sms"' "${deploy_file}" 2>/dev/null; then
        echo -e "  ${GREEN}✅${NC} Native sms (Send Text) tool present — from number from .env VAPI_SMS_FROM_NUMBER (or placeholder)"
    fi

    if [ "${DRY_RUN}" = true ]; then
        echo ""
        echo -e "  ${YELLOW}🔍 Dry run — would deploy to assistant ${ASSISTANT_ID}${NC}"
        echo ""
        python3 -m json.tool "${deploy_file}" | head -40
        echo "  ..."
        rm -f "${deploy_file}"
        return
    fi

    echo -e "  📤 Pushing to Vapi assistant ${ASSISTANT_ID}..."

    # Try Vapi CLI first, fall back to curl
    if command -v vapi &>/dev/null; then
        local output
        if output=$(vapi assistant update "${ASSISTANT_ID}" --file "${deploy_file}" 2>&1); then
            echo -e "  ${GREEN}✅ Assistant config deployed via Vapi CLI${NC}"
        else
            echo -e "  ${YELLOW}⚠️  CLI update returned non-zero — trying curl fallback...${NC}"
            curl_fallback "${deploy_file}"
        fi
    else
        curl_fallback "${deploy_file}"
    fi
    rm -f "${deploy_file}"
}

deploy_evals() {
    echo ""
    echo -e "${CYAN}─── Deploying Analysis Plan (Evaluation Criteria) ───${NC}"
    echo ""

    if [ ! -f "${ANALYSIS_CONFIG}" ]; then
        echo -e "  ${RED}❌ ${ANALYSIS_CONFIG} not found${NC}"
        echo "  Run: ./deploy-vapi.sh --gen-evals to generate it first"
        return 1
    fi

    if ! python3 -m json.tool "${ANALYSIS_CONFIG}" >/dev/null 2>&1; then
        echo -e "${RED}❌ Invalid JSON in ${ANALYSIS_CONFIG}${NC}"
        exit 1
    fi
    echo -e "  ${GREEN}✅${NC} Analysis plan JSON is valid"

    if [ "${DRY_RUN}" = true ]; then
        echo ""
        echo -e "  ${YELLOW}🔍 Dry run — would deploy analysis plan to assistant ${ASSISTANT_ID}${NC}"
        echo ""
        python3 -m json.tool "${ANALYSIS_CONFIG}"
        return
    fi

    echo -e "  📤 Pushing analysis plan to Vapi assistant ${ASSISTANT_ID}..."

    if command -v vapi &>/dev/null; then
        local output
        if output=$(vapi assistant update "${ASSISTANT_ID}" --file "${ANALYSIS_CONFIG}" 2>&1); then
            echo -e "  ${GREEN}✅ Analysis plan deployed via Vapi CLI${NC}"
        else
            echo -e "  ${YELLOW}⚠️  CLI returned non-zero — trying curl fallback...${NC}"
            curl_fallback "${ANALYSIS_CONFIG}"
        fi
    else
        curl_fallback "${ANALYSIS_CONFIG}"
    fi
}

verify_backend_integration() {
    echo ""
    echo -e "${CYAN}─── Verifying Backend Integration ───${NC}"
    echo ""

    local server_url
    server_url=$(python3 - <<'PY'
import json
try:
    with open("vapi/riley-assistant.json", "r", encoding="utf-8") as f:
        cfg = json.load(f)
    print(cfg.get("serverUrl", "").strip())
except Exception:
    print("")
PY
)

    if [ -z "${server_url}" ]; then
        echo -e "  ${YELLOW}⚠️  No serverUrl found in riley-assistant.json${NC}"
        return
    fi

    local health_url="${server_url%/api/tools}/health"
    echo -e "  🔗 Assistant serverUrl: ${server_url}"
    echo -e "  🩺 Health check: ${health_url}"

    local http_code
    http_code=$(curl -s -o /tmp/vapi_backend_health -w "%{http_code}" "${health_url}" 2>/dev/null || echo "000")
    local body
    body=$(cat /tmp/vapi_backend_health 2>/dev/null || true)

    if [ "${http_code}" = "200" ]; then
        echo -e "  ${GREEN}✅ Backend integration reachable (HTTP 200)${NC}"
        echo "  Response: ${body}"
    else
        echo -e "  ${RED}❌ Backend integration check failed (HTTP ${http_code})${NC}"
        echo "  Response: ${body}"
        echo "  Check: ./run-vapi-local.sh running, tunnel up, then curl your public /health URL"
        exit 1
    fi
}

curl_fallback() {
    local config_file="$1"
    if [ -z "${VAPI_API_KEY:-}" ]; then
        echo -e "  ${RED}❌ VAPI_API_KEY not set${NC}"
        echo "  export VAPI_API_KEY=\"your-key-here\""
        exit 1
    fi

    local http_code
    http_code=$(curl -s -o /tmp/vapi_deploy_response -w "%{http_code}" \
        -X PATCH "https://api.vapi.ai/assistant/${ASSISTANT_ID}" \
        -H "Authorization: Bearer ${VAPI_API_KEY}" \
        -H "Content-Type: application/json" \
        -d @"${config_file}" 2>/dev/null)

    local body
    body=$(cat /tmp/vapi_deploy_response 2>/dev/null)

    if [ "${http_code}" = "200" ]; then
        echo -e "  ${GREEN}✅ Deployed successfully (HTTP ${http_code})${NC}"
    elif [ "${http_code}" = "401" ]; then
        echo -e "  ${RED}❌ Unauthorized (401) — check VAPI_API_KEY${NC}"
        exit 1
    elif [ "${http_code}" = "404" ]; then
        echo -e "  ${RED}❌ Assistant not found (404) — check ASSISTANT_ID${NC}"
        exit 1
    else
        echo -e "  ${RED}❌ HTTP ${http_code}${NC}"
        echo "  Response: $(echo "${body}" | python3 -m json.tool 2>/dev/null || echo "${body}")"
        exit 1
    fi
}

get_live_config() {
    echo -e "${CYAN}─── Fetching Live Config from Vapi ───${NC}"
    echo ""

    if command -v vapi &>/dev/null; then
        vapi assistant get "${ASSISTANT_ID}" | python3 -m json.tool 2>/dev/null || \
            vapi assistant get "${ASSISTANT_ID}"
    else
        if [ -z "${VAPI_API_KEY:-}" ]; then
            echo -e "${RED}❌ VAPI_API_KEY not set${NC}"
            exit 1
        fi
        curl -s "https://api.vapi.ai/assistant/${ASSISTANT_ID}" \
            -H "Authorization: Bearer ${VAPI_API_KEY}" | python3 -m json.tool
    fi
}

list_assistants() {
    echo -e "${CYAN}─── All Vapi Assistants ───${NC}"
    echo ""

    if command -v vapi &>/dev/null; then
        vapi assistant list | python3 -m json.tool 2>/dev/null || vapi assistant list
    else
        if [ -z "${VAPI_API_KEY:-}" ]; then
            echo -e "${RED}❌ VAPI_API_KEY not set${NC}"
            exit 1
        fi
        curl -s "https://api.vapi.ai/assistant" \
            -H "Authorization: Bearer ${VAPI_API_KEY}" | python3 -m json.tool
    fi
}

# ─── Main ─────────────────────────────────────────────────────

DRY_RUN=false
DEPLOY_EVALS=false
ACTION="deploy"

for arg in "$@"; do
    case "$arg" in
        --dry-run)  DRY_RUN=true ;;
        --evals)    DEPLOY_EVALS=true ;;
        --list)     ACTION="list" ;;
        --get)      ACTION="get" ;;
        --help|-h)  usage ;;
        *)          echo -e "${RED}Unknown option: ${arg}${NC}"; usage ;;
    esac
done

echo_logo

case "${ACTION}" in
    list)
        list_assistants
        ;;
    get)
        get_live_config
        ;;
    deploy)
        check_cli
        deploy_config
        if [ "${DEPLOY_EVALS}" = true ]; then
            deploy_evals
        fi
        verify_backend_integration
        echo ""
        echo -e "${GREEN}✅ Done.${NC}"
        echo ""
        echo -e "Next steps:"
        echo "  1. Test the assistant: call the assigned phone number or use Vapi dashboard"
        echo "  2. Check call results: vapi call list"
        echo "  3. View evals: vapi call get <call-id>"
        echo "  4. Run evaluation test: ${SCRIPT_DIR}/run-vapi-eval.sh"
        ;;
esac
