#!/usr/bin/env pwsh
#
# binmate installer for Windows
# 
# Usage:
#   irm https://raw.githubusercontent.com/cturner8/copilot-cli-challenge/main/install.ps1 | iex
#
# Environment variables:
#   BINMATE_VERSION     - Specific version to install (e.g., "v1.0.0" or "latest", default: "latest")
#   BINMATE_INSTALL_DIR - Installation directory (default: "$env:LOCALAPPDATA\binmate\bin")
#   BINMATE_SKIP_AUTO_IMPORT - Skip automatic post-install import (default: disabled)
#

$ErrorActionPreference = "Stop"

# Configuration
$GITHUB_REPO = "cturner8/copilot-cli-challenge"
$BINARY_NAME = "binmate"
$script:VERSION = if ($env:BINMATE_VERSION) { $env:BINMATE_VERSION } else { "latest" }
$INSTALL_DIR = if ($env:BINMATE_INSTALL_DIR) { $env:BINMATE_INSTALL_DIR } else { "$env:LOCALAPPDATA\binmate\bin" }
$SKIP_AUTO_IMPORT = $env:BINMATE_SKIP_AUTO_IMPORT

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

function Is-Truthy {
    param([string]$Value)
    return $Value -in @("1", "true", "TRUE", "yes", "YES", "on", "ON")
}

function Get-Platform {
    $arch = switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { "amd64" }
        "ARM64" { "arm64" }
        default {
            Log-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
            exit 1
        }
    }
    return "windows_$arch"
}

function Get-LatestVersion {
    Log-Info "Fetching latest version..."
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$GITHUB_REPO/releases/latest"
        return $response.tag_name
    }
    catch {
        Log-Error "Failed to fetch latest version: $_"
        exit 1
    }
}

function Test-VersionExists {
    param([string]$Version)
    
    try {
        $null = Invoke-RestMethod -Uri "https://api.github.com/repos/$GITHUB_REPO/releases/tags/$Version" -ErrorAction Stop
        return $true
    }
    catch {
        Log-Error "Version $Version not found in releases"
        exit 1
    }
}

