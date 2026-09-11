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

.PARAMETER Register
  Register with Claude Code: "project" writes .\.mcp.json in the current directory
  (commit it to share with the team), "user" adds it to your Claude Code user
  config, "no" skips. Default: ask when interactive, otherwise "no". (env: REGISTER)

.PARAMETER NoRegister
  Never register and do not prompt.
#>
[CmdletBinding()]
param(
    [string]$Version = $env:VERSION,
    [string]$Dir = $env:INSTALL_DIR,
    [switch]$NoPath,
    [ValidateSet('project', 'user', 'no', '')]
    [string]$Register = "$env:REGISTER",
    [switch]$NoRegister
)

$ErrorActionPreference = 'Stop'
try { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 } catch {}

$Repo = 'InsonusK/mcp-server-sparx-ea'
$Binary = 'mcp-server-sparx-ea'
$McpName = if ($env:MCP_NAME) { $env:MCP_NAME } else { 'sparx-ea' }
if ($NoRegister) { $Register = 'no' }

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

    $onPath = $false
    if (-not $NoPath) {
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if (($userPath -split ';') -notcontains $Dir) {
            [Environment]::SetEnvironmentVariable('Path', ($userPath.TrimEnd(';') + ";$Dir"), 'User')
            Write-Host "added $Dir to your user PATH - restart the shell to pick it up"
        }
        $onPath = $true
    }

    # --- register with Claude Code ------------------------------------------
    # "project" writes .\.mcp.json (commit it to share); "user" adds it to your
    # personal config. See docs/setup-with-an-agent.md.
    $mcpCmd = if ($onPath) { $Binary } else { Join-Path $Dir "$Binary.exe" }

    if (-not $Register) {
        $ans = Read-Host "Register `"$McpName`" with Claude Code? [p]roject (.\.mcp.json), [u]ser, [N]o"
        $Register = switch -Regex ($ans) {
            '^(p|project)$' { 'project' }
            '^(u|user)$'    { 'user' }
            default         { 'no' }
        }
    }

    if ($Register -eq 'project') {
        $claude = Get-Command claude -ErrorAction SilentlyContinue
        if ($claude) {
            & claude mcp add --scope project $McpName -- $mcpCmd
            if ($LASTEXITCODE -eq 0) {
                Write-Host "registered `"$McpName`" in .\.mcp.json (commit it to share)"
            } else {
                Write-Warning "'claude mcp add' failed - see docs/setup-with-an-agent.md"
            }
        } else {
            $file = Join-Path (Get-Location) '.mcp.json'
            $data = if (Test-Path $file) {
                try { Get-Content $file -Raw | ConvertFrom-Json } catch { [pscustomobject]@{} }
            } else { [pscustomobject]@{} }
            if (-not $data.PSObject.Properties['mcpServers']) {
                $data | Add-Member -NotePropertyName mcpServers -NotePropertyValue ([pscustomobject]@{})
            }
            $entry = [pscustomobject]@{ command = $mcpCmd; args = @() }
            if ($data.mcpServers.PSObject.Properties[$McpName]) {
                $data.mcpServers.$McpName = $entry
            } else {
                $data.mcpServers | Add-Member -NotePropertyName $McpName -NotePropertyValue $entry
            }
            $data | ConvertTo-Json -Depth 10 | Set-Content $file
            Write-Host "wrote $file - commit it so the team gets `"$McpName`""
        }
    } elseif ($Register -eq 'user') {
        $claude = Get-Command claude -ErrorAction SilentlyContinue
        if ($claude) {
            & claude mcp add --scope user $McpName -- $mcpCmd
            if ($LASTEXITCODE -eq 0) {
                Write-Host "registered `"$McpName`" in your Claude Code user config"
            } else {
                Write-Warning "'claude mcp add' failed - see docs/setup-with-an-agent.md"
            }
        } else {
            Write-Host "note: the 'claude' CLI is not on PATH - install it, then run:"
            Write-Host "  claude mcp add --scope user $McpName -- $mcpCmd"
        }
    }
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
