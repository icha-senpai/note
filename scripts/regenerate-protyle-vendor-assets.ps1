<#
.SYNOPSIS
Regenerates app-loaded Protyle vendor assets that can be copied from pinned npm package tarballs.

.DESCRIPTION
This script updates only the runtime assets that Scribli loads from app/stage/protyle/js and that have a direct npm tarball source. Custom bundles such as Lute, Flowchart, Highlight.js, third-languages.js, and mermaid/icons.json are documented in docs/GENERATED-ASSETS.md and intentionally skipped here.

.PARAMETER KeepScratch
Keep downloaded package tarballs and extraction output under .tools/protyle-vendor-assets for inspection.
#>

[CmdletBinding()]
param(
    [switch]$KeepScratch
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = (Resolve-Path (Join-Path $ScriptDir "..")).Path
$StageRoot = Join-Path $RepoRoot "app\stage\protyle\js"
$ScratchRoot = Join-Path $RepoRoot ".tools\protyle-vendor-assets"
$TarRoot = Join-Path $ScratchRoot "tarballs"
$ExtractRoot = Join-Path $ScratchRoot "packages"
$env:NPM_CONFIG_CACHE = Join-Path $RepoRoot ".tools\npm-cache"

function Assert-UnderRepo {
    param([string]$Path)

    $fullPath = [System.IO.Path]::GetFullPath($Path)
    $repoPrefix = $RepoRoot.TrimEnd("\") + "\"
    if (!$fullPath.StartsWith($repoPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Refusing to touch path outside repository: $fullPath"
    }
}

function Reset-Directory {
    param([string]$Path)

    Assert-UnderRepo $Path
    if (Test-Path -LiteralPath $Path) {
        Remove-Item -LiteralPath $Path -Recurse -Force
    }
    New-Item -ItemType Directory -Force -Path $Path | Out-Null
}

function Ensure-Directory {
    param([string]$Path)

    Assert-UnderRepo $Path
    New-Item -ItemType Directory -Force -Path $Path | Out-Null
}

function Get-Package {
    param([string]$Spec)

    $safeName = $Spec -replace "[^A-Za-z0-9_.-]", "_"
    $packageRoot = Join-Path $ExtractRoot $safeName
    Ensure-Directory $TarRoot
    Reset-Directory $packageRoot

    $packOutput = & npm pack $Spec --pack-destination $TarRoot --silent
    if ($LASTEXITCODE -ne 0) {
        throw "npm pack failed for $Spec"
    }

    $tarballName = $packOutput | Where-Object { $_ -like "*.tgz" } | Select-Object -Last 1
    if (!$tarballName) {
        throw "npm pack did not report a tarball for $Spec"
    }

    $tarballPath = Join-Path $TarRoot $tarballName
    if (!(Test-Path -LiteralPath $tarballPath)) {
        throw "Expected tarball was not created: $tarballPath"
    }

    & tar -xzf $tarballPath -C $packageRoot
    if ($LASTEXITCODE -ne 0) {
        throw "tar extraction failed for $tarballPath"
    }

    return Join-Path $packageRoot "package"
}

function Copy-PackageFile {
    param(
        [string]$PackageRoot,
        [string]$Source,
        [string]$Destination
    )

    $sourcePath = Join-Path $PackageRoot $Source
    $destinationPath = Join-Path $StageRoot $Destination
    if (!(Test-Path -LiteralPath $sourcePath)) {
        throw "Missing package file: $sourcePath"
    }
    Ensure-Directory (Split-Path -Parent $destinationPath)
    Copy-Item -LiteralPath $sourcePath -Destination $destinationPath -Force
}

function Copy-PackageDirectory {
    param(
        [string]$PackageRoot,
        [string]$Source,
        [string]$Destination
    )

    $sourcePath = Join-Path $PackageRoot $Source
    $destinationPath = Join-Path $StageRoot $Destination
    if (!(Test-Path -LiteralPath $sourcePath)) {
        throw "Missing package directory: $sourcePath"
    }
    Reset-Directory $destinationPath
    Get-ChildItem -LiteralPath $sourcePath -Force | ForEach-Object {
        Copy-Item -LiteralPath $_.FullName -Destination $destinationPath -Recurse -Force
    }
}

Reset-Directory $ScratchRoot
Ensure-Directory $env:NPM_CONFIG_CACHE

$abcjs = Get-Package "abcjs@6.5.0"
Copy-PackageFile $abcjs "dist\abcjs-basic-min.js" "abcjs\abcjs-basic-min.js"

$echarts = Get-Package "echarts@5.3.2"
Copy-PackageFile $echarts "dist\echarts.min.js" "echarts\echarts.min.js"

$echartsGl = Get-Package "echarts-gl@2.0.9"
Copy-PackageFile $echartsGl "dist\echarts-gl.min.js" "echarts\echarts-gl.min.js"

$viz = Get-Package "@viz-js/viz@3.11.0"
Copy-PackageFile $viz "lib\viz-standalone.js" "graphviz\viz.js"

$htmlToImage = Get-Package "html-to-image@1.11.13"
Copy-PackageFile $htmlToImage "dist\html-to-image.js" "html-to-image.min.js"

$katex = Get-Package "katex@0.16.9"
Copy-PackageFile $katex "dist\katex.min.js" "katex\katex.min.js"
Copy-PackageFile $katex "dist\katex.min.css" "katex\katex.min.css"
Copy-PackageFile $katex "dist\contrib\mhchem.min.js" "katex\mhchem.min.js"
Copy-PackageDirectory $katex "dist\fonts" "katex\fonts"

$mathjax = Get-Package "mathjax-full@3.1.2"
Copy-PackageFile $mathjax "es5\tex-svg-full.js" "mathjax\tex-svg-full.js"

$mermaid = Get-Package "mermaid@11.13.0"
Copy-PackageFile $mermaid "dist\mermaid.min.js" "mermaid\mermaid.min.js"

$zenuml = Get-Package "@mermaid-js/mermaid-zenuml@0.2.2"
Copy-PackageFile $zenuml "dist\mermaid-zenuml.min.js" "mermaid\mermaid-zenuml.min.js"
$zenumlDestination = Join-Path $StageRoot "mermaid\mermaid-zenuml.min.js"
$zenumlContent = Get-Content -LiteralPath $zenumlDestination -Raw
$zenumlContent = $zenumlContent -replace '^"use strict";', ""
if ($zenumlContent -notmatch 'globalThis\["mermaid-zenuml"\]\s*=') {
    $zenumlContent = $zenumlContent.TrimEnd() + "`r`nglobalThis[""mermaid-zenuml""] = globalThis.__esbuild_esm_mermaid_nm[""mermaid-zenuml""].default;"
}
if ($zenumlContent -notmatch 'window\.zenuml\s*=') {
    $zenumlContent = $zenumlContent.TrimEnd() + "`r`nwindow.zenuml = globalThis[""mermaid-zenuml""];`r`n"
}
Set-Content -LiteralPath $zenumlDestination -Value $zenumlContent -NoNewline

$pdfjs = Get-Package "pdfjs-dist@4.7.76"
Copy-PackageFile $pdfjs "legacy\build\pdf.min.mjs" "pdf\pdf.min.mjs"
Copy-PackageFile $pdfjs "legacy\build\pdf.worker.min.mjs" "pdf\pdf.worker.min.mjs"
Copy-PackageFile $pdfjs "legacy\build\pdf.sandbox.min.mjs" "pdf\pdf.sandbox.min.mjs"
Copy-PackageDirectory $pdfjs "cmaps" "pdf\cmaps"
Copy-PackageDirectory $pdfjs "standard_fonts" "pdf\standard_fonts"

$plantuml = Get-Package "plantuml-encoder@1.4.0"
Copy-PackageFile $plantuml "dist\plantuml-encoder.min.js" "plantuml\plantuml-encoder.min.js"

$viewerjs = Get-Package "viewerjs@1.11.7"
Copy-PackageFile $viewerjs "dist\viewer.min.js" "viewerjs\viewer.js"

$visNetwork = Get-Package "vis-network@9.1.13"
Copy-PackageFile $visNetwork "dist\vis-network.min.js" "vis\vis-network.min.js"

if (!$KeepScratch) {
    Reset-Directory $ScratchRoot
}

Write-Host "Regenerated direct-copy Protyle vendor assets."
Write-Host "Still manual/custom: lute.min.js, flowchart.min.js, highlight.min.js, third-languages.js, and mermaid/icons.json."
