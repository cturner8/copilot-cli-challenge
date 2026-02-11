package binary

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"cturner8/binmate/internal/core/crypto"
	"cturner8/binmate/internal/core/url"
	v "cturner8/binmate/internal/core/version"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

// AddBinaryFromURL adds a binary by parsing a GitHub release URL
func AddBinaryFromURL(rawURL string, authenticated bool, dbService *repository.Service) (*database.Binary, error) {
	// Parse the GitHub release URL
	parsed, err := url.ParseGitHubReleaseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Generate binary ID and name
	binaryID := url.GenerateBinaryID(parsed.AssetName)
	binaryName := url.GenerateBinaryName(parsed.AssetName)

	// Create binary definition
	binary := &database.Binary{
		UserID:        binaryID,
		Name:          binaryName,
		Provider:      "github",
		ProviderPath:  fmt.Sprintf("%s/%s", parsed.Owner, parsed.Repo),
		Format:        parsed.Format,
		ConfigVersion: 0,        // TUI-added binaries have ConfigVersion=0
		Source:        "manual", // User-added binaries are marked as manual
		Authenticated: authenticated,
	}

	// Compute config digest
	binary.ConfigDigest = crypto.ComputeDigest(
		binary.UserID, binary.Name, "", binary.Provider,
		binary.ProviderPath, "", binary.Format, "", "",
	)

	// Check if binary already exists
	existing, err := dbService.Binaries.GetByUserID(binaryID)
	if err == nil {
		// Binary exists, return it
		log.Printf("Binary %s already exists", binaryID)
		return existing, nil
	} else if err != database.ErrNotFound {
		return nil, fmt.Errorf("failed to check existing binary: %w", err)
	}

	// Create the binary
	if err := dbService.Binaries.Create(binary); err != nil {
		return nil, fmt.Errorf("failed to create binary: %w", err)
	}

	log.Printf("Binary %s added successfully", binaryID)
	return binary, nil
}

// RemoveBinary removes a binary and all its installations
func RemoveBinary(binaryID string, dbService *repository.Service, removeFiles bool) error {
	// Get the binary
	binary, err := dbService.Binaries.GetByUserID(binaryID)
	if err != nil {
		return fmt.Errorf("binary not found: %w", err)
	}

	// Get all installations
	installations, err := dbService.Installations.ListByBinary(binary.ID)
	if err != nil {
		return fmt.Errorf("failed to list installations: %w", err)
	}

	// Delete physical files and symlinks if requested
	if removeFiles {
		// Get the active version to find the symlink path
		version, err := dbService.Versions.Get(binary.ID)
		if err == nil {
			// Symlink exists, try to remove it
			if version.SymlinkPath != "" {
				if err := os.Remove(version.SymlinkPath); err != nil {
					// Log warning but continue - symlink may have been manually deleted
					if !os.IsNotExist(err) {
						log.Printf("Warning: failed to remove symlink %s: %v", version.SymlinkPath, err)
					}
				} else {
					log.Printf("Removed symlink: %s", version.SymlinkPath)
				}
			}
		}

		// Delete all installation directories
		for _, inst := range installations {
			if inst.InstalledPath != "" {
				if err := os.RemoveAll(inst.InstalledPath); err != nil {
					// Log warning but continue - files may have been manually deleted
					if !os.IsNotExist(err) {
						log.Printf("Warning: failed to remove installation at %s: %v", inst.InstalledPath, err)
					}
				} else {
					log.Printf("Removed installation: %s", inst.InstalledPath)
				}
			}
		}
	}

	// Delete from database (cascade will handle installations, versions)
	if err := dbService.Binaries.Delete(binary.ID); err != nil {
		return fmt.Errorf("failed to delete binary: %w", err)
	}

	log.Printf("Binary %s removed successfully (%d installations)", binaryID, len(installations))
	return nil
}

// ListBinariesWithDetails retrieves all binaries with version information
func ListBinariesWithDetails(dbService *repository.Service) ([]*repository.BinaryWithVersionDetails, error) {
	return dbService.Binaries.ListWithVersionDetails("No active version")
}

// GetBinaryByID retrieves a binary by its user ID
func GetBinaryByID(binaryID string, dbService *repository.Service) (*database.Binary, error) {
	binary, err := dbService.Binaries.GetByUserID(binaryID)
	if err != nil {
		return nil, fmt.Errorf("binary not found: %w", err)
	}
	return binary, nil
}

// ImportBinary imports an existing binary from the file system
func ImportBinary(path string, name string, dbService *repository.Service) (*database.Binary, error) {
	return ImportBinaryWithOptions(path, name, "", false, dbService)
}

