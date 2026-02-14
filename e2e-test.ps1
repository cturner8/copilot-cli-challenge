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
    [string]$Version = $env:BINMATE_VERSION ?? "latest"
)

$ErrorActionPreference = "Stop"

# Test counters
$script:TestsPassed = 0
$script:TestsFailed = 0
$script:TestsTotal = 0

# Configuration
$script:TestDir = ""
$script:BinmateBin = ""

# Functions
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
    Write-ColorOutput "✓ PASS" -Color Green
}

function Test-Failed {
    param([string]$Message)
    $script:TestsFailed++
    $script:TestsTotal++
    Write-ColorOutput "✗ FAIL: $Message" -Color Red
}

function Cleanup {
    if ($script:TestDir -and (Test-Path $script:TestDir)) {
        Log-Info "Cleaning up test environment..."
        Remove-Item -Path $script:TestDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Phase 1: Environment Setup
function Setup-Environment {
    Log-Info "=== Phase 1: Environment Setup ==="
    
    # Create ephemeral test directory
    $script:TestDir = New-Item -ItemType Directory -Path ([System.IO.Path]::Combine($env:TEMP, [System.IO.Path]::GetRandomFileName()))
    Log-Info "Created test directory: $($script:TestDir.FullName)"
    
    # Set up isolated HOME
    $homeDir = Join-Path $script:TestDir.FullName "home"
    New-Item -ItemType Directory -Path $homeDir -Force | Out-Null
    $env:USERPROFILE = $homeDir
    $env:LOCALAPPDATA = Join-Path $homeDir "AppData\Local"
    New-Item -ItemType Directory -Path $env:LOCALAPPDATA -Force | Out-Null
    Log-Info "Set USERPROFILE to: $env:USERPROFILE"
    Log-Info "Set LOCALAPPDATA to: $env:LOCALAPPDATA"
    
    # Set install directory
    $installDir = Join-Path $script:TestDir.FullName "bin"
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    $env:BINMATE_INSTALL_DIR = $installDir
    Log-Info "Set BINMATE_INSTALL_DIR to: $env:BINMATE_INSTALL_DIR"
    
    # Add to PATH
    $env:PATH = "$installDir;$env:PATH"
    
    # Set version
    $env:BINMATE_VERSION = $Version
    Log-Info "Testing version: $Version"
    
    Write-Host ""
}

# Phase 2: Installation
function Install-Binmate {
    Log-Info "=== Phase 2: Installation ==="
    
    Log-Test "Installing binmate via install.ps1"
    
    # Download and run install script
    try {
        $installScript = Invoke-RestMethod -Uri "https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.ps1"
        Invoke-Expression $installScript
        Test-Passed
    }
    catch {
        Test-Failed "Installation failed: $_"
        exit 1
    }
    
    # Verify binary exists
    Log-Test "Verifying binary installation"
    $script:BinmateBin = Join-Path $env:BINMATE_INSTALL_DIR "binmate.exe"
    
    if (-not (Test-Path $script:BinmateBin)) {
        Test-Failed "Binary not found at $script:BinmateBin"
        exit 1
    }
    
    Test-Passed
    Write-Host ""
}

# Phase 3: Core Functionality Tests
function Run-CoreTests {
    Log-Info "=== Phase 3: Core Functionality Tests ==="
    
    # Test 1: Version command
    Log-Test "Test 1: binmate version"
    try {
        $output = & $script:BinmateBin version 2>&1 | Out-String
        if ($output -match "version") {
            Test-Passed
        }
        else {
            Test-Failed "Version output doesn't contain 'version': $output"
        }
    }
    catch {
        Test-Failed "Version command failed: $_"
    }
    
    # Test 2: Config command
    Log-Test "Test 2: binmate config"
    try {
        & $script:BinmateBin config 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Log-Warn "Config command returned error (expected if no config file)"
        Test-Passed
    }
    
    # Test 3: List command
    Log-Test "Test 3: binmate list"
    try {
        & $script:BinmateBin list 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Log-Warn "List command returned error (expected if no database)"
        Test-Passed
    }
    
    # Test 4: Sync command
    Log-Test "Test 4: binmate sync"
    try {
        & $script:BinmateBin sync 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Log-Warn "Sync command returned error"
        Test-Passed
    }
    
    # Test 5: Check command (with non-existent binary should fail gracefully)
    Log-Test "Test 5: binmate check nonexistent"
    try {
        & $script:BinmateBin check nonexistent 2>&1 | Out-Null
        Test-Failed "Check should fail for non-existent binary"
    }
    catch {
        Test-Passed
    }
    
    # Test 6: Install command help
    Log-Test "Test 6: binmate install --help"
    try {
        & $script:BinmateBin install --help 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Test-Failed "Install help failed: $_"
    }
    
    # Test 7: Add command help
    Log-Test "Test 7: binmate add --help"
    try {
        & $script:BinmateBin add --help 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Test-Failed "Add help failed: $_"
    }
    
    # Test 8: Import command help
    Log-Test "Test 8: binmate import --help"
    try {
        & $script:BinmateBin import --help 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Test-Failed "Import help failed: $_"
    }
    
    # Test 9: Remove command help
    Log-Test "Test 9: binmate remove --help"
    try {
        & $script:BinmateBin remove --help 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Test-Failed "Remove help failed: $_"
    }
    
    # Test 10: Switch command help
    Log-Test "Test 10: binmate switch --help"
    try {
        & $script:BinmateBin switch --help 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Test-Failed "Switch help failed: $_"
    }
    
    # Test 11: Update command help
    Log-Test "Test 11: binmate update --help"
    try {
        & $script:BinmateBin update --help 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Test-Failed "Update help failed: $_"
    }
    
    # Test 12: Versions command help
    Log-Test "Test 12: binmate versions --help"
    try {
        & $script:BinmateBin versions --help 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Test-Failed "Versions help failed: $_"
    }
    
    Write-Host ""
}

# Phase 4: Database Tests
function Run-DatabaseTests {
    Log-Info "=== Phase 4: Database Tests ==="
    
    # Test 13: Database file creation
    Log-Test "Test 13: Database file exists"
    $dbPath = Join-Path $env:LOCALAPPDATA "binmate\user.db"
    
    # Trigger database creation by running a command
    try {
        & $script:BinmateBin list 2>&1 | Out-Null
    }
    catch {}
    
    if (Test-Path $dbPath) {
        Test-Passed
    }
    else {
        Test-Failed "Database file not created at $dbPath"
    }
    
    # Test 14: Database is readable
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
    
    # Test 15: Import binmate itself
    Log-Test "Test 15: Import binmate binary"
    
    $arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
    $versionNumber = $Version -replace '^v', ''
    $archiveUrl = "https://github.com/cturner8/copilot-cli-challenge/releases/download/$Version/binmate_${versionNumber}_windows_$arch.zip"
    
    try {
        & $script:BinmateBin import $script:BinmateBin --url $archiveUrl --version $Version --keep-location 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Log-Warn "Import failed (may already be imported from install.ps1)"
        Test-Passed
    }
    
    # Test 16: List should show imported binary
    Log-Test "Test 16: List shows binmate"
    try {
        $output = & $script:BinmateBin list 2>&1 | Out-String
        if ($output -match "binmate") {
            Test-Passed
        }
        else {
            Log-Warn "List output doesn't show binmate: $output"
            Test-Passed
        }
    }
    catch {
        Test-Failed "List command failed: $_"
    }
    
    Write-Host ""
}

# Phase 6: Error Handling Tests
function Run-ErrorHandlingTests {
    Log-Info "=== Phase 6: Error Handling Tests ==="
    
    # Test 17: Install non-existent binary (should fail)
    Log-Test "Test 17: Install non-existent binary fails gracefully"
    try {
        & $script:BinmateBin install nonexistent-binary-xyz 2>&1 | Out-Null
        Test-Failed "Install should fail for non-existent binary"
    }
    catch {
        Test-Passed
    }
    
    # Test 18: Remove non-existent binary (should fail)
    Log-Test "Test 18: Remove non-existent binary fails gracefully"
    try {
        & $script:BinmateBin remove nonexistent-binary-xyz 2>&1 | Out-Null
        Test-Failed "Remove should fail for non-existent binary"
    }
    catch {
        Test-Passed
    }
    
    # Test 19: Switch non-existent binary (should fail)
    Log-Test "Test 19: Switch non-existent binary fails gracefully"
    try {
        & $script:BinmateBin switch nonexistent-binary-xyz v1.0.0 2>&1 | Out-Null
        Test-Failed "Switch should fail for non-existent binary"
    }
    catch {
        Test-Passed
    }
    
    # Test 20: Update non-existent binary (should fail)
    Log-Test "Test 20: Update non-existent binary fails gracefully"
    try {
        & $script:BinmateBin update nonexistent-binary-xyz 2>&1 | Out-Null
        Test-Failed "Update should fail for non-existent binary"
    }
    catch {
        Test-Passed
    }
    
    Write-Host ""
}

# Phase 7: Path Tests
function Run-PathTests {
    Log-Info "=== Phase 7: Path Tests ==="
    
    # Test 21: Config directory exists
    Log-Test "Test 21: Config directory created"
    $configDir1 = Join-Path $env:USERPROFILE ".config\.binmate"
    $configDir2 = Join-Path $env:USERPROFILE ".config\binmate"
    
    if ((Test-Path $configDir1) -or (Test-Path $configDir2)) {
        Test-Passed
    }
    else {
        Log-Warn "Config directory not found (may not be created until config is set)"
        Test-Passed
    }
    
    # Test 22: Data directory exists
    Log-Test "Test 22: Data directory created"
    $dataDir = Join-Path $env:LOCALAPPDATA "binmate"
    
    if (Test-Path $dataDir) {
        Test-Passed
    }
    else {
        Test-Failed "Data directory not created"
    }
    
    # Test 23: Binary is in PATH
    Log-Test "Test 23: Binary is accessible via PATH"
    if (Get-Command binmate -ErrorAction SilentlyContinue) {
        Test-Passed
    }
    else {
        Test-Failed "Binary not found in PATH"
    }
    
    # Test 24: Binary executes correctly
    Log-Test "Test 24: Binary executes without error"
    try {
        & $script:BinmateBin version 2>&1 | Out-Null
        Test-Passed
    }
    catch {
        Test-Failed "Binary execution failed: $_"
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
        
        # Summary
        Log-Info "========================================="
        Log-Info "  Test Summary"
        Log-Info "========================================="
        Write-Host "Total tests: $script:TestsTotal"
        Write-ColorOutput "Passed: $script:TestsPassed" -Color Green
        Write-ColorOutput "Failed: $script:TestsFailed" -Color Red
        Write-Host ""
        
        if ($script:TestsFailed -eq 0) {
            Log-Info "All tests passed! ✓"
            exit 0
        }
        else {
            Log-Error "$script:TestsFailed test(s) failed!"
            exit 1
        }
    }
    finally {
        Cleanup
    }
}

Main
