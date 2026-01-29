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

// ReadAndSync reads config and syncs to database
func ReadAndSync(dbService *repository.Service) (Config, error) {
	config := ReadConfig()

	if err := SyncToDatabase(config, dbService); err != nil {
		return config, fmt.Errorf("failed to sync to database: %w", err)
	}

	return config, nil
}
