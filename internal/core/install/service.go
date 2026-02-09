package install

import (
	"fmt"
	"log"
	"os"
	"time"

	"cturner8/binmate/internal/core/crypto"
	v "cturner8/binmate/internal/core/version"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
	"cturner8/binmate/internal/providers/github"
)

// InstallBinaryResult contains the results of a binary installation
type InstallBinaryResult struct {
	Binary       *database.Binary
	Installation *database.Installation
	Version      string
}

// InstallBinary installs a specific version of a binary
func InstallBinary(binaryID string, version string, dbService *repository.Service) (*InstallBinaryResult, error) {
	// Get the binary configuration
	binaryConfig, err := dbService.Binaries.GetByUserID(binaryID)
	if err != nil {
		return nil, fmt.Errorf("binary not found: %w", err)
	}

	if binaryConfig.Provider != "github" {
		return nil, fmt.Errorf("only github provider is currently supported")
	}

	// Fetch release and asset information
	release, asset, err := github.FetchReleaseAsset(binaryConfig, version)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	// Download the asset
	downloadPath, err := github.DownloadAsset(asset.BrowserDownloadUrl, asset.Name, binaryConfig.Authenticated)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}

	// Verify downloaded archive checksum if digest is provided
	if asset.Digest != "" {
		if err := crypto.VerifyDigest(downloadPath, asset.Digest); err != nil {
			return nil, fmt.Errorf("checksum verification failed: %w", err)
		}
		log.Printf("✓ archive checksum verified")
	}

	// Resolve version (convert "latest" to actual tag name)
	resolvedVersion := version
	if version == "latest" {
		resolvedVersion = release.TagName
	}

	// Check if this version is already installed
	existingInstallation, err := dbService.Installations.Get(binaryConfig.ID, resolvedVersion)
	if err == nil {
		log.Printf("Version %s already installed", resolvedVersion)
		return &InstallBinaryResult{
			Binary:       binaryConfig,
			Installation: existingInstallation,
			Version:      resolvedVersion,
		}, nil
	} else if err != database.ErrNotFound {
		return nil, fmt.Errorf("failed to check existing installation: %w", err)
	}

	// Extract the asset
	destPath, err := ExtractAsset(downloadPath, binaryConfig, resolvedVersion)
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	// Compute checksum of extracted binary
	binaryChecksum, err := crypto.ComputeSHA256(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute binary checksum: %w", err)
	}

	// Get file size of extracted binary
	fileInfo, err := os.Stat(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Handle optional InstallPath
	customInstallPath := ""
	if binaryConfig.InstallPath != nil {
		customInstallPath = *binaryConfig.InstallPath
	}

	// Set active version (create symlink)
	symlinkPath, err := v.SetActiveVersion(destPath, customInstallPath, binaryConfig.Name, binaryConfig.Alias)
	if err != nil {
		return nil, fmt.Errorf("failed to set active version: %w", err)
	}

	// Create installation record
	installation := &database.Installation{
		Version:           resolvedVersion,
		InstalledPath:     destPath,
		InstalledAt:       time.Now().Unix(),
		BinaryID:          binaryConfig.ID,
		SourceURL:         asset.BrowserDownloadUrl,
		Checksum:          binaryChecksum,
		ChecksumAlgorithm: "SHA256",
		FileSize:          fileInfo.Size(),
	}

	if err := dbService.Installations.Create(installation); err != nil {
		return nil, fmt.Errorf("failed to save installation: %w", err)
	}

	// Update active version
	if err := dbService.Versions.Set(binaryConfig.ID, installation.ID, symlinkPath); err != nil {
		return nil, fmt.Errorf("failed to save version: %w", err)
	}

	log.Printf("Successfully installed %s version %s", binaryID, resolvedVersion)

	return &InstallBinaryResult{
		Binary:       binaryConfig,
		Installation: installation,
		Version:      resolvedVersion,
	}, nil
}

// UpdateToLatest updates a binary to the latest available version
func UpdateToLatest(binaryID string, dbService *repository.Service) (*InstallBinaryResult, error) {
	return InstallBinary(binaryID, "latest", dbService)
}
