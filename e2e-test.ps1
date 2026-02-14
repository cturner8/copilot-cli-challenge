#!/usr/bin/env pwsh
#
# End-to-End Test Script for binmate (Windows)
#
# This script tests binmate installation and core functionality in an ephemeral environment.
# It can be run locally or in CI/CD pipelines.
#
# Usage:
#   .\e2e-test.ps1 [-Version <version>]
#
# Parameters:
#   -Version - binmate version to test (default: "latest")
#
# Environment variables:
#   BINMATE_VERSION - Alternative way to specify version
#

param(
    [string]$Version = $(if ($env:BINMATE_VERSION) { $env:BINMATE_VERSION } else { "latest" })
)

$ErrorActionPreference = "Stop"

# Test counters
$script:TestsPassed = 0
$script:TestsFailed = 0
$script:TestsTotal = 0

# Configuration
$script:TestDir = ""
$script:BinmateBin = ""
$script:ManagedBinDir = ""
$script:ConfigPath = ""
$script:ScriptDir = Split-Path -Parent $PSCommandPath
$script:ResolvedVersion = ""
$script:ReleaseTag = ""
$script:BinmateArchiveUrl = ""
$script:ImportedBinaryId = "binmate"
$script:FzfLatestVersion = ""
$script:FzfPreviousVersion = ""
$script:GitHubToken = if ($env:GITHUB_TOKEN) { $env:GITHUB_TOKEN } elseif ($env:GH_TOKEN) { $env:GH_TOKEN } else { "" }

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Log-Info {
    param([string]$Message)
    Write-ColorOutput "==> $Message" -Color Green
}

function Log-Warn {
    param([string]$Message)
    Write-ColorOutput "Warning: $Message" -Color Yellow
}

function Log-Error {
    param([string]$Message)
    Write-ColorOutput "Error: $Message" -Color Red
}

function Log-Test {
    param([string]$Message)
    Write-ColorOutput "[TEST] $Message" -Color Blue
}

function Test-Passed {
    $script:TestsPassed++
    $script:TestsTotal++
    Write-ColorOutput "PASS" -Color Green
}

function Test-Failed {
    param([string]$Message)
    $script:TestsFailed++
    $script:TestsTotal++
    Write-ColorOutput "FAIL: $Message" -Color Red
}

function Cleanup {
    if ($script:TestDir -and (Test-Path $script:TestDir)) {
        Log-Info "Cleaning up test environment..."
        Remove-Item -Path $script:TestDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Invoke-NativeCommand {
    param(
        [Parameter(Mandatory = $true)]
        [string]$FilePath,
        [string[]]$Arguments = @()
    )

    $global:LASTEXITCODE = 0
    $output = & $FilePath @Arguments 2>&1 | Out-String
    $exitCode = if ($null -ne $LASTEXITCODE) { [int]$LASTEXITCODE } else { 0 }

    [PSCustomObject]@{
        Output   = $output.Trim()
        ExitCode = $exitCode
    }
}

function Get-GitHubHeaders {
    if ([string]::IsNullOrWhiteSpace($script:GitHubToken)) {
        return @{}
    }

    return @{
        Authorization         = "Bearer $($script:GitHubToken)"
        "X-GitHub-Api-Version" = "2022-11-28"
    }
}

function Invoke-GitHubRestMethod {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Uri
    )

    $headers = Get-GitHubHeaders
    if ($headers.Count -gt 0) {
        return Invoke-RestMethod -Uri $Uri -Headers $headers
    }

    return Invoke-RestMethod -Uri $Uri
}

function Invoke-GitHubWebRequest {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Uri,
        [Parameter(Mandatory = $true)]
        [string]$OutFile
    )

    $headers = Get-GitHubHeaders
    if ($headers.Count -gt 0) {
        Invoke-WebRequest -Uri $Uri -OutFile $OutFile -Headers $headers
        return
    }

    Invoke-WebRequest -Uri $Uri -OutFile $OutFile
}

function Download-TestConfig {
    $configDir = Join-Path $env:APPDATA ".binmate"
    $configFile = Join-Path $configDir "config.json"
    $localConfig = Join-Path $script:ScriptDir "config.json"
    $remoteConfig = "https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/config.json"

    New-Item -ItemType Directory -Path $configDir -Force | Out-Null

    if (Test-Path $localConfig) {
        Copy-Item -Path $localConfig -Destination $configFile -Force
    }
    else {
        Invoke-GitHubWebRequest -Uri $remoteConfig -OutFile $configFile
    }

    $script:ConfigPath = $configFile
    $env:BINMATE_CONFIG_PATH = $configFile
}

