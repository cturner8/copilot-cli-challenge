package repository

import (
	"database/sql"
	"fmt"
	"time"

	"cturner8/binmate/internal/database"
)

type BinariesRepository struct {
	db *database.DB
}

func NewBinariesRepository(db *database.DB) *BinariesRepository {
	return &BinariesRepository{db: db}
}

// Create inserts a new binary
func (r *BinariesRepository) Create(binary *database.Binary) error {
	now := time.Now().Unix()
	binary.CreatedAt = now
	binary.UpdatedAt = now

	result, err := r.db.Exec(`
INSERT INTO binaries (user_id, name, alias, provider, provider_path, install_path, 
format, asset_regex, release_regex, created_at, updated_at, config_version)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, binary.UserID, binary.Name, binary.Alias, binary.Provider, binary.ProviderPath,
		binary.InstallPath, binary.Format, binary.AssetRegex, binary.ReleaseRegex,
		binary.CreatedAt, binary.UpdatedAt, binary.ConfigVersion)

	if err != nil {
		return fmt.Errorf("failed to create binary: %w", err)
	}

	binary.ID, err = result.LastInsertId()
	return err
}

// Update updates an existing binary
func (r *BinariesRepository) Update(binary *database.Binary) error {
	binary.UpdatedAt = time.Now().Unix()

	result, err := r.db.Exec(`
UPDATE binaries 
SET user_id = ?, name = ?, alias = ?, provider = ?, provider_path = ?,
install_path = ?, format = ?, asset_regex = ?, release_regex = ?,
updated_at = ?, config_version = ?
WHERE id = ?
`, binary.UserID, binary.Name, binary.Alias, binary.Provider, binary.ProviderPath,
		binary.InstallPath, binary.Format, binary.AssetRegex, binary.ReleaseRegex,
		binary.UpdatedAt, binary.ConfigVersion, binary.ID)

	if err != nil {
		return fmt.Errorf("failed to update binary: %w", err)
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

// Get retrieves a binary by ID
func (r *BinariesRepository) Get(id int64) (*database.Binary, error) {
	binary := &database.Binary{}
	err := r.db.QueryRow(`
SELECT id, user_id, name, alias, provider, provider_path, install_path,
format, asset_regex, release_regex, created_at, updated_at, config_version
FROM binaries WHERE id = ?
`, id).Scan(&binary.ID, &binary.UserID, &binary.Name, &binary.Alias, &binary.Provider,
		&binary.ProviderPath, &binary.InstallPath, &binary.Format, &binary.AssetRegex,
		&binary.ReleaseRegex, &binary.CreatedAt, &binary.UpdatedAt, &binary.ConfigVersion)

	if err == sql.ErrNoRows {
		return nil, database.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get binary: %w", err)
	}

	return binary, nil
}

// GetByUserID retrieves a binary by user_id
func (r *BinariesRepository) GetByUserID(userID string) (*database.Binary, error) {
	binary := &database.Binary{}
	err := r.db.QueryRow(`
SELECT id, user_id, name, alias, provider, provider_path, install_path,
format, asset_regex, release_regex, created_at, updated_at, config_version
FROM binaries WHERE user_id = ?
`, userID).Scan(&binary.ID, &binary.UserID, &binary.Name, &binary.Alias, &binary.Provider,
		&binary.ProviderPath, &binary.InstallPath, &binary.Format, &binary.AssetRegex,
		&binary.ReleaseRegex, &binary.CreatedAt, &binary.UpdatedAt, &binary.ConfigVersion)

	if err == sql.ErrNoRows {
		return nil, database.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get binary by user_id: %w", err)
	}

	return binary, nil
}

// GetByName retrieves a binary by name
func (r *BinariesRepository) GetByName(name string) (*database.Binary, error) {
	binary := &database.Binary{}
	err := r.db.QueryRow(`
SELECT id, user_id, name, alias, provider, provider_path, install_path,
format, asset_regex, release_regex, created_at, updated_at, config_version
FROM binaries WHERE name = ?
`, name).Scan(&binary.ID, &binary.UserID, &binary.Name, &binary.Alias, &binary.Provider,
		&binary.ProviderPath, &binary.InstallPath, &binary.Format, &binary.AssetRegex,
		&binary.ReleaseRegex, &binary.CreatedAt, &binary.UpdatedAt, &binary.ConfigVersion)

	if err == sql.ErrNoRows {
		return nil, database.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get binary by name: %w", err)
	}

	return binary, nil
}

// List retrieves all binaries
func (r *BinariesRepository) List() ([]*database.Binary, error) {
	rows, err := r.db.Query(`
SELECT id, user_id, name, alias, provider, provider_path, install_path,
format, asset_regex, release_regex, created_at, updated_at, config_version
FROM binaries ORDER BY name
`)
	if err != nil {
		return nil, fmt.Errorf("failed to list binaries: %w", err)
	}
	defer rows.Close()

	var binaries []*database.Binary
	for rows.Next() {
		binary := &database.Binary{}
		err := rows.Scan(&binary.ID, &binary.UserID, &binary.Name, &binary.Alias,
			&binary.Provider, &binary.ProviderPath, &binary.InstallPath, &binary.Format,
			&binary.AssetRegex, &binary.ReleaseRegex, &binary.CreatedAt, &binary.UpdatedAt,
			&binary.ConfigVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to scan binary: %w", err)
		}
		binaries = append(binaries, binary)
	}

	return binaries, rows.Err()
}

// Delete removes a binary (cascade deletes installations, versions, downloads)
func (r *BinariesRepository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM binaries WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete binary: %w", err)
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

// SyncFromConfig syncs binaries from config file to database
func (r *BinariesRepository) SyncFromConfig(configBinaries []ConfigBinary, configVersion int) error {
	// Get existing binaries from database
	existing, err := r.List()
	if err != nil {
		return fmt.Errorf("failed to list existing binaries: %w", err)
	}

	// Create map of existing binaries by user_id
	existingMap := make(map[string]*database.Binary)
	for _, b := range existing {
		existingMap[b.UserID] = b
	}

	// Track which user_ids are in config
	configUserIDs := make(map[string]bool)

	// Upsert binaries from config
	for _, cb := range configBinaries {
		configUserIDs[cb.ID] = true

		if existingBinary, exists := existingMap[cb.ID]; exists {
			// Update existing binary
			existingBinary.Name = cb.Name
			existingBinary.Alias = stringToPtr(cb.Alias)
			existingBinary.Provider = cb.Provider
			existingBinary.ProviderPath = cb.Path
			existingBinary.InstallPath = stringToPtr(cb.InstallPath)
			existingBinary.Format = cb.Format
			existingBinary.AssetRegex = stringToPtr(cb.AssetRegex)
			existingBinary.ReleaseRegex = stringToPtr(cb.ReleaseRegex)
			existingBinary.ConfigVersion = configVersion

			if err := r.Update(existingBinary); err != nil {
				return fmt.Errorf("failed to update binary %s: %w", cb.ID, err)
			}
		} else {
			// Create new binary
			binary := &database.Binary{
				UserID:        cb.ID,
				Name:          cb.Name,
				Alias:         stringToPtr(cb.Alias),
				Provider:      cb.Provider,
				ProviderPath:  cb.Path,
				InstallPath:   stringToPtr(cb.InstallPath),
				Format:        cb.Format,
				AssetRegex:    stringToPtr(cb.AssetRegex),
				ReleaseRegex:  stringToPtr(cb.ReleaseRegex),
				ConfigVersion: configVersion,
			}

			if err := r.Create(binary); err != nil {
				return fmt.Errorf("failed to create binary %s: %w", cb.ID, err)
			}
		}
	}

	// Remove binaries that are no longer in config
	for userID := range existingMap {
		if !configUserIDs[userID] {
			if err := r.Delete(existingMap[userID].ID); err != nil {
				return fmt.Errorf("failed to delete binary %s: %w", userID, err)
			}
		}
	}

	return nil
}

func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
