[CmdletBinding()]
param(
    [ValidateRange(1, 3600)]
    [int]$TimeoutSeconds = 120
)

$ErrorActionPreference = "Stop"

$backendRoot = Split-Path -Parent $PSScriptRoot
$posRoot = Split-Path -Parent $backendRoot
$contractsRoot = Join-Path $posRoot "contracts\openapi"
$goCache = Join-Path $posRoot ".cache\go-build"
New-Item -ItemType Directory -Force $goCache | Out-Null
New-Item -ItemType Directory -Force $contractsRoot | Out-Null

$env:GOCACHE = (Resolve-Path $goCache).Path
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)

function Export-OpenAPI {
    param(
        [string]$Package,
        [string]$OutputPath
    )

    $stdoutPath = [System.IO.Path]::GetTempFileName()
    $stderrPath = [System.IO.Path]::GetTempFileName()

    try {
        $process = Start-Process `
            -FilePath "go" `
            -ArgumentList @("run", $Package) `
            -WorkingDirectory $backendRoot `
            -RedirectStandardOutput $stdoutPath `
            -RedirectStandardError $stderrPath `
            -NoNewWindow `
            -PassThru

        if (-not $process.WaitForExit($TimeoutSeconds * 1000)) {
            $process.Kill()
            $process.WaitForExit()
            throw "Timed out exporting OpenAPI from $Package after $TimeoutSeconds seconds."
        }

        $stderr = [System.IO.File]::ReadAllText($stderrPath)
        if ($stderr) {
            [Console]::Error.Write($stderr)
        }
        if ($process.ExitCode -ne 0) {
            exit $process.ExitCode
        }

        $openapi = [System.IO.File]::ReadAllText($stdoutPath)
        [System.IO.File]::WriteAllText($OutputPath, $openapi, $utf8NoBom)
    }
    finally {
        Remove-Item -LiteralPath $stdoutPath, $stderrPath -Force -ErrorAction SilentlyContinue
    }
}

Push-Location $backendRoot
try {
    Export-OpenAPI "./services/store-edge/cmd/export-openapi" (Join-Path $contractsRoot "store-edge.openapi.json")
    Export-OpenAPI "./services/central-backend/cmd/export-openapi" (Join-Path $contractsRoot "central.openapi.json")
    Export-OpenAPI "./services/hardware-agent/cmd/export-openapi" (Join-Path $contractsRoot "hardware-agent.openapi.json")
}
finally {
    Pop-Location
}
