#!/usr/bin/env bash
set -euo pipefail

# Build and test vapi-backend on this machine only — no Docker.

ROOT="$(cd "$(dirname "$0")" && pwd)"

if ! command -v go &>/dev/null; then
  echo "go is required on PATH" >&2
  exit 1
fi

echo "── vapi-backend: go build ──"
cd "$ROOT/vapi-backend"
go build -o /dev/null ./cmd/main.go

echo "── vapi-backend: go test ./... ──"
go test ./...

echo "Done."
