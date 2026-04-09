#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ASSISTANT_CONFIG="${ROOT_DIR}/vapi/riley-assistant.json"
VAPI_CLI_CONFIG="${HOME}/.vapi-cli.yaml"

if [ ! -f "${ASSISTANT_CONFIG}" ]; then
  echo "❌ Missing ${ASSISTANT_CONFIG}"
  exit 1
fi

if [ ! -f "${VAPI_CLI_CONFIG}" ]; then
  echo "❌ Missing ${VAPI_CLI_CONFIG}. Run: vapi login"
  exit 1
fi

REPO_ROOT="${ROOT_DIR}" python3 <<'PY'
import json
import os
import pathlib
import requests
import yaml

root = pathlib.Path(os.environ["REPO_ROOT"])
assistant_config = root / "vapi" / "riley-assistant.json"
cli_config = pathlib.Path.home() / ".vapi-cli.yaml"
env_candidates = [root / ".env", root / "vapi-backend" / ".env"]

cfg = yaml.safe_load(cli_config.read_text(encoding="utf-8"))
api_key = cfg.get("api_key")
if not api_key:
    raise SystemExit("❌ No api_key found in ~/.vapi-cli.yaml")

tool_api_key = ""
for env_file in env_candidates:
    if not env_file.exists():
        continue
    for line in env_file.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if line.startswith("#") or not line.startswith("TOOL_API_KEY="):
            continue
        tool_api_key = line.split("=", 1)[1].strip().strip("'\"")
        break
    if tool_api_key:
        break
if not tool_api_key:
    raise SystemExit("❌ TOOL_API_KEY missing in repo .env or vapi-backend/.env")

assistant = json.loads(assistant_config.read_text(encoding="utf-8"))
server_url = assistant.get("serverUrl")
if not server_url:
    raise SystemExit("❌ serverUrl missing in vapi/riley-assistant.json")

local_functions = []
for t in assistant.get("model", {}).get("tools", []):
    if t.get("type") == "function" and t.get("function", {}).get("name"):
        local_functions.append(t["function"])

headers = {"Authorization": f"Bearer {api_key}"}
api_headers = {"Authorization": f"Bearer {api_key}", "Content-Type": "application/json"}

resp = requests.get("https://api.vapi.ai/tool", headers=headers, timeout=30)
resp.raise_for_status()
remote_tools = resp.json()

name_to_tool = {}
for t in remote_tools:
    if t.get("type") == "function":
        name = t.get("function", {}).get("name")
        if name:
            name_to_tool[name] = t

synced = 0
missing = []
for fn in local_functions:
    name = fn["name"]
    remote = name_to_tool.get(name)
    if not remote:
        missing.append(name)
        continue
    tool_id = remote["id"]
    payload = {
        "server": {
            "url": server_url,
            "headers": {
                "Authorization": f"Bearer {tool_api_key}",
            },
        },
    }
    patch = requests.patch(
        f"https://api.vapi.ai/tool/{tool_id}",
        headers=api_headers,
        data=json.dumps(payload),
        timeout=30,
    )
    patch.raise_for_status()
    synced += 1
    print(f"✅ synced auth: {name} -> {tool_id}")

print(f"\nDone. Updated {synced} tool(s) with bearer auth.")
if missing:
    print("⚠️ Missing tools in Vapi (create first):")
    for n in missing:
        print(f"  - {n}")
PY

