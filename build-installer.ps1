#requires -Version 5.1
<#
.SYNOPSIS
  Builda checker.exe + syntaxchecker-mcp.exe e produce dist\setup.exe via Inno Setup.

.EXAMPLE
  .\build-installer.ps1
  .\build-installer.ps1 -Version 1.2.3
  .\build-installer.ps1 -Iscc "C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
#>
[CmdletBinding()]
param(
    [string]$Version,
    [string]$Iscc
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

if (-not $Version) {
    try   { $Version = (& git describe --tags --always 2>$null).Trim() }
    catch { $Version = '' }
    if (-not $Version) { $Version = 'dev' }
}

if (-not $Iscc) {
    $cmd = Get-Command iscc.exe -ErrorAction SilentlyContinue
    if ($cmd) {
        $Iscc = $cmd.Source
    } else {
        $candidates = @(
            "$env:ProgramFiles\Inno Setup 6\ISCC.exe",
            "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
            "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe"
        )
        $Iscc = $candidates | Where-Object { Test-Path $_ } | Select-Object -First 1
    }
}
if (-not $Iscc -or -not (Test-Path $Iscc)) {
    throw "ISCC.exe non trovato. Installa Inno Setup 6 o passa -Iscc <path>."
}

Write-Host "==> Version : $Version"
Write-Host "==> ISCC    : $Iscc"

$env:GOOS        = 'windows'
$env:GOARCH      = 'amd64'
$env:CGO_ENABLED = '0'

$dist = Join-Path $root 'dist'
New-Item -ItemType Directory -Force -Path $dist | Out-Null

$buildDate = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
try   { $commit = (& git rev-parse --short HEAD 2>$null).Trim() } catch { $commit = 'none' }
if (-not $commit) { $commit = 'none' }

$ldCheck = "-s -w -X main.version=$Version -X main.commit=$commit -X main.buildDate=$buildDate"
$ldMcp   = "-s -w"

Write-Host "==> build checker.exe"
Push-Location (Join-Path $root 'apps\checker')
& go build -trimpath -ldflags $ldCheck -o (Join-Path $dist 'checker.exe') .
if ($LASTEXITCODE -ne 0) { Pop-Location; throw "go build checker fallita" }
Pop-Location

Write-Host "==> build syntaxchecker-mcp.exe"
Push-Location (Join-Path $root 'apps\mcp-server')
& go build -trimpath -ldflags $ldMcp -o (Join-Path $dist 'syntaxchecker-mcp.exe') .
if ($LASTEXITCODE -ne 0) { Pop-Location; throw "go build mcp-server fallita" }
Pop-Location

Write-Host "==> compile installer.iss"
& $Iscc "/DMyAppVersion=$Version" (Join-Path $root 'installer.iss')
if ($LASTEXITCODE -ne 0) { throw "iscc fallito" }

Write-Host ""
Write-Host "OK -> $(Join-Path $dist 'setup.exe')"
