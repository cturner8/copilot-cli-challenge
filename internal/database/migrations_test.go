package database

import (
	"path/filepath"
	"testing"
)

func TestMigrate(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Verify migrations table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&tableName)
	if err != nil {
		t.Errorf("Migrations table not found: %v", err)
	}

	// Verify version is recorded
	var version int
	err = db.QueryRow("SELECT version FROM migrations WHERE version = 1").Scan(&version)
	if err != nil {
		t.Errorf("Migration version not recorded: %v", err)
	}
	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}
}

func TestInitialize(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Verify all tables exist
	tables := []string{"migrations", "binaries", "installations", "versions", "downloads", "logs"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s not found: %v", table, err)
		}
	}

	// Verify indexes exist
	indexes := []string{
		"idx_binaries_user_id",
		"idx_installations_binary_id",
		"idx_versions_installation_id",
		"idx_downloads_binary_id",
		"idx_logs_timestamp",
	}
	for _, index := range indexes {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", index).Scan(&name)
		if err != nil {
			t.Errorf("Index %s not found: %v", index, err)
		}
	}
}

func TestGetCurrentVersion(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Version should be 0 before migrations
	version, err := db.getCurrentVersion()
	if err != nil {
		t.Fatalf("Failed to get current version: %v", err)
	}
	if version != 0 {
		t.Errorf("Expected version 0, got %d", version)
	}

	// Run migrations
	if err := db.Migrate(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Version should be 1 after migrations
	version, err = db.getCurrentVersion()
	if err != nil {
		t.Fatalf("Failed to get current version: %v", err)
	}
	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}
}
