<#
.SYNOPSIS
Regenerates the browser Lute bundle used by Scribli's editor.

.DESCRIPTION
Lute's browser bundle is produced with GopherJS v1.21.0, which is tied to the
Go 1.21 toolchain. Scribli's main kernel can use the Go version declared in
kernel/go.mod; this script keeps the older Go version isolated under .tools
only for regenerating app/stage/protyle/js/lute/lute.min.js.

.PARAMETER KeepScratch
Keep the temporary GOPATH source mirror and scratch output under .tools/lute-js-build.

.PARAMETER SkipToolchainDownload
Fail instead of downloading Go 1.21.13 if the repo-local toolchain is missing.
#>

[CmdletBinding()]
param(
    [switch]$KeepScratch,
    [switch]$SkipToolchainDownload
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = (Resolve-Path (Join-Path $ScriptDir "..")).Path
$ToolsRoot = Join-Path $RepoRoot ".tools"
$ScratchRoot = Join-Path $ToolsRoot "lute-js-build"
$GoVersion = "1.21.13"
$GoRoot = Join-Path $ToolsRoot "go$GoVersion"
$GoZip = Join-Path $ToolsRoot "go$GoVersion.windows-amd64.zip"
$GoUrl = "https://go.dev/dl/go$GoVersion.windows-amd64.zip"
$GoExe = Join-Path $GoRoot "bin\go.exe"
$Gopath = Join-Path $ScratchRoot "gopath"
$Gobin = Join-Path $ScratchRoot "bin"
$GopherJSVersion = "v1.21.0"
$GopherJSExe = Join-Path $Gobin "gopherjs.exe"
$LocalLuteRoot = Join-Path $RepoRoot "third_party\forks\lute"
$GopathLuteRoot = Join-Path $Gopath "src\github.com\icha-senpai\note\third_party\forks\lute"
$GopathGopherJSRoot = Join-Path $Gopath "src\github.com\gopherjs\gopherjs"
$ScratchOutput = Join-Path $ScratchRoot "lute.min.js"
$BundleOutput = Join-Path $RepoRoot "app\stage\protyle\js\lute\lute.min.js"

function Assert-UnderRepo {
    param([string]$Path)

    $fullPath = [System.IO.Path]::GetFullPath($Path)
    $repoPrefix = $RepoRoot.TrimEnd("\") + "\"
    if (!$fullPath.StartsWith($repoPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Refusing to touch path outside repository: $fullPath"
    }
}

function Ensure-Directory {
    param([string]$Path)

    Assert-UnderRepo $Path
    New-Item -ItemType Directory -Force -Path $Path | Out-Null
}

function Reset-Directory {
    param([string]$Path)

    Assert-UnderRepo $Path
    if (Test-Path -LiteralPath $Path) {
        Remove-Item -LiteralPath $Path -Recurse -Force
    }
    New-Item -ItemType Directory -Force -Path $Path | Out-Null
}

function Clear-ReadOnly {
    param([string]$Path)

    if (Test-Path -LiteralPath $Path) {
        Get-ChildItem -LiteralPath $Path -Recurse -Force -File | ForEach-Object {
            $_.IsReadOnly = $false
        }
    }
}

function Invoke-Checked {
    param(
        [string]$FilePath,
        [string[]]$Arguments
    )

    & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed: $FilePath $($Arguments -join ' ')"
    }
}

Ensure-Directory $ToolsRoot

if (!(Test-Path -LiteralPath $GoExe)) {
    if ($SkipToolchainDownload) {
        throw "Missing $GoExe. Re-run without -SkipToolchainDownload to fetch Go $GoVersion from $GoUrl."
    }

    Write-Host "Downloading Go $GoVersion from $GoUrl..."
    Invoke-Checked "curl.exe" @("-L", $GoUrl, "-o", $GoZip)

    $ExtractRoot = Join-Path $ScratchRoot "go-extract"
    Reset-Directory $ExtractRoot
    if (Test-Path -LiteralPath $GoRoot) {
        Assert-UnderRepo $GoRoot
        Remove-Item -LiteralPath $GoRoot -Recurse -Force
    }
    Expand-Archive -LiteralPath $GoZip -DestinationPath $ExtractRoot -Force
    Move-Item -LiteralPath (Join-Path $ExtractRoot "go") -Destination $GoRoot
    Remove-Item -LiteralPath $ExtractRoot -Recurse -Force
}

$env:GOROOT = $GoRoot
$env:GOTOOLCHAIN = "local"
$env:GOENV = "off"
$env:GOTELEMETRY = "off"
$env:GOTELEMETRYDIR = Join-Path $ToolsRoot "go-telemetry"
$env:GOCACHE = Join-Path $ToolsRoot "go-build-cache-1.21"
$env:GOMODCACHE = Join-Path $ToolsRoot "go-mod-cache-1.21"
$env:GOPATH = $Gopath
$env:GOBIN = $Gobin
$env:APPDATA = Join-Path $ToolsRoot "appdata"
$env:PATH = (Join-Path $GoRoot "bin") + ";" + $env:PATH

Ensure-Directory $env:GOCACHE
Ensure-Directory $env:GOMODCACHE
Ensure-Directory $env:GOTELEMETRYDIR
Ensure-Directory $env:APPDATA
Reset-Directory $ScratchRoot
Ensure-Directory $Gopath
Ensure-Directory $Gobin

$goVersionOutput = & $GoExe version
if ($goVersionOutput -notmatch "go$([regex]::Escape($GoVersion))") {
    throw "Expected Go $GoVersion, got: $goVersionOutput"
}

Write-Host "Installing GopherJS $GopherJSVersion with $goVersionOutput..."
$env:GO111MODULE = "on"
Invoke-Checked $GoExe @("install", "github.com/gopherjs/gopherjs@$GopherJSVersion")

if (!(Test-Path -LiteralPath $GopherJSExe)) {
    throw "GopherJS was not installed at $GopherJSExe"
}

$GopherJSModuleSource = Join-Path $env:GOMODCACHE "github.com\gopherjs\gopherjs@$GopherJSVersion"
if (!(Test-Path -LiteralPath $GopherJSModuleSource)) {
    throw "Missing GopherJS module source in cache: $GopherJSModuleSource"
}

Ensure-Directory (Split-Path -Parent $GopathGopherJSRoot)
Copy-Item -LiteralPath $GopherJSModuleSource -Destination $GopathGopherJSRoot -Recurse -Force
Clear-ReadOnly $GopathGopherJSRoot

Ensure-Directory (Split-Path -Parent $GopathLuteRoot)
Copy-Item -LiteralPath $LocalLuteRoot -Destination $GopathLuteRoot -Recurse -Force
Clear-ReadOnly $GopathLuteRoot

Write-Host "Building Lute browser bundle..."
$env:GO111MODULE = "off"
Invoke-Checked $GopherJSExe @(
    "build",
    "--tags",
    "javascript",
    "-m",
    "-o",
    $ScratchOutput,
    "github.com/icha-senpai/note/third_party/forks/lute/javascript"
)

if (!(Test-Path -LiteralPath $ScratchOutput)) {
    throw "GopherJS completed but did not create $ScratchOutput"
}

$bundleSize = (Get-Item -LiteralPath $ScratchOutput).Length
if ($bundleSize -lt 1000000) {
    throw "Generated bundle is suspiciously small: $bundleSize bytes"
}

Ensure-Directory (Split-Path -Parent $BundleOutput)
Copy-Item -LiteralPath $ScratchOutput -Destination $BundleOutput -Force

if (!$KeepScratch) {
    Reset-Directory $ScratchRoot
}

Write-Host "Regenerated app\stage\protyle\js\lute\lute.min.js ($bundleSize bytes)."