function Get-ReleasePlatform {
    $archRaw = if ($env:PROCESSOR_ARCHITECTURE) { $env:PROCESSOR_ARCHITECTURE.ToUpperInvariant() } else { "" }
    $arch = switch -Regex ($archRaw) {
        "ARM64" { "arm64"; break }
        "AMD64|X86_64" { "amd64"; break }
        default { throw "Unsupported architecture: $archRaw" }
    }

    return "windows_$arch"
}

function Get-FzfReleaseVersions {
    try {
        $releases = Invoke-GitHubRestMethod -Uri "https://api.github.com/repos/junegunn/fzf/releases?per_page=5"
        if (-not $releases -or $releases.Count -lt 2) {
            return $false
        }

        $script:FzfPreviousVersion = ""
        foreach ($release in $releases) {
            if ($release.tag_name -and $release.tag_name -ne $script:FzfLatestVersion) {
                $script:FzfPreviousVersion = $release.tag_name
                break
            }
        }

        if ([string]::IsNullOrWhiteSpace($script:FzfLatestVersion) -or [string]::IsNullOrWhiteSpace($script:FzfPreviousVersion)) {
            return $false
        }

        return $true
    }
    catch {
        return $false
    }
}

# Phase 1: Environment Setup
function Setup-Environment {
    Log-Info "=== Phase 1: Environment Setup ==="

    $script:TestDir = Join-Path $env:TEMP ("binmate-e2e-" + [System.IO.Path]::GetRandomFileName())
    New-Item -ItemType Directory -Path $script:TestDir -Force | Out-Null
    Log-Info "Created test directory: $script:TestDir"

    $homeDir = Join-Path $script:TestDir "home"
    $localAppData = Join-Path $homeDir "AppData\Local"
    $appData = Join-Path $homeDir "AppData\Roaming"
    $installDir = Join-Path $script:TestDir "bin"

    New-Item -ItemType Directory -Path $homeDir, $localAppData, $appData, $installDir -Force | Out-Null

    $env:USERPROFILE = $homeDir
    $env:LOCALAPPDATA = $localAppData
    $env:APPDATA = $appData
    $env:BINMATE_INSTALL_DIR = $installDir
    $env:BINMATE_SKIP_AUTO_IMPORT = "true"
    $env:BINMATE_VERSION = $Version

    $script:ManagedBinDir = Join-Path $env:LOCALAPPDATA "binmate\bin"
    New-Item -ItemType Directory -Path $script:ManagedBinDir -Force | Out-Null

    $env:PATH = "$installDir;$script:ManagedBinDir;$env:PATH"

    Log-Info "Set USERPROFILE to: $env:USERPROFILE"
    Log-Info "Set LOCALAPPDATA to: $env:LOCALAPPDATA"
    Log-Info "Set APPDATA to: $env:APPDATA"
    Log-Info "Set BINMATE_INSTALL_DIR to: $env:BINMATE_INSTALL_DIR"
    Log-Info "Testing version: $Version"
    if (-not [string]::IsNullOrWhiteSpace($script:GitHubToken)) {
        Log-Info "Using GitHub authentication header for API requests"
    }

    try {
        Download-TestConfig
        Log-Info "Config file prepared at: $script:ConfigPath"
    }
    catch {
        Test-Failed "Failed to prepare config.json in isolated profile: $_"
        exit 1
    }

    Write-Host ""
}

