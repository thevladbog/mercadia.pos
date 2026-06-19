$ErrorActionPreference = "Stop"

$backendRoot = Split-Path -Parent $PSScriptRoot
$posRoot = Split-Path -Parent $backendRoot
$contractsRoot = Join-Path $posRoot "contracts\openapi"
$goCache = Join-Path $posRoot ".cache\go-build"
New-Item -ItemType Directory -Force $goCache | Out-Null
New-Item -ItemType Directory -Force $contractsRoot | Out-Null

$env:GOCACHE = (Resolve-Path $goCache).Path

Push-Location $backendRoot
try {
    go run ./services/store-edge/cmd/export-openapi | Set-Content -Encoding utf8 (Join-Path $contractsRoot "store-edge.openapi.json")
    go run ./services/central-backend/cmd/export-openapi | Set-Content -Encoding utf8 (Join-Path $contractsRoot "central.openapi.json")
    go run ./services/hardware-agent/cmd/export-openapi | Set-Content -Encoding utf8 (Join-Path $contractsRoot "hardware-agent.openapi.json")
}
finally {
    Pop-Location
}
