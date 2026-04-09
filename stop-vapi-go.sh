#!/usr/bin/env bash
# Stop local Vapi Go backend (go run ./cmd/main.go or compiled binary from vapi-backend).
set -euo pipefail

any=false
# Parent `go run` and the built test binary often both match
if pkill -f 'vapi-backend/cmd/main\.go' 2>/dev/null; then any=true; fi
if pkill -f 'go run.*/vapi-backend/cmd' 2>/dev/null; then any=true; fi
if pkill -f '/go-build[0-9]+/exe/main' 2>/dev/null; then any=true; fi
if pkill -f '/tmp/go-build[0-9]+/exe/main' 2>/dev/null; then any=true; fi

# Optional: old Docker binary name
pkill -x vapi-server 2>/dev/null && any=true || true

if [[ "$any" == false ]]; then
  echo "No matching Vapi Go processes found."
else
  echo "Stop signals sent."
fi