# Phase 2: Installation
function Install-Binmate {
    Log-Info "=== Phase 2: Installation ==="

    Log-Test "Installing binmate via install.ps1"

    $installScriptPath = Join-Path $script:ScriptDir "install.ps1"
    if (-not (Test-Path $installScriptPath)) {
        $installScriptPath = Join-Path $script:TestDir "install.ps1"
        try {
            Invoke-GitHubWebRequest -Uri "https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.ps1" -OutFile $installScriptPath
        }
        catch {
            Test-Failed "Failed to download install.ps1: $_"
            exit 1
        }
    }

    $installResult = Invoke-NativeCommand -FilePath "pwsh" -Arguments @("-NoProfile", "-ExecutionPolicy", "Bypass", "-File", $installScriptPath)
    if ($installResult.ExitCode -ne 0) {
        Test-Failed "Installation failed: $($installResult.Output)"
        exit 1
    }
    Test-Passed

    Log-Test "Verifying binary installation"
    $script:BinmateBin = Join-Path $env:BINMATE_INSTALL_DIR "binmate.exe"
    if (-not (Test-Path $script:BinmateBin)) {
        Test-Failed "Binary not found at $script:BinmateBin"
        exit 1
    }
    Test-Passed

    Log-Test "Resolving installed binmate version"
    $versionResult = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("version")
    if ($versionResult.ExitCode -ne 0) {
        Test-Failed "Failed to resolve installed version: $($versionResult.Output)"
        exit 1
    }

    if ($versionResult.Output -match 'binmate\s+([^\s]+)') {
        $script:ResolvedVersion = $Matches[1]
    }

    if ([string]::IsNullOrWhiteSpace($script:ResolvedVersion)) {
        Test-Failed "Unable to resolve installed version from output: $($versionResult.Output)"
        exit 1
    }

    $script:ReleaseTag = $script:ResolvedVersion
    if (-not $script:ReleaseTag.StartsWith("v")) {
        $script:ReleaseTag = "v$script:ReleaseTag"
    }

    try {
        $platform = Get-ReleasePlatform
    }
    catch {
        Test-Failed "$_"
        exit 1
    }

    Log-Test "Resolving release archive URL for $platform"
    try {
        $release = Invoke-GitHubRestMethod -Uri "https://api.github.com/repos/cturner8/copilot-cli-challenge/releases/tags/$($script:ReleaseTag)"
        $asset = $release.assets | Where-Object { $_.browser_download_url -match "_$platform\.zip$" } | Select-Object -First 1
        if (-not $asset) {
            Test-Failed "Could not resolve archive URL for platform $platform"
            exit 1
        }

        $script:BinmateArchiveUrl = $asset.browser_download_url
        $assetBaseName = [System.IO.Path]::GetFileNameWithoutExtension($asset.name)
        $script:ImportedBinaryId = ($assetBaseName -split '[-_]')[0]
        if ([string]::IsNullOrWhiteSpace($script:ImportedBinaryId)) {
            Test-Failed "Failed to derive imported binary ID from asset $($asset.name)"
            exit 1
        }
    }
    catch {
        Test-Failed "Failed to fetch release metadata for $($script:ReleaseTag): $_"
        exit 1
    }

    Test-Passed
    Write-Host ""
}

