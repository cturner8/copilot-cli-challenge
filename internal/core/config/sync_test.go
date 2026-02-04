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
