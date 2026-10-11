param(
    [string]$TunnelId = $env:CONTROL_PLANE_TUNNEL_ID,
    [string]$ProfileName = "scribli-local",
    [string]$McpServerUrl = "http://127.0.0.1:6806/mcp",
    [string]$ApiKeyFile = "",
    [switch]$DoctorOnly,
    [switch]$InitOnly,
    [switch]$OpenWebUi,
    [switch]$PromptForApiKey
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$toolRoot = Join-Path $repoRoot ".tools\tunnel-client"
$tunnelClient = Join-Path $toolRoot "v0.0.11\tunnel-client.exe"
$profileDir = Join-Path $toolRoot "profiles"
$profileFile = Join-Path $profileDir "$ProfileName.yaml"
$healthUrlFile = Join-Path $toolRoot "scribli-tunnel-health.url"
if ([string]::IsNullOrWhiteSpace($ApiKeyFile)) {
    $ApiKeyFile = Join-Path $repoRoot "tmp\key.txt"
}

function Has-ControlPlaneApiKey {
    return ![string]::IsNullOrWhiteSpace($env:CONTROL_PLANE_API_KEY) -or ![string]::IsNullOrWhiteSpace($env:OPENAI_API_KEY)
}

function Set-ControlPlaneApiKeyFromFile {
    param([string]$Path)

    if ((Has-ControlPlaneApiKey) -or [string]::IsNullOrWhiteSpace($Path) -or !(Test-Path -LiteralPath $Path)) {
        return
    }

    $apiKey = [string](Get-Content -LiteralPath $Path -Raw)
    $apiKey = $apiKey.Trim()
    if (![string]::IsNullOrWhiteSpace($apiKey)) {
        $env:CONTROL_PLANE_API_KEY = $apiKey
        Write-Host "Loaded CONTROL_PLANE_API_KEY from $Path"
    }
}

function Save-ControlPlaneApiKeyToFile {
    param(
        [string]$Path,
        [string]$ApiKey
    )

    if ([string]::IsNullOrWhiteSpace($Path) -or [string]::IsNullOrWhiteSpace($ApiKey)) {
        return
    }

    $directory = Split-Path -Parent $Path
    if (![string]::IsNullOrWhiteSpace($directory)) {
        New-Item -ItemType Directory -Force -Path $directory | Out-Null
    }
    Set-Content -LiteralPath $Path -Value $ApiKey -NoNewline
    Write-Host "Saved CONTROL_PLANE_API_KEY to $Path"
}

Set-ControlPlaneApiKeyFromFile -Path $ApiKeyFile

if (!(Test-Path -LiteralPath $tunnelClient)) {
    throw "Missing tunnel-client binary: $tunnelClient"
}

if ([string]::IsNullOrWhiteSpace($TunnelId) -and !(Test-Path -LiteralPath $profileFile)) {
    throw "Set CONTROL_PLANE_TUNNEL_ID in this terminal, or pass -TunnelId tunnel_..."
}

if ($PromptForApiKey -and !(Has-ControlPlaneApiKey)) {
    $secureApiKey = Read-Host "CONTROL_PLANE_API_KEY" -AsSecureString
    $apiKeyPtr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secureApiKey)
    try {
        $env:CONTROL_PLANE_API_KEY = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($apiKeyPtr)
        Save-ControlPlaneApiKeyToFile -Path $ApiKeyFile -ApiKey $env:CONTROL_PLANE_API_KEY
    } finally {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($apiKeyPtr)
    }
}

if (!$InitOnly -and !(Has-ControlPlaneApiKey)) {
    throw "Set CONTROL_PLANE_API_KEY in this terminal, put it in $ApiKeyFile, or run with -PromptForApiKey."
}

New-Item -ItemType Directory -Force -Path $profileDir | Out-Null

try {
    $request = [System.Net.HttpWebRequest]::Create($McpServerUrl)
    $request.Method = "GET"
    $request.Timeout = 5000
    $request.Accept = "text/event-stream"

    try {
        $response = $request.GetResponse()
        $statusCode = [int]$response.StatusCode
        $reader = New-Object System.IO.StreamReader($response.GetResponseStream())
        $content = $reader.ReadToEnd()
    } catch [System.Net.WebException] {
        if ($_.Exception.Response -eq $null) {
            throw
        }
        $response = $_.Exception.Response
        $statusCode = [int]$response.StatusCode
        $reader = New-Object System.IO.StreamReader($response.GetResponseStream())
        $content = $reader.ReadToEnd()
    } finally {
        if ($reader -ne $null) {
            $reader.Dispose()
        }
        if ($response -ne $null) {
            $response.Dispose()
        }
    }

    $healthyGet = $statusCode -eq 400 -and $content -match "Mcp-Session-Id"
    if (!$healthyGet -and ($statusCode -lt 200 -or $statusCode -ge 300)) {
        throw "Unexpected MCP status code ${statusCode}: $content"
    }
} catch {
    throw "Scribli MCP endpoint is not reachable at $McpServerUrl. Start Scribli first. Details: $($_.Exception.Message)"
}

if (![string]::IsNullOrWhiteSpace($TunnelId)) {
    & $tunnelClient init `
        --sample sample_mcp_remote_no_auth `
        --profile $ProfileName `
        --profile-dir $profileDir `
        --tunnel-id $TunnelId `
        --mcp-server-url $McpServerUrl `
        --health-listen-addr "127.0.0.1:0" `
        --force

    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
} else {
    Write-Host "Using existing tunnel-client profile '$ProfileName' in $profileDir"
}

if ($InitOnly) {
    Write-Host "Tunnel-client profile '$ProfileName' is ready in $profileDir"
    exit 0
}

& $tunnelClient doctor --profile $ProfileName --profile-dir $profileDir --explain
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

if ($DoctorOnly) {
    exit 0
}

$runArgs = @(
    "run",
    "--profile", $ProfileName,
    "--profile-dir", $profileDir,
    "--health.url-file", $healthUrlFile
)

if ($OpenWebUi) {
    $runArgs += "--open-web-ui"
}

& $tunnelClient @runArgs
