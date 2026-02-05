package repository

import (
	"testing"

	"cturner8/binmate/internal/database"
)

func TestBinariesRepository_SyncBinary(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	// Test creating a new binary via sync
	t.Run("Create new binary", func(t *testing.T) {
		configBinary := ConfigBinary{
			ID:           "sync-test-1",
			Name:         "SyncTest",
			Alias:        "st",
			Provider:     "github",
			Path:         "test/synctest",
			InstallPath:  "/usr/local/bin",
			Format:       ".tar.gz",
			AssetRegex:   ".*linux.*",
			ReleaseRegex: "v.*",
		}

		err := repo.SyncBinary(configBinary, 1)
		if err != nil {
			t.Fatalf("Failed to sync new binary: %v", err)
		}

		// Verify binary was created
		binary, err := repo.GetByUserID("sync-test-1")
		if err != nil {
			t.Fatalf("Failed to get synced binary: %v", err)
		}

		if binary.Name != "SyncTest" {
			t.Errorf("Expected Name SyncTest, got %s", binary.Name)
		}

		if binary.Alias == nil || *binary.Alias != "st" {
			t.Errorf("Expected Alias st, got %v", binary.Alias)
		}
	})

	// Test updating existing binary
	t.Run("Update existing binary", func(t *testing.T) {
		// Create initial binary
		configBinary1 := ConfigBinary{
			ID:           "sync-test-2",
			Name:         "Original",
			Provider:     "github",
			Path:         "test/original",
			Format:       ".tar.gz",
		}

		err := repo.SyncBinary(configBinary1, 1)
		if err != nil {
			t.Fatalf("Failed to sync initial binary: %v", err)
		}

		// Update with changed config
		configBinary2 := ConfigBinary{
			ID:           "sync-test-2",
			Name:         "Updated",
			Provider:     "github",
			Path:         "test/updated",
			Format:       ".tar.gz",
		}

		err = repo.SyncBinary(configBinary2, 2)
		if err != nil {
			t.Fatalf("Failed to sync updated binary: %v", err)
		}

		// Verify binary was updated
		binary, err := repo.GetByUserID("sync-test-2")
		if err != nil {
			t.Fatalf("Failed to get updated binary: %v", err)
		}

		if binary.Name != "Updated" {
			t.Errorf("Expected Name Updated, got %s", binary.Name)
		}

		if binary.ConfigVersion != 2 {
			t.Errorf("Expected ConfigVersion 2, got %d", binary.ConfigVersion)
		}
	})

	// Test sync with unchanged binary (digest match)
	t.Run("Skip sync when digest matches", func(t *testing.T) {
		configBinary := ConfigBinary{
			ID:       "sync-test-3",
			Name:     "Unchanged",
			Provider: "github",
			Path:     "test/unchanged",
			Format:   ".tar.gz",
		}

		// First sync
		err := repo.SyncBinary(configBinary, 1)
		if err != nil {
			t.Fatalf("Failed to sync binary: %v", err)
		}

		binary1, err := repo.GetByUserID("sync-test-3")
		if err != nil {
			t.Fatalf("Failed to get binary: %v", err)
		}

		originalUpdatedAt := binary1.UpdatedAt

		// Second sync with same config (should skip update)
		err = repo.SyncBinary(configBinary, 1)
		if err != nil {
			t.Fatalf("Failed to sync unchanged binary: %v", err)
		}

		binary2, err := repo.GetByUserID("sync-test-3")
		if err != nil {
			t.Fatalf("Failed to get binary after second sync: %v", err)
		}

		if binary2.UpdatedAt != originalUpdatedAt {
			t.Error("Expected UpdatedAt to remain unchanged when digest matches")
		}
	})
}

