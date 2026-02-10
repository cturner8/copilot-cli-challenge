package config

import (
	"fmt"

	"cturner8/binmate/internal/database/repository"
)

// SyncToDatabase syncs config binaries to the database
func SyncToDatabase(config Config, dbService *repository.Service) error {
	// Convert config binaries to repository format
	configBinaries := make([]repository.ConfigBinary, len(config.Binaries))
	for i, b := range config.Binaries {
		// Merge global config with binary-specific config
		merged := MergeBinaryWithGlobal(b, config.Global)

		configBinaries[i] = repository.ConfigBinary{
			ID:            merged.Id,
			Name:          merged.Name,
			Alias:         merged.Alias,
			Provider:      merged.Provider,
			Path:          merged.Path,
			InstallPath:   merged.InstallPath,
			Format:        merged.Format,
			AssetRegex:    merged.AssetRegex,
			ReleaseRegex:  merged.ReleaseRegex,
			Authenticated: merged.Authenticated,
		}
	}

	// Sync to database
	if err := dbService.Binaries.SyncFromConfig(configBinaries, config.Version); err != nil {
		return fmt.Errorf("failed to sync config to database: %w", err)
	}

	return nil
}

// SyncBinary syncs a specific binary from config to database
func SyncBinary(binaryID string, config Config, dbService *repository.Service) error {
	// Find the binary in config
	binary, err := GetBinary(binaryID, config.Binaries)
	if err != nil {
		return fmt.Errorf("binary not found in config: %w", err)
	}

	// Merge global config with binary-specific config
	merged := MergeBinaryWithGlobal(binary, config.Global)

	// Convert to repository format
	configBinary := repository.ConfigBinary{
		ID:            merged.Id,
		Name:          merged.Name,
		Alias:         merged.Alias,
		Provider:      merged.Provider,
		Path:          merged.Path,
		InstallPath:   merged.InstallPath,
		Format:        merged.Format,
		AssetRegex:    merged.AssetRegex,
		ReleaseRegex:  merged.ReleaseRegex,
		Authenticated: merged.Authenticated,
	}

	// Sync single binary to database
	if err := dbService.Binaries.SyncBinary(configBinary, config.Version); err != nil {
		return fmt.Errorf("failed to sync binary to database: %w", err)
	}

	return nil
}

// ReadAndSync reads config and syncs to database
func ReadAndSync(config Config, dbService *repository.Service) (Config, error) {
	if err := SyncToDatabase(config, dbService); err != nil {
		return config, fmt.Errorf("failed to sync to database: %w", err)
	}

	return config, nil
}
