<#
.SYNOPSIS
  Install mcp-server-sparx-ea from a GitHub release (Windows).

.DESCRIPTION
  irm https://raw.githubusercontent.com/InsonusK/mcp-server-sparx-ea/master/scripts/install.ps1 | iex

  With parameters, wrap it in a scriptblock:
  & ([scriptblock]::Create((irm https://raw.githubusercontent.com/InsonusK/mcp-server-sparx-ea/master/scripts/install.ps1))) -Version 0.5.0

  Detects amd64 / arm64, downloads the .zip, verifies it against SHA256SUMS,
  and installs mcp-server-sparx-ea.exe. Linux / macOS use scripts/install.sh.

.PARAMETER Version
  Release to install, e.g. 0.5.0. Default: the latest release. (env: VERSION)

.PARAMETER Dir
  Install directory. Default: %LOCALAPPDATA%\Programs\mcp-server-sparx-ea.
  (env: INSTALL_DIR)

.PARAMETER NoPath
  Do not add Dir to the user PATH.
#>
[CmdletBinding()]
param(
    [string]$Version = $env:VERSION,
    [string]$Dir = $env:INSTALL_DIR,
    [switch]$NoPath
)

$ErrorActionPreference = 'Stop'
try { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 } catch {}

$Repo = 'InsonusK/mcp-server-sparx-ea'
$Binary = 'mcp-server-sparx-ea'

if (-not $Dir) { $Dir = Join-Path $env:LOCALAPPDATA "Programs\$Binary" }

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { throw "unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

if (-not $Version) {
    $latest = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
    $Version = $latest.tag_name
    if (-not $Version) { throw 'could not resolve the latest release tag' }
}
$Version = $Version.TrimStart('v')

$archive = "${Binary}_v${Version}_windows_${arch}.zip"
$base = "https://github.com/$Repo/releases/download/v$Version"
$tmp = Join-Path ([IO.Path]::GetTempPath()) ("mcp-sparx-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    Write-Host "downloading $archive ..."
    $zip = Join-Path $tmp $archive
    try {
        Invoke-WebRequest "$base/$archive" -OutFile $zip -UseBasicParsing
    } catch {
        throw "download failed - check that release v$Version has an asset for windows/$arch"
    }

    try {
        $sums = (Invoke-WebRequest "$base/SHA256SUMS" -UseBasicParsing).Content
        $line = ($sums -split "`n" | Where-Object { $_ -match ([regex]::Escape($archive) + '\s*$') } | Select-Object -First 1)
        $want = if ($line) { ($line -split '\s+')[0] } else { $null }
        if ($want) {
            $got = (Get-FileHash $zip -Algorithm SHA256).Hash
            if ($got -ne $want.ToUpper()) { throw "checksum mismatch: want $want, got $got" }
            Write-Host "checksum ok"
        } else {
            Write-Warning "no checksum for $archive in SHA256SUMS - skipping"
        }
    } catch {
        if ("$_" -like '*checksum mismatch*') { throw }
        Write-Warning "checksum check skipped: $_"
    }

    Expand-Archive -Path $zip -DestinationPath $tmp -Force
    $src = Join-Path $tmp "$Binary.exe"
    if (-not (Test-Path $src)) { throw "archive did not contain $Binary.exe" }

    New-Item -ItemType Directory -Path $Dir -Force | Out-Null
    Copy-Item $src (Join-Path $Dir "$Binary.exe") -Force
    Write-Host "installed $Binary v$Version -> $Dir\$Binary.exe"

    if (-not $NoPath) {
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if (($userPath -split ';') -notcontains $Dir) {
            [Environment]::SetEnvironmentVariable('Path', ($userPath.TrimEnd(';') + ";$Dir"), 'User')
            Write-Host "added $Dir to your user PATH - restart the shell to pick it up"
        }
    }
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
