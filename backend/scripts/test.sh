#!/usr/bin/env bash
# Bash equivalent of test.ps1 for macOS/Linux developers.
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
backend_root="$repo_root/backend"
go_cache="$repo_root/.cache/go-build"

mkdir -p "$go_cache"
export GOCACHE="$go_cache"

cd "$backend_root"
go test ./packages/platform/... ./services/store-edge/... ./services/central-backend/... ./services/hardware-agent/...
