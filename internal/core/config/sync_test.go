package config

import (
	"path/filepath"
	"testing"

	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

func TestSyncToDatabase(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)

	// Create test config
	config := Config{
		Version: 1,
		Binaries: []Binary{
			{
				Id:       "test1",
				Name:     "test-binary",
				Provider: "github",
				Path:     "owner/repo",
				Format:   ".tar.gz",
			},
		},
	}

	// Sync to database
	if err := SyncToDatabase(config, dbService); err != nil {
		t.Fatalf("Failed to sync to database: %v", err)
	}

	// Verify binary was created
	binaries, err := dbService.Binaries.List()
	if err != nil {
		t.Fatalf("Failed to list binaries: %v", err)
	}

	if len(binaries) != 1 {
		t.Errorf("Expected 1 binary, got %d", len(binaries))
	}

	if binaries[0].UserID != "test1" {
		t.Errorf("Expected UserID 'test1', got '%s'", binaries[0].UserID)
	}

	if binaries[0].Name != "test-binary" {
		t.Errorf("Expected Name 'test-binary', got '%s'", binaries[0].Name)
	}

	// Update config (add binary)
	config.Binaries = append(config.Binaries, Binary{
		Id:       "test2",
		Name:     "another-binary",
		Provider: "github",
		Path:     "owner/repo2",
		Format:   ".zip",
	})

	// Sync again
	if err := SyncToDatabase(config, dbService); err != nil {
		t.Fatalf("Failed to sync updated config: %v", err)
	}

	// Verify we now have 2 binaries
	binaries, err = dbService.Binaries.List()
	if err != nil {
		t.Fatalf("Failed to list binaries: %v", err)
	}

	if len(binaries) != 2 {
		t.Errorf("Expected 2 binaries, got %d", len(binaries))
	}

	// Remove one from config
	config.Binaries = config.Binaries[1:] // Keep only test2

	// Sync again
	if err := SyncToDatabase(config, dbService); err != nil {
		t.Fatalf("Failed to sync with removed binary: %v", err)
	}

	// Verify we now have 1 binary
	binaries, err = dbService.Binaries.List()
	if err != nil {
		t.Fatalf("Failed to list binaries: %v", err)
	}

	if len(binaries) != 1 {
		t.Errorf("Expected 1 binary after removal, got %d", len(binaries))
	}

	if binaries[0].UserID != "test2" {
		t.Errorf("Expected remaining binary to be 'test2', got '%s'", binaries[0].UserID)
	}
}

func TestSyncToDatabase_ManuallyAddedBinariesPersist(t *testing.T) {
	// This test verifies that manually added binaries (via add command or TUI)
	// are NOT deleted when syncing config, even if they're not in the config file.
	// See bug bm-5m2 for context.

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)

	// Create initial config with one binary
	config := Config{
		Version: 1,
		Binaries: []Binary{
			{
				Id:       "config-binary",
				Name:     "from-config",
				Provider: "github",
				Path:     "owner/repo",
				Format:   ".tar.gz",
			},
		},
	}

	// Sync to database - this creates a config-managed binary
	if err := SyncToDatabase(config, dbService); err != nil {
		t.Fatalf("Failed to sync to database: %v", err)
	}

	// Manually add a binary (simulating `binmate add <url>` or TUI add)
	manualBinary := &database.Binary{
		UserID:        "manual-binary",
		Name:          "manually-added",
		Provider:      "github",
		ProviderPath:  "owner/other-repo",
		Format:        ".zip",
		ConfigVersion: 0,
		Source:        "manual",
	}
	if err := dbService.Binaries.Create(manualBinary); err != nil {
		t.Fatalf("Failed to create manual binary: %v", err)
	}

	// Verify we have 2 binaries (1 config, 1 manual)
	binaries, err := dbService.Binaries.List()
	if err != nil {
		t.Fatalf("Failed to list binaries: %v", err)
	}
	if len(binaries) != 2 {
		t.Fatalf("Expected 2 binaries, got %d", len(binaries))
	}

	// Remove config-binary from config and sync
	config.Binaries = []Binary{} // Empty config

	if err := SyncToDatabase(config, dbService); err != nil {
		t.Fatalf("Failed to sync empty config: %v", err)
	}

	// Verify manual binary still exists, but config binary was deleted
	binaries, err = dbService.Binaries.List()
	if err != nil {
		t.Fatalf("Failed to list binaries after sync: %v", err)
	}

	if len(binaries) != 1 {
		t.Fatalf("Expected 1 binary (manual) after sync, got %d", len(binaries))
	}

	if binaries[0].UserID != "manual-binary" {
		t.Errorf("Expected manual-binary to persist, got '%s'", binaries[0].UserID)
	}

	if binaries[0].Source != "manual" {
		t.Errorf("Expected source='manual', got '%s'", binaries[0].Source)
	}
}

func TestSyncToDatabase_SourceFieldSetCorrectly(t *testing.T) {
	// Verify that binaries from config get source='config'

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)

	config := Config{
		Version: 1,
		Binaries: []Binary{
			{
				Id:       "test1",
				Name:     "test-binary",
				Provider: "github",
				Path:     "owner/repo",
				Format:   ".tar.gz",
			},
		},
	}

	if err := SyncToDatabase(config, dbService); err != nil {
		t.Fatalf("Failed to sync to database: %v", err)
	}

	binaries, err := dbService.Binaries.List()
	if err != nil {
		t.Fatalf("Failed to list binaries: %v", err)
	}

	if len(binaries) != 1 {
		t.Fatalf("Expected 1 binary, got %d", len(binaries))
	}

	if binaries[0].Source != "config" {
		t.Errorf("Expected source='config' for config-managed binary, got '%s'", binaries[0].Source)
	}
}