# Phase 3: Core Functionality Tests
function Run-CoreTests {
    Log-Info "=== Phase 3: Core Functionality Tests ==="

    Log-Test "Test 1: binmate version"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("version")
    if ($result.ExitCode -eq 0 -and $result.Output -match "binmate") {
        Test-Passed
    }
    else {
        Test-Failed "Version output doesn't contain 'binmate': $($result.Output)"
    }

    Log-Test "Test 2: binmate version --verbose"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("version", "--verbose")
    if ($result.ExitCode -eq 0 -and $result.Output -match "version:" -and $result.Output -match "commit:") {
        Test-Passed
    }
    else {
        Test-Failed "Verbose version output missing expected fields: $($result.Output)"
    }

    Log-Test "Test 3: binmate config"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("config")
    if ($result.ExitCode -eq 0 -and $result.Output -match "binmate Configuration") {
        Test-Passed
    }
    else {
        Test-Failed "Config output missing header: $($result.Output)"
    }

    Log-Test "Test 4: binmate config --json"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("config", "--json")
    if ($result.ExitCode -eq 0 -and ($result.Output -match '"Binaries"' -or $result.Output -match '"binaries"') -and $result.Output -match "{") {
        Test-Passed
    }
    else {
        Test-Failed "Config JSON output missing expected fields: $($result.Output)"
    }

    Log-Test "Test 5: binmate list (pre-sync)"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("list")
    if ($result.ExitCode -eq 0 -and $result.Output -match "No binaries installed") {
        Test-Passed
    }
    else {
        Test-Failed "Expected no binaries before sync: $($result.Output)"
    }

    Log-Test "Test 6: binmate sync"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("sync")
    if ($result.ExitCode -eq 0 -and $result.Output -match "Sync complete") {
        Test-Passed
    }
    else {
        Test-Failed "Sync output missing completion message: $($result.Output)"
    }

    Log-Test "Test 7: binmate list (post-sync)"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("list")
    if ($result.ExitCode -eq 0 -and $result.Output -match "(?i)fzf") {
        Test-Passed
    }
    else {
        Test-Failed "Expected fzf in list output after sync: $($result.Output)"
    }

    Log-Test "Test 8: binmate check --all"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("check", "--all")
    if ($result.ExitCode -eq 0) {
        Test-Passed
    }
    else {
        Test-Failed "Check --all command failed: $($result.Output)"
    }

    Log-Test "Test 9: binmate import (self)"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("import", $script:BinmateBin, "--url", $script:BinmateArchiveUrl, "--version", $script:ReleaseTag, "--keep-location")
    if ($result.ExitCode -eq 0) {
        Test-Passed
    }
    else {
        Test-Failed "Import command failed: $($result.Output)"
    }

    Log-Test "Test 10: binmate versions --binary $($script:ImportedBinaryId)"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("versions", "--binary", $script:ImportedBinaryId)
    if ($result.ExitCode -eq 0 -and $result.Output -match [regex]::Escape($script:ReleaseTag)) {
        Test-Passed
    }
    else {
        Test-Failed "Imported binary version list does not include $($script:ReleaseTag): $($result.Output)"
    }

    Log-Test "Test 11: binmate install --binary fzf --version latest"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("install", "--binary", "fzf", "--version", "latest")
    if ($result.ExitCode -eq 0) {
        Test-Passed
    }
    else {
        Test-Failed "Failed to install fzf latest: $($result.Output)"
    }

    Log-Test "Test 12: binmate versions --binary fzf"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("versions", "--binary", "fzf")
    if ($result.ExitCode -eq 0) {
        $activeLine = ($result.Output -split "`r?`n" | Where-Object { $_ -match '^\*\s+' } | Select-Object -First 1)
        if ($activeLine) {
            $parts = $activeLine -split '\s+'
            if ($parts.Count -ge 2 -and -not [string]::IsNullOrWhiteSpace($parts[1])) {
                $script:FzfLatestVersion = $parts[1]
                Test-Passed
            }
            else {
                Test-Failed "Unable to determine active fzf version: $($result.Output)"
            }
        }
        else {
            Test-Failed "Unable to determine active fzf version: $($result.Output)"
        }
    }
    else {
        Test-Failed "versions command failed for fzf: $($result.Output)"
    }

    Log-Test "Test 13: binmate list includes fzf"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("list")
    if ($result.ExitCode -eq 0 -and $result.Output -match "(?i)fzf") {
        Test-Passed
    }
    else {
        Test-Failed "fzf missing from list output: $($result.Output)"
    }

    Log-Test "Test 14: fzf --version"
    $fzfCommand = Get-Command fzf -ErrorAction SilentlyContinue
    if (-not $fzfCommand) {
        Test-Failed "fzf command is not executable from PATH"
    }
    else {
        $resolvedPath = $fzfCommand.Source
        $expectedPrefix = (Join-Path $env:LOCALAPPDATA "binmate\bin").ToLowerInvariant()
        if (-not $resolvedPath.ToLowerInvariant().StartsWith($expectedPrefix)) {
            Test-Failed "fzf resolved outside isolated profile: $resolvedPath"
        }
        else {
            $result = Invoke-NativeCommand -FilePath $resolvedPath -Arguments @("--version")
            if ($result.ExitCode -eq 0 -and -not [string]::IsNullOrWhiteSpace($result.Output)) {
                Test-Passed
            }
            else {
                Test-Failed "fzf --version failed: $($result.Output)"
            }
        }
    }

    Log-Test "Test 15: binmate update --binary fzf"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("update", "--binary", "fzf")
    if ($result.ExitCode -eq 0) {
        Test-Passed
    }
    else {
        Test-Failed "Update command failed for fzf: $($result.Output)"
    }

    Log-Test "Test 16: binmate update --binary $($script:ImportedBinaryId)"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("update", "--binary", $script:ImportedBinaryId)
    if ($result.ExitCode -eq 0) {
        Test-Passed
    }
    else {
        Test-Failed "Update command failed for imported binary ($($script:ImportedBinaryId)): $($result.Output)"
    }

    Log-Test "Test 17: Resolve latest and previous fzf releases"
    if (Get-FzfReleaseVersions) {
        if ($script:FzfLatestVersion -eq $script:FzfPreviousVersion) {
            Test-Failed "Unable to determine a distinct previous fzf version"
        }
        else {
            Test-Passed
        }
    }
    else {
        Test-Failed "Failed to fetch fzf release versions from GitHub API"
    }

    Log-Test "Test 18: binmate install --binary fzf --version previous"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("install", "--binary", "fzf", "--version", $script:FzfPreviousVersion)
    if ($result.ExitCode -eq 0) {
        Test-Passed
    }
    else {
        Test-Failed "Failed to install previous fzf version ($($script:FzfPreviousVersion)): $($result.Output)"
    }

    Log-Test "Test 19: binmate switch --binary fzf --version latest-installed"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("switch", "--binary", "fzf", "--version", $script:FzfLatestVersion)
    if ($result.ExitCode -eq 0) {
        $verify = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("versions", "--binary", "fzf")
        if ($verify.ExitCode -eq 0 -and $verify.Output -match "^\*\s+$([regex]::Escape($script:FzfLatestVersion))\b") {
            Test-Passed
        }
        else {
            Test-Failed "Active fzf version is not $($script:FzfLatestVersion): $($verify.Output)"
        }
    }
    else {
        Test-Failed "Failed to switch fzf to $($script:FzfLatestVersion): $($result.Output)"
    }

    Log-Test "Test 20: binmate switch --binary fzf --version previous"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("switch", "--binary", "fzf", "--version", $script:FzfPreviousVersion)
    if ($result.ExitCode -eq 0) {
        $verify = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("versions", "--binary", "fzf")
        if ($verify.ExitCode -eq 0 -and $verify.Output -match "^\*\s+$([regex]::Escape($script:FzfPreviousVersion))\b") {
            Test-Passed
        }
        else {
            Test-Failed "Active fzf version is not $($script:FzfPreviousVersion): $($verify.Output)"
        }
    }
    else {
        Test-Failed "Failed to switch fzf to $($script:FzfPreviousVersion): $($result.Output)"
    }

    Log-Test "Test 21: binmate switch --binary fzf --version latest-installed (restore)"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("switch", "--binary", "fzf", "--version", $script:FzfLatestVersion)
    if ($result.ExitCode -eq 0) {
        Test-Passed
    }
    else {
        Test-Failed "Failed to switch fzf back to $($script:FzfLatestVersion): $($result.Output)"
    }

    Log-Test "Test 22: binmate remove --binary fzf --files"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("remove", "--binary", "fzf", "--files")
    if ($result.ExitCode -eq 0) {
        Test-Passed
    }
    else {
        Test-Failed "Remove command failed for fzf: $($result.Output)"
    }

    Log-Test "Test 23: binmate list excludes fzf after removal"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("list")
    if ($result.ExitCode -eq 0) {
        if ($result.Output -match "(?i)fzf") {
            Test-Failed "fzf still present in list after removal: $($result.Output)"
        }
        else {
            Test-Passed
        }
    }
    else {
        Test-Failed "List command failed after removing fzf: $($result.Output)"
    }

    Log-Test "Test 24: fzf binary removed from PATH"
    $fzfPath = Join-Path $script:ManagedBinDir "fzf"
    $fzfExePath = Join-Path $script:ManagedBinDir "fzf.exe"
    if ((Test-Path $fzfPath) -or (Test-Path $fzfExePath)) {
        Test-Failed "fzf binary still exists in managed bin directory"
    }
    else {
        Test-Passed
    }

    Write-Host ""
}

