package binary

import (
	"path/filepath"
	"testing"

	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

func setupTestDB(t *testing.T) (*repository.Service, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	dbService := repository.NewService(db)
	cleanup := func() {
		db.Close()
	}

	return dbService, cleanup
}

func TestAddBinaryFromURL(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "valid GitHub release URL",
			url:         "https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz",
			expectError: false,
		},
		{
			name:        "invalid URL",
			url:         "not-a-url",
			expectError: true,
		},
		{
			name:        "non-GitHub URL",
			url:         "https://example.com/file.tar.gz",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binary, err := AddBinaryFromURL(tt.url, dbService)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if binary == nil {
					t.Error("Expected binary but got nil")
				}
				if binary != nil {
					if binary.Provider != "github" {
						t.Errorf("Expected provider 'github', got %q", binary.Provider)
					}
					if binary.ConfigVersion != 0 {
						t.Errorf("Expected ConfigVersion 0, got %d", binary.ConfigVersion)
					}
				}
			}
		})
	}
}

func TestAddBinaryFromURL_Duplicate(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	url := "https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz"

	// Add binary first time
	binary1, err := AddBinaryFromURL(url, dbService)
	if err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	// Add same binary again - should return existing one
	binary2, err := AddBinaryFromURL(url, dbService)
	if err != nil {
		t.Fatalf("Second add failed: %v", err)
	}

	if binary1.ID != binary2.ID {
		t.Errorf("Expected same binary ID, got %d and %d", binary1.ID, binary2.ID)
	}
}

func TestRemoveBinary(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Add a binary first
	url := "https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz"
	binary, err := AddBinaryFromURL(url, dbService)
	if err != nil {
		t.Fatalf("Failed to add binary: %v", err)
	}

	// Remove it
	err = RemoveBinary(binary.UserID, dbService, false)
	if err != nil {
		t.Errorf("Failed to remove binary: %v", err)
	}

	// Verify it's gone
	_, err = dbService.Binaries.GetByUserID(binary.UserID)
	if err == nil {
		t.Error("Expected error when getting removed binary, got none")
	}
}

func TestRemoveBinary_NotFound(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	err := RemoveBinary("nonexistent", dbService, false)
	if err == nil {
		t.Error("Expected error for non-existent binary, got none")
	}
}

func TestListBinariesWithDetails(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Initially should be empty
	binaries, err := ListBinariesWithDetails(dbService)
	if err != nil {
		t.Fatalf("Failed to list binaries: %v", err)
	}
	if len(binaries) != 0 {
		t.Errorf("Expected 0 binaries, got %d", len(binaries))
	}

	// Add a binary
	url := "https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz"
	_, err = AddBinaryFromURL(url, dbService)
	if err != nil {
		t.Fatalf("Failed to add binary: %v", err)
	}

	// Now should have 1
	binaries, err = ListBinariesWithDetails(dbService)
	if err != nil {
		t.Fatalf("Failed to list binaries after add: %v", err)
	}
	if len(binaries) != 1 {
		t.Errorf("Expected 1 binary, got %d", len(binaries))
	}
}

func TestGetBinaryByID(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Add a binary
	url := "https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz"
	added, err := AddBinaryFromURL(url, dbService)
	if err != nil {
		t.Fatalf("Failed to add binary: %v", err)
	}

	// Get it by ID
	retrieved, err := GetBinaryByID(added.UserID, dbService)
	if err != nil {
		t.Errorf("Failed to get binary by ID: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected binary but got nil")
	}
	if retrieved.ID != added.ID {
		t.Errorf("Expected ID %d, got %d", added.ID, retrieved.ID)
	}
}

func TestGetBinaryByID_NotFound(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := GetBinaryByID("nonexistent", dbService)
	if err == nil {
		t.Error("Expected error for non-existent binary, got none")
	}
}

func TestImportBinary(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Import is not yet implemented
	_, err := ImportBinary("/path/to/binary", "testbinary", dbService)
	if err == nil {
		t.Error("Expected error for unimplemented import, got none")
	}
}
