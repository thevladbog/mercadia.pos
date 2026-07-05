#!/usr/bin/env bash
# Bash equivalent of export-openapi.ps1 for macOS/Linux developers.
#
# Note: unlike export-openapi.ps1 (which enforces a per-export timeout,
# default 120s), this script only applies a timeout when the `timeout`
# command is available on PATH (present on Linux/CI via coreutils, but not
# on stock macOS). On macOS without GNU coreutils installed, exports run
# without a timeout guard.
set -euo pipefail

TIMEOUT_SECONDS="${TIMEOUT_SECONDS:-120}"

script_dir="$(cd "$(dirname "$0")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
backend_root="$repo_root/backend"
contracts_root="$repo_root/contracts/openapi"
go_cache="$repo_root/.cache/go-build"

mkdir -p "$go_cache"
mkdir -p "$contracts_root"
export GOCACHE="$go_cache"

if command -v timeout >/dev/null 2>&1; then
  run_with_timeout() { timeout "$TIMEOUT_SECONDS" "$@"; }
else
  run_with_timeout() { "$@"; }
fi

cd "$backend_root"
run_with_timeout go run ./services/store-edge/cmd/export-openapi > "$contracts_root/store-edge.openapi.json"
run_with_timeout go run ./services/central-backend/cmd/export-openapi > "$contracts_root/central.openapi.json"
run_with_timeout go run ./services/hardware-agent/cmd/export-openapi > "$contracts_root/hardware-agent.openapi.json"
