package binary

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
			binary, err := AddBinaryFromURL(tt.url, false, dbService)
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
	binary1, err := AddBinaryFromURL(url, false, dbService)
	if err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	// Add same binary again - should return existing one
	binary2, err := AddBinaryFromURL(url, false, dbService)
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
	binary, err := AddBinaryFromURL(url, false, dbService)
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
	_, err = AddBinaryFromURL(url, false, dbService)
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
	added, err := AddBinaryFromURL(url, false, dbService)
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

	// Create a test binary file
	tmpDir := t.TempDir()
	testBinaryPath := filepath.Join(tmpDir, "testbinary")
	err := os.WriteFile(testBinaryPath, []byte("#!/bin/sh\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Import the binary
	binary, err := ImportBinary(testBinaryPath, "testbinary", dbService)
	if err != nil {
		t.Fatalf("Failed to import binary: %v", err)
	}

	if binary == nil {
		t.Fatal("Expected binary but got nil")
	}

	if binary.Provider != "local" {
		t.Errorf("Expected provider 'local', got %q", binary.Provider)
	}

	if binary.Source != "manual" {
		t.Errorf("Expected source 'manual', got %q", binary.Source)
	}

	// Verify installation was created
	installations, err := dbService.Installations.ListByBinary(binary.ID)
	if err != nil {
		t.Fatalf("Failed to list installations: %v", err)
	}

	if len(installations) != 1 {
		t.Errorf("Expected 1 installation, got %d", len(installations))
	}

	// Verify version was set
	version, err := dbService.Versions.Get(binary.ID)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if version.SymlinkPath == "" {
		t.Error("Expected symlink path to be set")
	}
}

func TestImportBinary_NonExistentFile(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := ImportBinary("/nonexistent/path/to/binary", "testbinary", dbService)
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}
}

