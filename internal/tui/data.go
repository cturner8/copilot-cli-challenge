package tui

import (
	"fmt"

	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

// BinaryWithMetadata represents a binary with additional metadata
type BinaryWithMetadata struct {
	Binary         *database.Binary
	ActiveVersion  string
	InstallCount   int
	ActiveInstallation *database.Installation
}

// getBinariesWithMetadata fetches all binaries with their metadata
func getBinariesWithMetadata(dbService *repository.Service) ([]BinaryWithMetadata, error) {
	binaries, err := dbService.Binaries.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list binaries: %w", err)
	}

	result := make([]BinaryWithMetadata, 0, len(binaries))
	for _, binary := range binaries {
		metadata := BinaryWithMetadata{
			Binary:       binary,
			ActiveVersion: "none",
			InstallCount:  0,
		}

		// Get installation count
		installations, err := dbService.Installations.ListByBinary(binary.ID)
		if err == nil {
			metadata.InstallCount = len(installations)
		}

		// Get active version
		_, installation, err := dbService.Versions.GetWithInstallation(binary.ID)
		if err == nil && installation != nil {
			metadata.ActiveVersion = installation.Version
			metadata.ActiveInstallation = installation
		}

		result = append(result, metadata)
	}

	return result, nil
}

// getVersionsForBinary fetches all installations for a binary ordered by date
func getVersionsForBinary(dbService *repository.Service, binaryID int64) ([]*database.Installation, error) {
	installations, err := dbService.Installations.ListByBinary(binaryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list installations: %w", err)
	}
	return installations, nil
}

// getActiveVersion gets the currently active version for a binary
func getActiveVersion(dbService *repository.Service, binaryID int64) (*database.Installation, error) {
	_, installation, err := dbService.Versions.GetWithInstallation(binaryID)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}
	return installation, nil
}
