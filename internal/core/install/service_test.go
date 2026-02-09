package install

import (
	"os"
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
	// This test verifies that after installation, InstalledPath stores the actual
	// binary path (not the symlink path) to prevent circular symlinks when switching
	// versions. See bug bm-j1h for context.
	//
	// NOTE: We cannot test InstallBinary() directly because it requires network calls
	// to GitHub API and actual file downloads. Instead, we simulate the post-installation
	// state by manually creating the filesystem structure and database records that
	// InstallBinary() would create, then verify the contract is correct.
	//
	// The contract is:
	// - installation.InstalledPath = path to actual binary file (e.g., /versions/bin/v1.0.0/bin)
	// - version.SymlinkPath = path to symlink (e.g., /bin/bin)
	// - These MUST be different to avoid circular symlinks

	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	binary := createTestBinary(t, dbService, "test")

	// Simulate what InstallBinary does after extraction:
	// 1. Extract binary to versioned directory
	versionDir := filepath.Join(tmpDir, "versions", "testbin", "v1.0.0")
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		t.Fatalf("Failed to create version directory: %v", err)
	}
	binaryPath := filepath.Join(versionDir, "testbin")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	// 2. Create symlink to bin directory
	binDir := filepath.Join(tmpDir, "bin")
	symlinkPath := filepath.Join(binDir, "testbin")

	// This mimics SetActiveVersion being called
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}
	if err := os.Symlink(binaryPath, symlinkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// 3. Store installation with BINARY PATH (not symlink path) - this is the fix
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "v1.0.0",
		InstalledPath:     binaryPath, // MUST be binary path, not symlinkPath
		SourceURL:         "https://github.com/test/test/releases/download/v1.0.0/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "SHA256",
	}

	if err := dbService.Installations.Create(installation); err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	// 4. Store version with SYMLINK PATH
	if err := dbService.Versions.Set(binary.ID, installation.ID, symlinkPath); err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Verify the contract:
	retrieved, err := dbService.Installations.Get(binary.ID, "v1.0.0")
	if err != nil {
		t.Fatalf("Failed to retrieve installation: %v", err)
	}

	if retrieved.InstalledPath != binaryPath {
		t.Errorf("InstalledPath should store actual binary path, got %s, want %s", retrieved.InstalledPath, binaryPath)
	}

	version, err := dbService.Versions.Get(binary.ID)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if version.SymlinkPath != symlinkPath {
		t.Errorf("SymlinkPath should be stored in versions table, got %s, want %s", version.SymlinkPath, symlinkPath)
	}

	// The critical assertion: these MUST be different
	if retrieved.InstalledPath == version.SymlinkPath {
		t.Error("InstalledPath and SymlinkPath should be different to avoid circular symlinks")
	}

	// Verify the symlink actually points to the binary (not to itself)
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}
	if target != binaryPath {
		t.Errorf("Symlink should point to binary path, got %s, want %s", target, binaryPath)
	}
}