// ImportBinaryWithOptions imports an existing binary with additional options
func ImportBinaryWithOptions(path string, name string, version string, keepLocation bool, dbService *repository.Service) (*database.Binary, error) {
	// 1. Validate input
	if name == "" {
		return nil, fmt.Errorf("binary name cannot be empty")
	}

	// 2. Verify the file exists and is accessible
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to access file: %w", err)
	}

	// Check if it's a regular file
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("path is not a regular file: %s", path)
	}

	// Check if file is executable (on Unix-like systems)
	if fileInfo.Mode().Perm()&0111 == 0 {
		log.Printf("Warning: file %s is not marked as executable", path)
	}

	// 3. Compute checksum of the binary
	checksum, err := crypto.ComputeSHA256(path)
	if err != nil {
		return nil, fmt.Errorf("failed to compute checksum: %w", err)
	}

	// 4. Determine version
	if version == "" {
		// Default to timestamp-based version
		version = fmt.Sprintf("imported-%d", time.Now().Unix())
		log.Printf("No version provided, using: %s", version)
	}

	// 5. Generate binary ID
	binaryID := name

	// 6. Check if binary already exists
	existing, err := dbService.Binaries.GetByUserID(binaryID)
	if err == nil {
		// Binary exists, check if this version is already installed
		_, err := dbService.Installations.Get(existing.ID, version)
		if err == nil {
			log.Printf("Binary %s version %s already imported", binaryID, version)
			return existing, nil
		} else if err != database.ErrNotFound {
			return nil, fmt.Errorf("failed to check existing installation: %w", err)
		}
		// Version doesn't exist, will add it below
	} else if err != database.ErrNotFound {
		return nil, fmt.Errorf("failed to check existing binary: %w", err)
	}

	// 7. Determine installed path
	var installedPath string
	if keepLocation {
		// Use the original path
		installedPath = path
		log.Printf("Using original location: %s", installedPath)
	} else {
		// Copy to managed location
		destDir, err := getManagedPath(binaryID, version)
		if err != nil {
			return nil, fmt.Errorf("failed to determine managed path: %w", err)
		}

		// Create destination directory
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create destination directory: %w", err)
		}

		installedPath = filepath.Join(destDir, name)

		// Copy the binary
		if err := copyFile(path, installedPath); err != nil {
			return nil, fmt.Errorf("failed to copy binary: %w", err)
		}

		log.Printf("Copied binary to managed location: %s", installedPath)
	}

	// 8. Create or get binary record
	var binary *database.Binary
	if existing != nil {
		binary = existing
	} else {
		// Create new binary definition
		binary = &database.Binary{
			UserID:        binaryID,
			Name:          name,
			Provider:      "local",
			ProviderPath:  path, // Store original path
			Format:        "binary",
			ConfigVersion: 0,
			Source:        "manual",
			Authenticated: false,
		}

		// Compute config digest
		binary.ConfigDigest = crypto.ComputeDigest(
			binary.UserID, binary.Name, "", binary.Provider,
			binary.ProviderPath, "", binary.Format, "", "",
		)

		// Create the binary
		if err := dbService.Binaries.Create(binary); err != nil {
			return nil, fmt.Errorf("failed to create binary: %w", err)
		}

		log.Printf("Binary %s created successfully", binaryID)
	}

	// 9. Create installation record
	sourceURL := fmt.Sprintf("file://%s", path)
	installation := &database.Installation{
		Version:           version,
		InstalledPath:     installedPath,
		InstalledAt:       time.Now().Unix(),
		BinaryID:          binary.ID,
		SourceURL:         sourceURL,
		Checksum:          checksum,
		ChecksumAlgorithm: "SHA256",
		FileSize:          fileInfo.Size(),
	}

	if err := dbService.Installations.Create(installation); err != nil {
		return nil, fmt.Errorf("failed to save installation: %w", err)
	}

	// 10. Create symlink
	customInstallPath := ""
	if binary.InstallPath != nil {
		customInstallPath = *binary.InstallPath
	}

	symlinkPath, err := v.SetActiveVersion(installedPath, customInstallPath, binary.Name, binary.Alias)
	if err != nil {
		return nil, fmt.Errorf("failed to create symlink: %w", err)
	}

	// 11. Update active version
	if err := dbService.Versions.Set(binary.ID, installation.ID, symlinkPath); err != nil {
		return nil, fmt.Errorf("failed to save version: %w", err)
	}

	log.Printf("Successfully imported %s version %s", binaryID, version)
	return binary, nil
}

// getManagedPath returns the managed path for an imported binary
func getManagedPath(binaryID string, version string) (string, error) {
	baseDir, err := getLocalDataDir()
	if err != nil {
		return "", fmt.Errorf("unable to locate data directory: %w", err)
	}

	destDir := filepath.Join(baseDir, "binmate", "versions", binaryID, version)
	return destDir, nil
}

// getLocalDataDir returns the local data directory
func getLocalDataDir() (string, error) {
	if runtime.GOOS == "windows" {
		return os.UserCacheDir()
	}

	xdgDataHome, xdgDataHomeSet := os.LookupEnv("XDG_DATA_HOME")
	if xdgDataHomeSet {
		return xdgDataHome, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}

	return filepath.Join(homeDir, ".local", "share"), nil
}

// copyFile copies a file from src to dst, preserving permissions
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Get source file permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
