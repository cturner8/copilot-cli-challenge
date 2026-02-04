package repository

import (
	"database/sql"
	"fmt"
	"time"

	"cturner8/binmate/internal/database"
)

type DownloadsRepository struct {
	db *database.DB
}

func NewDownloadsRepository(db *database.DB) *DownloadsRepository {
	return &DownloadsRepository{db: db}
}

// Create inserts a new download record
func (r *DownloadsRepository) Create(download *database.Download) error {
	now := time.Now().Unix()
	download.DownloadedAt = now
	download.LastAccessedAt = now

	result, err := r.db.Exec(`
INSERT INTO downloads (binary_id, version, cache_path, source_url, file_size,
checksum, checksum_algorithm, downloaded_at, last_accessed_at, is_complete)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, download.BinaryID, download.Version, download.CachePath, download.SourceURL,
		download.FileSize, download.Checksum, download.ChecksumAlgorithm,
		download.DownloadedAt, download.LastAccessedAt, boolToInt(download.IsComplete))

	if err != nil {
		return fmt.Errorf("failed to create download: %w", err)
	}

	download.ID, err = result.LastInsertId()
	return err
}

// Get retrieves a download by binary ID and version
func (r *DownloadsRepository) Get(binaryID int64, version string) (*database.Download, error) {
	download := &database.Download{}
	var isComplete int

	err := r.db.QueryRow(`
SELECT id, binary_id, version, cache_path, source_url, file_size,
checksum, checksum_algorithm, downloaded_at, last_accessed_at, is_complete
FROM downloads WHERE binary_id = ? AND version = ?
`, binaryID, version).Scan(&download.ID, &download.BinaryID, &download.Version,
		&download.CachePath, &download.SourceURL, &download.FileSize, &download.Checksum,
		&download.ChecksumAlgorithm, &download.DownloadedAt, &download.LastAccessedAt, &isComplete)

	if err == sql.ErrNoRows {
		return nil, database.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get download: %w", err)
	}

	download.IsComplete = intToBool(isComplete)
	return download, nil
}

// UpdateLastAccessed updates the last accessed timestamp
func (r *DownloadsRepository) UpdateLastAccessed(id int64) error {
	now := time.Now().Unix()

	result, err := r.db.Exec(`
UPDATE downloads SET last_accessed_at = ? WHERE id = ?
`, now, id)

	if err != nil {
		return fmt.Errorf("failed to update last accessed: %w", err)
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

// MarkComplete marks a download as complete
func (r *DownloadsRepository) MarkComplete(id int64) error {
	result, err := r.db.Exec(`
UPDATE downloads SET is_complete = 1 WHERE id = ?
`, id)

	if err != nil {
		return fmt.Errorf("failed to mark download complete: %w", err)
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

// ListForCleanup retrieves downloads for LRU cleanup
func (r *DownloadsRepository) ListForCleanup(cutoffTime int64, limit int) ([]*database.Download, error) {
	rows, err := r.db.Query(`
SELECT id, binary_id, version, cache_path, source_url, file_size,
checksum, checksum_algorithm, downloaded_at, last_accessed_at, is_complete
FROM downloads
WHERE last_accessed_at < ?
ORDER BY last_accessed_at ASC
LIMIT ?
`, cutoffTime, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to list downloads for cleanup: %w", err)
	}
	defer rows.Close()

	return r.scanDownloads(rows)
}

// GetIncomplete retrieves incomplete downloads for resume
func (r *DownloadsRepository) GetIncomplete() ([]*database.Download, error) {
	rows, err := r.db.Query(`
SELECT id, binary_id, version, cache_path, source_url, file_size,
checksum, checksum_algorithm, downloaded_at, last_accessed_at, is_complete
FROM downloads
WHERE is_complete = 0
`)

	if err != nil {
		return nil, fmt.Errorf("failed to get incomplete downloads: %w", err)
	}
	defer rows.Close()

	return r.scanDownloads(rows)
}

// Delete removes a download record
func (r *DownloadsRepository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM downloads WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete download: %w", err)
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

// ListByBinary retrieves all downloads for a binary
func (r *DownloadsRepository) ListByBinary(binaryID int64) ([]*database.Download, error) {
	rows, err := r.db.Query(`
SELECT id, binary_id, version, cache_path, source_url, file_size,
checksum, checksum_algorithm, downloaded_at, last_accessed_at, is_complete
FROM downloads
WHERE binary_id = ?
ORDER BY downloaded_at DESC
`, binaryID)

	if err != nil {
		return nil, fmt.Errorf("failed to list downloads by binary: %w", err)
	}
	defer rows.Close()

	return r.scanDownloads(rows)
}

// scanDownloads is a helper to scan download rows
func (r *DownloadsRepository) scanDownloads(rows *sql.Rows) ([]*database.Download, error) {
	var downloads []*database.Download
	for rows.Next() {
		download := &database.Download{}
		var isComplete int

		err := rows.Scan(&download.ID, &download.BinaryID, &download.Version,
			&download.CachePath, &download.SourceURL, &download.FileSize,
			&download.Checksum, &download.ChecksumAlgorithm, &download.DownloadedAt,
			&download.LastAccessedAt, &isComplete)

		if err != nil {
			return nil, fmt.Errorf("failed to scan download: %w", err)
		}

		download.IsComplete = intToBool(isComplete)
		downloads = append(downloads, download)
	}

	return downloads, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}
