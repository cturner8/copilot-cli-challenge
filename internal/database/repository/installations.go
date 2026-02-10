package repository

import (
	"database/sql"
	"fmt"
	"time"

	"cturner8/binmate/internal/database"
)

type InstallationsRepository struct {
	db *database.DB
}

func NewInstallationsRepository(db *database.DB) *InstallationsRepository {
	return &InstallationsRepository{db: db}
}

// Create inserts a new installation
func (r *InstallationsRepository) Create(installation *database.Installation) error {
	installation.InstalledAt = time.Now().Unix()

	result, err := r.db.Exec(`
INSERT INTO installations (binary_id, version, installed_path, source_url,
file_size, checksum, checksum_algorithm, installed_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, installation.BinaryID, installation.Version, installation.InstalledPath,
		installation.SourceURL, installation.FileSize, installation.Checksum,
		installation.ChecksumAlgorithm, installation.InstalledAt)

	if err != nil {
		return fmt.Errorf("failed to create installation: %w", err)
	}

	installation.ID, err = result.LastInsertId()
	return err
}

// Get retrieves an installation by binary ID and version
func (r *InstallationsRepository) Get(binaryID int64, version string) (*database.Installation, error) {
	installation := &database.Installation{}
	err := r.db.QueryRow(`
SELECT id, binary_id, version, installed_path, source_url, file_size,
checksum, checksum_algorithm, installed_at
FROM installations WHERE binary_id = ? AND version = ?
`, binaryID, version).Scan(&installation.ID, &installation.BinaryID, &installation.Version,
		&installation.InstalledPath, &installation.SourceURL, &installation.FileSize,
		&installation.Checksum, &installation.ChecksumAlgorithm, &installation.InstalledAt)

	if err == sql.ErrNoRows {
		return nil, database.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get installation: %w", err)
	}

	return installation, nil
}

// GetByID retrieves an installation by ID
func (r *InstallationsRepository) GetByID(id int64) (*database.Installation, error) {
	installation := &database.Installation{}
	err := r.db.QueryRow(`
SELECT id, binary_id, version, installed_path, source_url, file_size,
checksum, checksum_algorithm, installed_at
FROM installations WHERE id = ?
`, id).Scan(&installation.ID, &installation.BinaryID, &installation.Version,
		&installation.InstalledPath, &installation.SourceURL, &installation.FileSize,
		&installation.Checksum, &installation.ChecksumAlgorithm, &installation.InstalledAt)

	if err == sql.ErrNoRows {
		return nil, database.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get installation: %w", err)
	}

	return installation, nil
}

// ListByBinary retrieves all installations for a binary
func (r *InstallationsRepository) ListByBinary(binaryID int64) ([]*database.Installation, error) {
	rows, err := r.db.Query(`
SELECT id, binary_id, version, installed_path, source_url, file_size,
checksum, checksum_algorithm, installed_at
FROM installations 
WHERE binary_id = ?
ORDER BY installed_at DESC
`, binaryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list installations: %w", err)
	}
	defer rows.Close()

	var installations []*database.Installation
	for rows.Next() {
		installation := &database.Installation{}
		err := rows.Scan(&installation.ID, &installation.BinaryID, &installation.Version,
			&installation.InstalledPath, &installation.SourceURL, &installation.FileSize,
			&installation.Checksum, &installation.ChecksumAlgorithm, &installation.InstalledAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan installation: %w", err)
		}
		installations = append(installations, installation)
	}

	return installations, rows.Err()
}

// GetLatest retrieves the most recently installed version for a binary
func (r *InstallationsRepository) GetLatest(binaryID int64) (*database.Installation, error) {
	installation := &database.Installation{}
	err := r.db.QueryRow(`
SELECT id, binary_id, version, installed_path, source_url, file_size,
checksum, checksum_algorithm, installed_at
FROM installations 
WHERE binary_id = ?
ORDER BY installed_at DESC
LIMIT 1
`, binaryID).Scan(&installation.ID, &installation.BinaryID, &installation.Version,
		&installation.InstalledPath, &installation.SourceURL, &installation.FileSize,
		&installation.Checksum, &installation.ChecksumAlgorithm, &installation.InstalledAt)

	if err == sql.ErrNoRows {
		return nil, database.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest installation: %w", err)
	}

	return installation, nil
}

// Delete removes an installation
func (r *InstallationsRepository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM installations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete installation: %w", err)
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

// VerifyChecksum compares stored checksum with expected value
func (r *InstallationsRepository) VerifyChecksum(id int64, expectedChecksum string) (bool, error) {
	var exists int
	err := r.db.QueryRow(`
		SELECT 1 FROM installations 
		WHERE id = ? AND checksum = ?
	`, id, expectedChecksum).Scan(&exists)

	if err == sql.ErrNoRows {
		// Check if installation exists but checksum doesn't match
		var storedChecksum string
		checkErr := r.db.QueryRow(`
			SELECT checksum FROM installations WHERE id = ?
		`, id).Scan(&storedChecksum)

		if checkErr == sql.ErrNoRows {
			return false, database.ErrNotFound
		}
		if checkErr != nil {
			return false, fmt.Errorf("failed to verify checksum: %w", checkErr)
		}

		// Installation exists but checksum doesn't match
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to verify checksum: %w", err)
	}

	// Checksum matches
	return true, nil
}