# Phase 4: Database Tests
function Run-DatabaseTests {
    Log-Info "=== Phase 4: Database Tests ==="

    Log-Test "Test 13: Database file exists"
    $dbPath = Join-Path $env:LOCALAPPDATA "binmate\user.db"
    $null = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("list")

    if (Test-Path $dbPath -PathType Leaf) {
        Test-Passed
    }
    else {
        Test-Failed "Database file not created at $dbPath"
    }

    Log-Test "Test 14: Database is readable"
    if (Test-Path $dbPath -PathType Leaf) {
        Test-Passed
    }
    else {
        Test-Failed "Database file is not readable"
    }

    Write-Host ""
}

# Phase 5: Import Tests
function Run-ImportTests {
    Log-Info "=== Phase 5: Import Tests ==="

    Log-Test "Test 27: Re-import binmate binary (idempotency)"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("import", $script:BinmateBin, "--url", $script:BinmateArchiveUrl, "--version", $script:ReleaseTag, "--keep-location")
    if ($result.ExitCode -eq 0) {
        Test-Passed
    }
    else {
        Test-Failed "Re-import command failed: $($result.Output)"
    }

    Log-Test "Test 28: binmate list shows $($script:ImportedBinaryId)"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("list")
    if ($result.ExitCode -eq 0 -and $result.Output -match [regex]::Escape($script:ImportedBinaryId)) {
        Test-Passed
    }
    else {
        Test-Failed "List output doesn't show $($script:ImportedBinaryId): $($result.Output)"
    }

    Write-Host ""
}

