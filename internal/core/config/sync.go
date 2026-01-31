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
		configBinaries[i] = repository.ConfigBinary{
			ID:           b.Id,
			Name:         b.Name,
			Alias:        b.Alias,
			Provider:     b.Provider,
			Path:         b.Path,
			InstallPath:  b.InstallPath,
			Format:       b.Format,
			AssetRegex:   b.AssetRegex,
			ReleaseRegex: b.ReleaseRegex,
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

	// Convert to repository format
	configBinary := repository.ConfigBinary{
		ID:           binary.Id,
		Name:         binary.Name,
		Alias:        binary.Alias,
		Provider:     binary.Provider,
		Path:         binary.Path,
		InstallPath:  binary.InstallPath,
		Format:       binary.Format,
		AssetRegex:   binary.AssetRegex,
		ReleaseRegex: binary.ReleaseRegex,
	}

	// Sync single binary to database
	if err := dbService.Binaries.SyncBinary(configBinary, config.Version); err != nil {
		return fmt.Errorf("failed to sync binary to database: %w", err)
	}

	return nil
}

// ReadAndSync reads config and syncs to database
func ReadAndSync(dbService *repository.Service) (Config, error) {
	config := ReadConfig()

	if err := SyncToDatabase(config, dbService); err != nil {
		return config, fmt.Errorf("failed to sync to database: %w", err)
	}

	return config, nil
}