function Install-Binmate {
    param(
        [string]$Version,
        [string]$Platform,
        [string]$DownloadUrl,
        [string]$ArchiveName
    )
    
    $checksumUrl = "https://github.com/$GITHUB_REPO/releases/download/$Version/checksums.txt"
    $tmpDir = New-Item -ItemType Directory -Path ([System.IO.Path]::Combine($env:TEMP, [System.IO.Path]::GetRandomFileName()))
    
    try {
        Log-Info "Downloading $BINARY_NAME $Version for $Platform..."
        
        # Download archive
        $archivePath = Join-Path $tmpDir.FullName $ArchiveName
        try {
            Invoke-WebRequest -Uri $DownloadUrl -OutFile $archivePath -ErrorAction Stop
        }
        catch {
            Log-Error "Failed to download $ArchiveName from $DownloadUrl"
            throw
        }
        
        # Download checksums
        Log-Info "Downloading checksums..."
        $checksumPath = Join-Path $tmpDir.FullName "checksums.txt"
        try {
            Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath -ErrorAction Stop
            
            # Verify checksum
            Log-Info "Verifying checksum..."
            $fileHash = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLower()
            $checksumContent = Get-Content $checksumPath
            $expectedHash = ($checksumContent | Select-String -Pattern $ArchiveName | ForEach-Object { $_.Line -split '\s+' } | Select-Object -First 1).ToLower()
            
            if ($fileHash -ne $expectedHash) {
                Log-Error "Checksum verification failed"
                Log-Error "Expected: $expectedHash"
                Log-Error "Got: $fileHash"
                exit 1
            }
        }
        catch {
            Log-Warn "Failed to download or verify checksums, skipping verification"
        }
        
        # Extract archive
        Log-Info "Extracting archive..."
        try {
            Expand-Archive -Path $archivePath -DestinationPath $tmpDir.FullName -Force
        }
        catch {
            Log-Error "Failed to extract archive"
            throw
        }
        
        # Install binary
        Log-Info "Installing to $INSTALL_DIR\$BINARY_NAME.exe..."
        
        # Create install directory if it doesn't exist
        if (-not (Test-Path $INSTALL_DIR)) {
            try {
                New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
            }
            catch {
                Log-Error "Failed to create directory $INSTALL_DIR"
                throw
            }
        }
        
        # Move binary to install directory
        $binarySource = Join-Path $tmpDir.FullName "$BINARY_NAME.exe"
        $binaryDest = Join-Path $INSTALL_DIR "$BINARY_NAME.exe"
        
        try {
            Copy-Item -Path $binarySource -Destination $binaryDest -Force
        }
        catch {
            Log-Error "Failed to install binary to $INSTALL_DIR"
            throw
        }
    }
    finally {
        # Cleanup
        Remove-Item -Path $tmpDir.FullName -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Invoke-AutoImport {
    param(
        [string]$InstalledPath,
        [string]$DownloadUrl,
        [string]$Version
    )
    
    if (Is-Truthy $SKIP_AUTO_IMPORT) {
        Log-Warn "Skipping automatic import because BINMATE_SKIP_AUTO_IMPORT is set"
        Log-Info "To import manually, run:"
        Write-Host "    & `"$InstalledPath`" import `"$InstalledPath`" --name `"$BINARY_NAME`" --url `"$DownloadUrl`" --version `"$Version`" --keep-location"
        return
    }
    
    Log-Info "Importing $BINARY_NAME for self-management..."
    try {
        & $InstalledPath import $InstalledPath --name $BINARY_NAME --url $DownloadUrl --version $Version --keep-location
        Log-Info "Automatic import completed"
    }
    catch {
        Log-Warn "Automatic import failed. To import manually, run:"
        Write-Host "    & `"$InstalledPath`" import `"$InstalledPath`" --name `"$BINARY_NAME`" --url `"$DownloadUrl`" --version `"$Version`" --keep-location"
    }
}

# Main
function Main {
    Log-Info "binmate installer"
    Write-Host ""
    
    # Detect platform
    $platform = Get-Platform
    Log-Info "Detected platform: $platform"
    
    # Determine version
    if ($VERSION -eq "latest") {
        $script:VERSION = Get-LatestVersion
        Log-Info "Latest version: $VERSION"
    }
    else {
        Log-Info "Installing version: $VERSION"
        Test-VersionExists $VERSION
    }
    
    # Strip 'v' prefix for archive name
    $versionNumber = $VERSION -replace '^v', ''
    $archiveName = "copilot-cli-challenge_${versionNumber}_${platform}.zip"
    $downloadUrl = "https://github.com/$GITHUB_REPO/releases/download/$VERSION/$archiveName"
    
    # Download and install
    Install-Binmate -Version $VERSION -Platform $platform -DownloadUrl $downloadUrl -ArchiveName $archiveName
    
    $installedPath = Join-Path $INSTALL_DIR "$BINARY_NAME.exe"
    Invoke-AutoImport -InstalledPath $installedPath -DownloadUrl $downloadUrl -Version $VERSION
    
    Write-Host ""
    Log-Info "Successfully installed $BINARY_NAME $VERSION to $installedPath"
    Write-Host ""
    Log-Info "Run '$BINARY_NAME --help' to get started"
    
    # Check if install directory is in PATH
    $pathDirs = $env:PATH -split ';'
    if ($INSTALL_DIR -notin $pathDirs) {
        Write-Host ""
        Log-Warn "$INSTALL_DIR is not in your PATH"
        Log-Warn "Add it to your PATH by running:"
        Write-Host "    `$env:PATH += `";$INSTALL_DIR`""
        Log-Warn "To make it permanent, add to your PowerShell profile:"
        Write-Host "    [Environment]::SetEnvironmentVariable('PATH', `$env:PATH + ';$INSTALL_DIR', 'User')"
    }
}

# Run main function
Main
