param(
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\Scribli",
    [int]$Port = 6806,
    [int]$ReadyTimeoutSeconds = 20,
    [switch]$NoReadyCheck
)

$ErrorActionPreference = "Stop"

$exe = Join-Path $InstallDir "Scribli.exe"
if (-not (Test-Path $exe)) {
    throw "Scribli.exe was not found at $exe"
}

# Codex shells can inject this for Node tooling; Scribli must inherit a normal Electron environment.
Remove-Item Env:ELECTRON_RUN_AS_NODE -ErrorAction SilentlyContinue

$process = Start-Process $exe -PassThru
Write-Host "Started Scribli PID $($process.Id)"

if ($NoReadyCheck) {
    return
}

$deadline = (Get-Date).AddSeconds($ReadyTimeoutSeconds)
$versionUri = "http://127.0.0.1:$Port/api/system/version"
while ((Get-Date) -lt $deadline) {
    try {
        $response = Invoke-RestMethod -Method Post -Uri $versionUri -TimeoutSec 2
        Write-Host "Scribli kernel API ready on port $Port, version $($response.data)"
        return
    } catch {
        if ($process.HasExited) {
            throw "Scribli exited before the kernel API became ready. Exit code: $($process.ExitCode)"
        }
        Start-Sleep -Milliseconds 500
    }
}

throw "Timed out waiting for Scribli kernel API at $versionUri"
