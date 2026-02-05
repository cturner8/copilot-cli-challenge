package repository

import (
	"path/filepath"
	"testing"

	"cturner8/binmate/internal/database"
)

func TestNewService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.DB == nil {
		t.Error("Expected DB to be set")
	}

	if service.Binaries == nil {
		t.Error("Expected Binaries repository to be initialized")
	}

	if service.Installations == nil {
		t.Error("Expected Installations repository to be initialized")
	}

	if service.Versions == nil {
		t.Error("Expected Versions repository to be initialized")
	}

	if service.Downloads == nil {
		t.Error("Expected Downloads repository to be initialized")
	}

	if service.Logs == nil {
		t.Error("Expected Logs repository to be initialized")
	}
}

func TestService_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	service := NewService(db)

	err = service.Close()
	if err != nil {
		t.Fatalf("Failed to close service: %v", err)
	}

	// Try to use the database after closing (should fail)
	err = db.Ping()
	if err == nil {
		t.Error("Expected ping to fail after close")
	}
}

func TestService_AllRepositoriesAccessible(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create binary using service
	binary := &database.Binary{
		UserID:        "test-service",
		Name:          "TestService",
		Provider:      "github",
		ProviderPath:  "test/service",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:service",
		ConfigVersion: 1,
	}

	err := service.Binaries.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary via service: %v", err)
	}

	// Create installation using service
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = service.Installations.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation via service: %v", err)
	}

	// Set version using service
	err = service.Versions.Set(binary.ID, installation.ID, "/usr/local/bin/test")
	if err != nil {
		t.Fatalf("Failed to set version via service: %v", err)
	}

	// Create download using service
	download := &database.Download{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		CachePath:         "/tmp/cache/test.tar.gz",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
		IsComplete:        true,
	}

	err = service.Downloads.Create(download)
	if err != nil {
		t.Fatalf("Failed to create download via service: %v", err)
	}

	// Create log using service
	logID, err := service.Logs.LogStart("install", "binary", "test-service", "Test operation")
	if err != nil {
		t.Fatalf("Failed to create log via service: %v", err)
	}

	if logID == 0 {
		t.Error("Expected logID to be set")
	}

	// Verify all entities were created
	retrievedBinary, err := service.Binaries.Get(binary.ID)
	if err != nil {
		t.Errorf("Failed to retrieve binary: %v", err)
	}
	if retrievedBinary.ID != binary.ID {
		t.Errorf("Binary ID mismatch")
	}

	retrievedInstallation, err := service.Installations.GetByID(installation.ID)
	if err != nil {
		t.Errorf("Failed to retrieve installation: %v", err)
	}
	if retrievedInstallation.ID != installation.ID {
		t.Errorf("Installation ID mismatch")
	}

	retrievedVersion, err := service.Versions.Get(binary.ID)
	if err != nil {
		t.Errorf("Failed to retrieve version: %v", err)
	}
	if retrievedVersion.BinaryID != binary.ID {
		t.Errorf("Version BinaryID mismatch")
	}

	retrievedDownload, err := service.Downloads.Get(binary.ID, "1.0.0")
	if err != nil {
		t.Errorf("Failed to retrieve download: %v", err)
	}
	if retrievedDownload.ID != download.ID {
		t.Errorf("Download ID mismatch")
	}

	logs, err := service.Logs.GetRecent(1)
	if err != nil {
		t.Errorf("Failed to retrieve logs: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}
}