# Phase 6: Error Handling Tests
function Run-ErrorHandlingTests {
    Log-Info "=== Phase 6: Error Handling Tests ==="

    Log-Test "Test 29: Install non-existent binary fails gracefully"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("install", "--binary", "nonexistent-binary-xyz")
    if ($result.ExitCode -ne 0) {
        Test-Passed
    }
    else {
        Test-Failed "Install should fail for non-existent binary"
    }

    Log-Test "Test 30: Remove non-existent binary fails gracefully"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("remove", "--binary", "nonexistent-binary-xyz")
    if ($result.ExitCode -ne 0) {
        Test-Passed
    }
    else {
        Test-Failed "Remove should fail for non-existent binary"
    }

    Log-Test "Test 31: Switch non-existent binary fails gracefully"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("switch", "--binary", "nonexistent-binary-xyz", "--version", "v1.0.0")
    if ($result.ExitCode -ne 0) {
        Test-Passed
    }
    else {
        Test-Failed "Switch should fail for non-existent binary"
    }

    Log-Test "Test 32: Update non-existent binary fails gracefully"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("update", "--binary", "nonexistent-binary-xyz")
    if ($result.ExitCode -ne 0) {
        Test-Passed
    }
    else {
        Test-Failed "Update should fail for non-existent binary"
    }

    Write-Host ""
}

# Phase 7: Path Tests
function Run-PathTests {
    Log-Info "=== Phase 7: Path Tests ==="

    Log-Test "Test 33: Isolated config.json exists"
    if ($script:ConfigPath -and (Test-Path $script:ConfigPath -PathType Leaf)) {
        Test-Passed
    }
    else {
        Test-Failed "Config file not found at $script:ConfigPath"
    }

    Log-Test "Test 34: Data directory created"
    $dataDir = Join-Path $env:LOCALAPPDATA "binmate"
    if (Test-Path $dataDir) {
        Test-Passed
    }
    else {
        Test-Failed "Data directory not created"
    }

    Log-Test "Test 35: Binary is accessible via PATH"
    if (Get-Command binmate -ErrorAction SilentlyContinue) {
        Test-Passed
    }
    else {
        Test-Failed "Binary not found in PATH"
    }

    Log-Test "Test 36: Binary executes without error"
    $result = Invoke-NativeCommand -FilePath $script:BinmateBin -Arguments @("version")
    if ($result.ExitCode -eq 0) {
        Test-Passed
    }
    else {
        Test-Failed "Binary execution failed: $($result.Output)"
    }

    Write-Host ""
}

# Main
function Main {
    try {
        Log-Info "========================================="
        Log-Info "  binmate End-to-End Test Suite (Windows)"
        Log-Info "========================================="
        Write-Host ""

        Setup-Environment
        Install-Binmate
        Run-CoreTests
        Run-DatabaseTests
        Run-ImportTests
        Run-ErrorHandlingTests
        Run-PathTests

        Log-Info "========================================="
        Log-Info "  Test Summary"
        Log-Info "========================================="
        Write-Host "Total tests: $script:TestsTotal"
        Write-ColorOutput "Passed: $script:TestsPassed" -Color Green
        Write-ColorOutput "Failed: $script:TestsFailed" -Color Red
        Write-Host ""

        if ($script:TestsFailed -eq 0) {
            Log-Info "All tests passed!"
            exit 0
        }

        Log-Error "$script:TestsFailed test(s) failed!"
        exit 1
    }
    finally {
        Cleanup
    }
}

Main