func TestImportBinary_EmptyName(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	testBinaryPath := filepath.Join(tmpDir, "testbinary")
	err := os.WriteFile(testBinaryPath, []byte("#!/bin/sh\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	_, err = ImportBinary(testBinaryPath, "", dbService)
	if err == nil {
		t.Error("Expected error for empty name, got none")
	}
}

func TestImportBinaryWithOptions_CustomVersion(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	testBinaryPath := filepath.Join(tmpDir, "testbinary")
	err := os.WriteFile(testBinaryPath, []byte("#!/bin/sh\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Import with custom version
	binary, err := ImportBinaryWithOptions(testBinaryPath, "testbinary", "", "1.2.3", false, false, dbService)
	if err != nil {
		t.Fatalf("Failed to import binary: %v", err)
	}

	// Verify version was set correctly
	installations, err := dbService.Installations.ListByBinary(binary.ID)
	if err != nil {
		t.Fatalf("Failed to list installations: %v", err)
	}

	if len(installations) != 1 {
		t.Fatalf("Expected 1 installation, got %d", len(installations))
	}

	if installations[0].Version != "1.2.3" {
		t.Errorf("Expected version '1.2.3', got %q", installations[0].Version)
	}
}

func TestImportBinaryWithOptions_KeepLocation(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	testBinaryPath := filepath.Join(tmpDir, "testbinary")
	err := os.WriteFile(testBinaryPath, []byte("#!/bin/sh\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Import with keep location
	binary, err := ImportBinaryWithOptions(testBinaryPath, "testbinary", "", "", false, true, dbService)
	if err != nil {
		t.Fatalf("Failed to import binary: %v", err)
	}

	// Verify installation uses original path
	installations, err := dbService.Installations.ListByBinary(binary.ID)
	if err != nil {
		t.Fatalf("Failed to list installations: %v", err)
	}

	if len(installations) != 1 {
		t.Fatalf("Expected 1 installation, got %d", len(installations))
	}

	if installations[0].InstalledPath != testBinaryPath {
		t.Errorf("Expected installed path %q, got %q", testBinaryPath, installations[0].InstalledPath)
	}
}

func TestImportBinaryWithOptions_GitHubURL(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	testBinaryPath := filepath.Join(tmpDir, "gh")
	err := os.WriteFile(testBinaryPath, []byte("#!/bin/sh\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Import with GitHub URL
	url := "https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz"
	binary, err := ImportBinaryWithOptions(testBinaryPath, "", url, "v2.30.0", false, false, dbService)
	if err != nil {
		t.Fatalf("Failed to import binary with GitHub URL: %v", err)
	}

	// Verify provider is github
	if binary.Provider != "github" {
		t.Errorf("Expected provider 'github', got %q", binary.Provider)
	}

	// Verify provider path is set correctly
	if binary.ProviderPath != "cli/cli" {
		t.Errorf("Expected provider path 'cli/cli', got %q", binary.ProviderPath)
	}

	// Verify format is .tar.gz
	if binary.Format != ".tar.gz" {
		t.Errorf("Expected format '.tar.gz', got %q", binary.Format)
	}

	// Verify binary ID is extracted from asset name
	if binary.UserID != "gh" {
		t.Errorf("Expected UserID 'gh', got %q", binary.UserID)
	}
}

func TestImportBinary_Duplicate(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	testBinaryPath := filepath.Join(tmpDir, "testbinary")
	err := os.WriteFile(testBinaryPath, []byte("#!/bin/sh\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Import first time
	binary1, err := ImportBinary(testBinaryPath, "testbinary", dbService)
	if err != nil {
		t.Fatalf("First import failed: %v", err)
	}

	// Import same binary again - should return existing one
	binary2, err := ImportBinary(testBinaryPath, "testbinary", dbService)
	if err != nil {
		t.Fatalf("Second import failed: %v", err)
	}

	if binary1.ID != binary2.ID {
		t.Errorf("Expected same binary ID, got %d and %d", binary1.ID, binary2.ID)
	}
}

func TestRemoveBinary_WithFiles(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test directory structure
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "test-binary", "1.0.0")
	symlinkPath := filepath.Join(tmpDir, "test-binary-link")

	// Create installation directory and a test file
	err := os.MkdirAll(installDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create install dir: %v", err)
	}

	testFilePath := filepath.Join(installDir, "test-binary")
	err = os.WriteFile(testFilePath, []byte("test binary content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a symlink
	err = os.Symlink(testFilePath, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Create binary in database
	binary := &database.Binary{
		UserID:        "test-binary-with-files",
		Name:          "TestBinary",
		Provider:      "github",
		ProviderPath:  "test/binary",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:test",
		ConfigVersion: 1,
	}

	err = dbService.Binaries.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installation record
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     installDir,
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = dbService.Installations.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	// Create version record with symlink
	err = dbService.Versions.Set(binary.ID, installation.ID, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Verify files exist before removal
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		t.Error("Install directory should exist before removal")
	}
	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		t.Error("Symlink should exist before removal")
	}

	// Remove binary with file deletion
	err = RemoveBinary(binary.UserID, dbService, true)
	if err != nil {
		t.Errorf("Failed to remove binary with files: %v", err)
	}

	// Verify files are deleted
	if _, err := os.Stat(installDir); !os.IsNotExist(err) {
		t.Error("Install directory should be deleted after removal")
	}
	if _, err := os.Lstat(symlinkPath); !os.IsNotExist(err) {
		t.Error("Symlink should be deleted after removal")
	}

	// Verify database records are deleted
	_, err = dbService.Binaries.GetByUserID(binary.UserID)
	if err == nil {
		t.Error("Binary should be deleted from database")
	}
}

func TestRemoveBinary_WithoutFiles(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test directory structure
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "test-binary", "1.0.0")
	symlinkPath := filepath.Join(tmpDir, "test-binary-link")

	// Create installation directory and a test file
	err := os.MkdirAll(installDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create install dir: %v", err)
	}

	testFilePath := filepath.Join(installDir, "test-binary")
	err = os.WriteFile(testFilePath, []byte("test binary content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a symlink
	err = os.Symlink(testFilePath, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Create binary in database
	binary := &database.Binary{
		UserID:        "test-binary-without-files",
		Name:          "TestBinary",
		Provider:      "github",
		ProviderPath:  "test/binary",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:test",
		ConfigVersion: 1,
	}

	err = dbService.Binaries.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installation record
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     installDir,
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = dbService.Installations.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	// Create version record with symlink
	err = dbService.Versions.Set(binary.ID, installation.ID, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Remove binary WITHOUT file deletion
	err = RemoveBinary(binary.UserID, dbService, false)
	if err != nil {
		t.Errorf("Failed to remove binary without files: %v", err)
	}

	// Verify files still exist
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		t.Error("Install directory should still exist when removeFiles=false")
	}
	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		t.Error("Symlink should still exist when removeFiles=false")
	}

	// Verify database records are deleted
	_, err = dbService.Binaries.GetByUserID(binary.UserID)
	if err == nil {
		t.Error("Binary should be deleted from database")
	}
}

func TestRemoveBinary_MissingFiles(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Create binary in database without creating actual files
	binary := &database.Binary{
		UserID:        "test-binary-missing-files",
		Name:          "TestBinary",
		Provider:      "github",
		ProviderPath:  "test/binary",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:test",
		ConfigVersion: 1,
	}

	err := dbService.Binaries.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installation record with non-existent path
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/nonexistent/path/to/binary",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = dbService.Installations.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	// Create version record with non-existent symlink
	err = dbService.Versions.Set(binary.ID, installation.ID, "/nonexistent/symlink")
	if err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Remove binary with file deletion - should not error on missing files
	err = RemoveBinary(binary.UserID, dbService, true)
	if err != nil {
		t.Errorf("Should not fail when files are missing: %v", err)
	}

	// Verify database records are deleted
	_, err = dbService.Binaries.GetByUserID(binary.UserID)
	if err == nil {
		t.Error("Binary should be deleted from database")
	}
}

func TestRemoveBinary_MultipleInstallations(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test directory structure with multiple versions
	tmpDir := t.TempDir()
	installDir1 := filepath.Join(tmpDir, "test-binary", "1.0.0")
	installDir2 := filepath.Join(tmpDir, "test-binary", "2.0.0")
	symlinkPath := filepath.Join(tmpDir, "test-binary-link")

	// Create both installation directories
	for _, dir := range []string{installDir1, installDir2} {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create install dir: %v", err)
		}

		testFilePath := filepath.Join(dir, "test-binary")
		err = os.WriteFile(testFilePath, []byte("test binary content"), 0755)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create a symlink to version 2.0.0
	err := os.Symlink(filepath.Join(installDir2, "test-binary"), symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Create binary in database
	binary := &database.Binary{
		UserID:        "test-binary-multiple",
		Name:          "TestBinary",
		Provider:      "github",
		ProviderPath:  "test/binary",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:test",
		ConfigVersion: 1,
	}

	err = dbService.Binaries.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installation records for both versions
	installation1 := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     installDir1,
		SourceURL:         "https://example.com/test-1.0.0.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = dbService.Installations.Create(installation1)
	if err != nil {
		t.Fatalf("Failed to create installation1: %v", err)
	}

	installation2 := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "2.0.0",
		InstalledPath:     installDir2,
		SourceURL:         "https://example.com/test-2.0.0.tar.gz",
		FileSize:          2048,
		Checksum:          "def456",
		ChecksumAlgorithm: "sha256",
	}

	err = dbService.Installations.Create(installation2)
	if err != nil {
		t.Fatalf("Failed to create installation2: %v", err)
	}

	// Set active version to 2.0.0
	err = dbService.Versions.Set(binary.ID, installation2.ID, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Remove binary with file deletion
	err = RemoveBinary(binary.UserID, dbService, true)
	if err != nil {
		t.Errorf("Failed to remove binary with multiple installations: %v", err)
	}

	// Verify all installation directories are deleted
	if _, err := os.Stat(installDir1); !os.IsNotExist(err) {
		t.Error("Installation directory 1.0.0 should be deleted")
	}
	if _, err := os.Stat(installDir2); !os.IsNotExist(err) {
		t.Error("Installation directory 2.0.0 should be deleted")
	}
	if _, err := os.Lstat(symlinkPath); !os.IsNotExist(err) {
		t.Error("Symlink should be deleted")
	}

	// Verify database records are deleted
	_, err = dbService.Binaries.GetByUserID(binary.UserID)
	if err == nil {
		t.Error("Binary should be deleted from database")
	}
}