func TestBinariesRepository_SyncFromConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	// Test syncing multiple binaries
	t.Run("Sync multiple binaries", func(t *testing.T) {
		configBinaries := []ConfigBinary{
			{
				ID:       "multi-1",
				Name:     "Binary1",
				Provider: "github",
				Path:     "test/binary1",
				Format:   ".tar.gz",
			},
			{
				ID:       "multi-2",
				Name:     "Binary2",
				Provider: "github",
				Path:     "test/binary2",
				Format:   ".zip",
			},
			{
				ID:       "multi-3",
				Name:     "Binary3",
				Provider: "github",
				Path:     "test/binary3",
				Format:   ".tar.gz",
			},
		}

		err := repo.SyncFromConfig(configBinaries, 1)
		if err != nil {
			t.Fatalf("Failed to sync binaries from config: %v", err)
		}

		// Verify all binaries were created
		binaries, err := repo.List()
		if err != nil {
			t.Fatalf("Failed to list binaries: %v", err)
		}

		if len(binaries) != 3 {
			t.Errorf("Expected 3 binaries, got %d", len(binaries))
		}
	})

	// Test removing binary not in config
	t.Run("Remove binary not in config", func(t *testing.T) {
		// Create initial set
		initialBinaries := []ConfigBinary{
			{
				ID:       "remove-1",
				Name:     "Binary1",
				Provider: "github",
				Path:     "test/binary1",
				Format:   ".tar.gz",
			},
			{
				ID:       "remove-2",
				Name:     "Binary2",
				Provider: "github",
				Path:     "test/binary2",
				Format:   ".tar.gz",
			},
		}

		err := repo.SyncFromConfig(initialBinaries, 1)
		if err != nil {
			t.Fatalf("Failed to sync initial binaries: %v", err)
		}

		// Sync again with only one binary
		updatedBinaries := []ConfigBinary{
			{
				ID:       "remove-1",
				Name:     "Binary1",
				Provider: "github",
				Path:     "test/binary1",
				Format:   ".tar.gz",
			},
		}

		err = repo.SyncFromConfig(updatedBinaries, 2)
		if err != nil {
			t.Fatalf("Failed to sync updated binaries: %v", err)
		}

		// Verify remove-2 was deleted
		_, err = repo.GetByUserID("remove-2")
		if err != database.ErrNotFound {
			t.Errorf("Expected binary remove-2 to be deleted, got err: %v", err)
		}

		// Verify remove-1 still exists
		_, err = repo.GetByUserID("remove-1")
		if err != nil {
			t.Errorf("Expected binary remove-1 to exist, got err: %v", err)
		}
	})

	// Test updating existing binary
	t.Run("Update existing binary in sync", func(t *testing.T) {
		// Create initial binary
		initialBinaries := []ConfigBinary{
			{
				ID:       "update-sync-1",
				Name:     "OriginalName",
				Provider: "github",
				Path:     "test/original",
				Format:   ".tar.gz",
			},
		}

		err := repo.SyncFromConfig(initialBinaries, 1)
		if err != nil {
			t.Fatalf("Failed to sync initial binary: %v", err)
		}

		// Update the binary
		updatedBinaries := []ConfigBinary{
			{
				ID:       "update-sync-1",
				Name:     "UpdatedName",
				Provider: "github",
				Path:     "test/updated",
				Format:   ".tar.gz",
			},
		}

		err = repo.SyncFromConfig(updatedBinaries, 2)
		if err != nil {
			t.Fatalf("Failed to sync updated binary: %v", err)
		}

		// Verify binary was updated
		binary, err := repo.GetByUserID("update-sync-1")
		if err != nil {
			t.Fatalf("Failed to get updated binary: %v", err)
		}

		if binary.Name != "UpdatedName" {
			t.Errorf("Expected Name UpdatedName, got %s", binary.Name)
		}

		if binary.ConfigVersion != 2 {
			t.Errorf("Expected ConfigVersion 2, got %d", binary.ConfigVersion)
		}
	})

	// Test skipping unchanged binary
	t.Run("Skip unchanged binary in sync", func(t *testing.T) {
		// Create binary
		binaries := []ConfigBinary{
			{
				ID:       "unchanged-sync-1",
				Name:     "Unchanged",
				Provider: "github",
				Path:     "test/unchanged",
				Format:   ".tar.gz",
			},
		}

		err := repo.SyncFromConfig(binaries, 1)
		if err != nil {
			t.Fatalf("Failed to sync binary: %v", err)
		}

		binary1, err := repo.GetByUserID("unchanged-sync-1")
		if err != nil {
			t.Fatalf("Failed to get binary: %v", err)
		}

		originalUpdatedAt := binary1.UpdatedAt

		// Sync again with same config
		err = repo.SyncFromConfig(binaries, 1)
		if err != nil {
			t.Fatalf("Failed to sync unchanged binary: %v", err)
		}

		binary2, err := repo.GetByUserID("unchanged-sync-1")
		if err != nil {
			t.Fatalf("Failed to get binary after second sync: %v", err)
		}

		if binary2.UpdatedAt != originalUpdatedAt {
			t.Error("Expected UpdatedAt to remain unchanged for matching digest")
		}
	})
}

func TestStringToPtr(t *testing.T) {
	// Test empty string returns nil
	result := stringToPtr("")
	if result != nil {
		t.Error("Expected nil for empty string")
	}

	// Test non-empty string returns pointer
	result = stringToPtr("test")
	if result == nil {
		t.Fatal("Expected non-nil for non-empty string")
	}

	if *result != "test" {
		t.Errorf("Expected 'test', got %s", *result)
	}
}
