package database

import (
	"path/filepath"
	"testing"
)

func TestSchemaStructure(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test foreign key enforcement
	t.Run("ForeignKeyEnforcement", func(t *testing.T) {
		// Try to insert installation without binary (should fail)
		_, err := db.Exec(`
INSERT INTO installations (binary_id, version, installed_path, source_url, file_size, checksum, installed_at)
VALUES (999, '1.0.0', '/path', 'http://example.com', 1000, 'checksum', 1234567890)
`)
		if err == nil {
			t.Error("Expected foreign key constraint error, got nil")
		}
	})

	// Test unique constraint on installations
	t.Run("UniqueInstallation", func(t *testing.T) {
		// Insert binary
		result, err := db.Exec(`
INSERT INTO binaries (user_id, name, provider, provider_path, format, created_at, updated_at)
VALUES ('test', 'test-binary', 'github', 'owner/repo', '.tar.gz', 1234567890, 1234567890)
`)
		if err != nil {
			t.Fatalf("Failed to insert binary: %v", err)
		}
		binaryID, _ := result.LastInsertId()

		// Insert first installation
		_, err = db.Exec(`
INSERT INTO installations (binary_id, version, installed_path, source_url, file_size, checksum, installed_at)
VALUES (?, '1.0.0', '/path1', 'http://example.com', 1000, 'checksum1', 1234567890)
`, binaryID)
		if err != nil {
			t.Fatalf("Failed to insert first installation: %v", err)
		}

		// Try to insert duplicate version (should fail)
		_, err = db.Exec(`
INSERT INTO installations (binary_id, version, installed_path, source_url, file_size, checksum, installed_at)
VALUES (?, '1.0.0', '/path2', 'http://example.com', 1000, 'checksum2', 1234567890)
`, binaryID)
		if err == nil {
			t.Error("Expected unique constraint error, got nil")
		}
	})

	// Test cascade delete
	t.Run("CascadeDelete", func(t *testing.T) {
		// Insert binary
		result, err := db.Exec(`
INSERT INTO binaries (user_id, name, provider, provider_path, format, created_at, updated_at)
VALUES ('cascade-test', 'cascade-binary', 'github', 'owner/repo', '.tar.gz', 1234567890, 1234567890)
`)
		if err != nil {
			t.Fatalf("Failed to insert binary: %v", err)
		}
		binaryID, _ := result.LastInsertId()

		// Insert installation
		_, err = db.Exec(`
INSERT INTO installations (binary_id, version, installed_path, source_url, file_size, checksum, installed_at)
VALUES (?, '1.0.0', '/path', 'http://example.com', 1000, 'checksum', 1234567890)
`, binaryID)
		if err != nil {
			t.Fatalf("Failed to insert installation: %v", err)
		}

		// Delete binary
		_, err = db.Exec(`DELETE FROM binaries WHERE id = ?`, binaryID)
		if err != nil {
			t.Fatalf("Failed to delete binary: %v", err)
		}

		// Verify installation was cascade deleted
		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM installations WHERE binary_id = ?`, binaryID).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count installations: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected 0 installations after cascade delete, got %d", count)
		}
	})
}
