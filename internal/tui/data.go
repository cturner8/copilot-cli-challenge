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

// getBinariesWithMetadata fetches all binaries with their metadata using a single query with joins
func getBinariesWithMetadata(dbService *repository.Service) ([]BinaryWithMetadata, error) {
	// Use a single query with joins to get all data at once
	query := `
		SELECT 
			b.id, b.user_id, b.name, b.alias, b.provider, b.provider_path, b.install_path,
			b.format, b.asset_regex, b.release_regex, b.config_digest, b.created_at, b.updated_at, b.config_version,
			COALESCE(i.version, 'none') as active_version,
			COALESCE(install_count.count, 0) as install_count,
			i.id as installation_id, i.installed_path, i.source_url, i.file_size,
			i.checksum, i.checksum_algorithm, i.installed_at
		FROM binaries b
		LEFT JOIN versions v ON b.id = v.binary_id
		LEFT JOIN installations i ON v.installation_id = i.id
		LEFT JOIN (
			SELECT binary_id, COUNT(*) as count
			FROM installations
			GROUP BY binary_id
		) install_count ON b.id = install_count.binary_id
		ORDER BY b.name
	`

	rows, err := dbService.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query binaries with metadata: %w", err)
	}
	defer rows.Close()

	var result []BinaryWithMetadata
	for rows.Next() {
		binary := &database.Binary{}
		metadata := BinaryWithMetadata{
			Binary: binary,
		}

		var activeVersion string
		var installationID, installedAt *int64
		var installedPath, sourceURL, checksum, checksumAlgorithm *string
		var fileSize *int64

		err := rows.Scan(
			&binary.ID, &binary.UserID, &binary.Name, &binary.Alias,
			&binary.Provider, &binary.ProviderPath, &binary.InstallPath, &binary.Format,
			&binary.AssetRegex, &binary.ReleaseRegex, &binary.ConfigDigest, &binary.CreatedAt, &binary.UpdatedAt,
			&binary.ConfigVersion, &activeVersion, &metadata.InstallCount,
			&installationID, &installedPath, &sourceURL, &fileSize,
			&checksum, &checksumAlgorithm, &installedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan binary metadata: %w", err)
		}

		metadata.ActiveVersion = activeVersion

		// If we have active installation data, populate it
		if installationID != nil {
			metadata.ActiveInstallation = &database.Installation{
				ID:                *installationID,
				BinaryID:          binary.ID,
				Version:           activeVersion,
				InstalledPath:     *installedPath,
				SourceURL:         *sourceURL,
				FileSize:          *fileSize,
				Checksum:          *checksum,
				ChecksumAlgorithm: *checksumAlgorithm,
				InstalledAt:       *installedAt,
			}
		}

		result = append(result, metadata)
	}

	return result, rows.Err()
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
