package repository

import (
	"database/sql"
	"fmt"
	"time"

	"cturner8/binmate/internal/database"
)

type VersionsRepository struct {
	db *database.DB
}

func NewVersionsRepository(db *database.DB) *VersionsRepository {
	return &VersionsRepository{db: db}
}

// Set activates a specific version for a binary
func (r *VersionsRepository) Set(binaryID, installationID int64, symlinkPath string) error {
	now := time.Now().Unix()

	_, err := r.db.Exec(`
INSERT INTO versions (binary_id, installation_id, activated_at, symlink_path)
VALUES (?, ?, ?, ?)
ON CONFLICT(binary_id) DO UPDATE SET
installation_id = excluded.installation_id,
activated_at = excluded.activated_at,
symlink_path = excluded.symlink_path
`, binaryID, installationID, now, symlinkPath)

	if err != nil {
		return fmt.Errorf("failed to set active version: %w", err)
	}

	return nil
}

// Get retrieves the active version for a binary
func (r *VersionsRepository) Get(binaryID int64) (*database.Version, error) {
	version := &database.Version{}
	err := r.db.QueryRow(`
SELECT binary_id, installation_id, activated_at, symlink_path
FROM versions WHERE binary_id = ?
`, binaryID).Scan(&version.BinaryID, &version.InstallationID,
		&version.ActivatedAt, &version.SymlinkPath)

	if err == sql.ErrNoRows {
		return nil, database.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}

	return version, nil
}

// Unset removes the active version for a binary
func (r *VersionsRepository) Unset(binaryID int64) error {
	result, err := r.db.Exec(`DELETE FROM versions WHERE binary_id = ?`, binaryID)
	if err != nil {
		return fmt.Errorf("failed to unset active version: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return database.ErrNotFound
	}

	return nil
}

// Switch changes the active version to a different installation
func (r *VersionsRepository) Switch(binaryID, installationID int64, symlinkPath string) error {
	return r.Set(binaryID, installationID, symlinkPath)
}

// GetWithInstallation retrieves active version with installation details
func (r *VersionsRepository) GetWithInstallation(binaryID int64) (*database.Version, *database.Installation, error) {
	version := &database.Version{}
	installation := &database.Installation{}

	err := r.db.QueryRow(`
SELECT v.binary_id, v.installation_id, v.activated_at, v.symlink_path,
i.id, i.binary_id, i.version, i.installed_path, i.source_url, i.file_size,
i.checksum, i.checksum_algorithm, i.installed_at
FROM versions v
JOIN installations i ON v.installation_id = i.id
WHERE v.binary_id = ?
`, binaryID).Scan(&version.BinaryID, &version.InstallationID, &version.ActivatedAt,
		&version.SymlinkPath, &installation.ID, &installation.BinaryID, &installation.Version,
		&installation.InstalledPath, &installation.SourceURL, &installation.FileSize,
		&installation.Checksum, &installation.ChecksumAlgorithm, &installation.InstalledAt)

	if err == sql.ErrNoRows {
		return nil, nil, database.ErrNotFound
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get active version with installation: %w", err)
	}

	return version, installation, nil
}
