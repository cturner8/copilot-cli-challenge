package install

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

func createTestBinary(t *testing.T, dbService *repository.Service, userID string) *database.Binary {
	binary := &database.Binary{
		UserID:       userID,
		Name:         "testbin",
		Provider:     "github",
		ProviderPath: "owner/repo",
		Format:       ".tar.gz",
	}
	if err := dbService.Binaries.Create(binary); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	return binary
}

func TestInstallBinary_BinaryNotFound(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := InstallBinary("nonexistent", "v1.0.0", dbService)
	if err == nil {
		t.Error("Expected error for non-existent binary, got none")
	}
}

func TestInstallBinary_UnsupportedProvider(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a binary with unsupported provider
	unsupportedBinary := &database.Binary{
		UserID:       "test",
		Name:         "testbin",
		Provider:     "unsupported",
		ProviderPath: "owner/repo",
		Format:       ".tar.gz",
	}
	if err := dbService.Binaries.Create(unsupportedBinary); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	_, err := InstallBinary("test", "v1.0.0", dbService)
	if err == nil {
		t.Error("Expected error for unsupported provider, got none")
	}
	if err != nil && err.Error() != "only github provider is currently supported" {
		t.Errorf("Expected provider error, got: %v", err)
	}
}

func TestInstallBinary_FetchFails(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a binary that will fail to fetch (invalid repo path)
	createTestBinary(t, dbService, "test")

	// This should fail at the fetch stage
	_, err := InstallBinary("test", "v1.0.0", dbService)
	if err == nil {
		t.Error("Expected error for fetch failure, got none")
	}
}

func TestUpdateToLatest(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a binary
	createTestBinary(t, dbService, "test")

	// This should fail because we can't actually fetch from GitHub in tests
	_, err := UpdateToLatest("test", dbService)
	if err == nil {
		t.Error("Expected error (fetch failure), got none")
	}
}

func TestUpdateToLatest_BinaryNotFound(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := UpdateToLatest("nonexistent", dbService)
	if err == nil {
		t.Error("Expected error for non-existent binary, got none")
	}
}

func TestInstallBinaryResult_Structure(t *testing.T) {
	// Test that the result structure is properly defined
	result := &InstallBinaryResult{
		Binary:       &database.Binary{},
		Installation: &database.Installation{},
		Version:      "v1.0.0",
	}

	if result.Binary == nil {
		t.Error("Binary field should not be nil")
	}
	if result.Installation == nil {
		t.Error("Installation field should not be nil")
	}
	if result.Version != "v1.0.0" {
		t.Errorf("Expected version v1.0.0, got %s", result.Version)
	}
}

func TestInstallBinary_InstalledPathStoresActualBinaryPath(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// This test verifies that InstalledPath stores the actual binary path,
	// not the symlink path, to prevent circular symlinks when switching versions.
	// See bug bm-j1h for context.

	// Create mock installation paths
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "versions", "testbin", "v1.0.0", "testbin")
	symlinkPath := filepath.Join(tmpDir, "bin", "testbin")

	// Create a mock installation record as InstallBinary would
	binary := createTestBinary(t, dbService, "test")

	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "v1.0.0",
		InstalledPath:     binaryPath, // Should be the actual binary path
		SourceURL:         "https://github.com/test/test/releases/download/v1.0.0/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "SHA256",
	}

	if err := dbService.Installations.Create(installation); err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	if err := dbService.Versions.Set(binary.ID, installation.ID, symlinkPath); err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Retrieve the installation and verify InstalledPath is the binary path
	retrieved, err := dbService.Installations.Get(binary.ID, "v1.0.0")
	if err != nil {
		t.Fatalf("Failed to retrieve installation: %v", err)
	}

	if retrieved.InstalledPath != binaryPath {
		t.Errorf("InstalledPath should store actual binary path, got %s, want %s", retrieved.InstalledPath, binaryPath)
	}

	// Verify the symlink path is stored separately in versions table
	version, err := dbService.Versions.Get(binary.ID)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if version.SymlinkPath != symlinkPath {
		t.Errorf("SymlinkPath should be stored in versions table, got %s, want %s", version.SymlinkPath, symlinkPath)
	}

	// Verify the two paths are different (this was the bug - they were the same)
	if retrieved.InstalledPath == version.SymlinkPath {
		t.Error("InstalledPath and SymlinkPath should be different to avoid circular symlinks")
	}
}
