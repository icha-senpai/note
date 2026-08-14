param(
    [string]$TunnelId = $env:CONTROL_PLANE_TUNNEL_ID,
    [string]$ProfileName = "scribli-local",
    [string]$McpServerUrl = "http://127.0.0.1:6806/mcp",
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

if (!(Test-Path -LiteralPath $tunnelClient)) {
    throw "Missing tunnel-client binary: $tunnelClient"
}

if ([string]::IsNullOrWhiteSpace($TunnelId) -and !(Test-Path -LiteralPath $profileFile)) {
    throw "Set CONTROL_PLANE_TUNNEL_ID in this terminal, or pass -TunnelId tunnel_..."
}

if ($PromptForApiKey -and [string]::IsNullOrWhiteSpace($env:CONTROL_PLANE_API_KEY) -and [string]::IsNullOrWhiteSpace($env:OPENAI_API_KEY)) {
    $secureApiKey = Read-Host "CONTROL_PLANE_API_KEY" -AsSecureString
    $apiKeyPtr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secureApiKey)
    try {
        $env:CONTROL_PLANE_API_KEY = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($apiKeyPtr)
    } finally {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($apiKeyPtr)
    }
}

if (!$InitOnly -and [string]::IsNullOrWhiteSpace($env:CONTROL_PLANE_API_KEY) -and [string]::IsNullOrWhiteSpace($env:OPENAI_API_KEY)) {
    throw "Set CONTROL_PLANE_API_KEY in this terminal or run with -PromptForApiKey. Do not paste it into chat."
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
