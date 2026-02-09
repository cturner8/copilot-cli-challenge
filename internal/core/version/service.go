package version

import (
	"fmt"
	"log"

	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

// SwitchVersion switches the active version of a binary
func SwitchVersion(binaryID string, version string, dbService *repository.Service) error {
	// Get the binary
	binary, err := dbService.Binaries.GetByUserID(binaryID)
	if err != nil {
		return fmt.Errorf("binary not found: %w", err)
	}

	// Get the installation for this version
	installation, err := dbService.Installations.Get(binary.ID, version)
	if err != nil {
		return fmt.Errorf("version %s not installed: %w", version, err)
	}

	// Handle optional InstallPath
	customInstallPath := ""
	if binary.InstallPath != nil {
		customInstallPath = *binary.InstallPath
	}

	// Update the symlink
	symlinkPath, err := SetActiveVersion(installation.InstalledPath, customInstallPath, binary.Name, binary.Alias)
	if err != nil {
		return fmt.Errorf("failed to set active version: %w", err)
	}

	// Update the versions table
	if err := dbService.Versions.Set(binary.ID, installation.ID, symlinkPath); err != nil {
		return fmt.Errorf("failed to update version record: %w", err)
	}

	log.Printf("Switched %s to version %s", binaryID, version)
	return nil
}

// GetActiveVersion retrieves the currently active version for a binary
func GetActiveVersion(binaryID string, dbService *repository.Service) (*database.Version, error) {
	// Get the binary
	binary, err := dbService.Binaries.GetByUserID(binaryID)
	if err != nil {
		return nil, fmt.Errorf("binary not found: %w", err)
	}

	// Get the active version
	version, err := dbService.Versions.Get(binary.ID)
	if err != nil {
		return nil, fmt.Errorf("no active version found: %w", err)
	}

	return version, nil
}

// ListVersions retrieves all installed versions for a binary
func ListVersions(binaryID string, dbService *repository.Service) ([]*database.Installation, error) {
	// Get the binary
	binary, err := dbService.Binaries.GetByUserID(binaryID)
	if err != nil {
		return nil, fmt.Errorf("binary not found: %w", err)
	}

	// Get all installations
	installations, err := dbService.Installations.ListByBinary(binary.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	return installations, nil
}
