$ErrorActionPreference = "Stop"

$backendRoot = Split-Path -Parent $PSScriptRoot
$posRoot = Split-Path -Parent $backendRoot
$goCache = Join-Path $posRoot ".cache\go-build"
New-Item -ItemType Directory -Force $goCache | Out-Null

$env:GOCACHE = (Resolve-Path $goCache).Path

Push-Location $backendRoot
try {
    go test `
        ./packages/platform/... `
        ./services/store-edge/... `
        ./services/central-backend/... `
        ./services/hardware-agent/...
}
finally {
    Pop-Location
}
