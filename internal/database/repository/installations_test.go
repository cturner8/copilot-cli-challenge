package repository

import (
	"testing"
	"time"

	"cturner8/binmate/internal/database"
)

func TestInstallationsRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)

	// Create binary first
	binary := &database.Binary{
		UserID:        "test-inst",
		Name:          "TestInst",
		Provider:      "github",
		ProviderPath:  "test/inst",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:inst",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	if installation.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	if installation.InstalledAt == 0 {
		t.Error("Expected InstalledAt to be set")
	}
}

func TestInstallationsRepository_CreateForeignKeyConstraint(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instRepo := NewInstallationsRepository(db)

	// Try to create installation with non-existent binary ID
	installation := &database.Installation{
		BinaryID:          999,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err := instRepo.Create(installation)
	if err == nil {
		t.Error("Expected foreign key constraint error, got nil")
	}
}

func TestInstallationsRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)

	binary := &database.Binary{
		UserID:        "test-get-inst",
		Name:          "TestGetInst",
		Provider:      "github",
		ProviderPath:  "test/getinst",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:getinst",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "2.0.0",
		InstalledPath:     "/usr/local/bin/test",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          2048,
		Checksum:          "def456",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	retrieved, err := instRepo.Get(binary.ID, "2.0.0")
	if err != nil {
		t.Fatalf("Failed to get installation: %v", err)
	}

	if retrieved.BinaryID != binary.ID {
		t.Errorf("Expected BinaryID %d, got %d", binary.ID, retrieved.BinaryID)
	}

	if retrieved.Version != "2.0.0" {
		t.Errorf("Expected Version 2.0.0, got %s", retrieved.Version)
	}

	if retrieved.FileSize != 2048 {
		t.Errorf("Expected FileSize 2048, got %d", retrieved.FileSize)
	}
}

func TestInstallationsRepository_GetNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instRepo := NewInstallationsRepository(db)

	_, err := instRepo.Get(999, "1.0.0")
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestInstallationsRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)

	binary := &database.Binary{
		UserID:        "test-getbyid-inst",
		Name:          "TestGetByIDInst",
		Provider:      "github",
		ProviderPath:  "test/getbyidinst",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:getbyidinst",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "3.0.0",
		InstalledPath:     "/usr/local/bin/test",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          3072,
		Checksum:          "ghi789",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	retrieved, err := instRepo.GetByID(installation.ID)
	if err != nil {
		t.Fatalf("Failed to get installation by ID: %v", err)
	}

	if retrieved.ID != installation.ID {
		t.Errorf("Expected ID %d, got %d", installation.ID, retrieved.ID)
	}

	if retrieved.Version != "3.0.0" {
		t.Errorf("Expected Version 3.0.0, got %s", retrieved.Version)
	}
}

func TestInstallationsRepository_GetByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instRepo := NewInstallationsRepository(db)

	_, err := instRepo.GetByID(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestInstallationsRepository_ListByBinary(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)

	binary := &database.Binary{
		UserID:        "test-list-inst",
		Name:          "TestListInst",
		Provider:      "github",
		ProviderPath:  "test/listinst",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:listinst",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create multiple installations
	installations := []string{"1.0.0", "1.1.0", "1.2.0"}
	for _, version := range installations {
		installation := &database.Installation{
			BinaryID:          binary.ID,
			Version:           version,
			InstalledPath:     "/usr/local/bin/test-" + version,
			SourceURL:         "https://example.com/test-" + version + ".tar.gz",
			FileSize:          1024,
			Checksum:          "checksum-" + version,
			ChecksumAlgorithm: "sha256",
		}

		err = instRepo.Create(installation)
		if err != nil {
			t.Fatalf("Failed to create installation %s: %v", version, err)
		}
	}

	list, err := instRepo.ListByBinary(binary.ID)
	if err != nil {
		t.Fatalf("Failed to list installations: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 installations, got %d", len(list))
	}

	// Verify ordered by installed_at DESC
	for i := 0; i < len(list)-1; i++ {
		if list[i].InstalledAt < list[i+1].InstalledAt {
			t.Error("Installations are not sorted by InstalledAt DESC")
		}
	}
}

func TestInstallationsRepository_ListByBinaryEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)

	binary := &database.Binary{
		UserID:        "test-list-empty-inst",
		Name:          "TestListEmptyInst",
		Provider:      "github",
		ProviderPath:  "test/listemptyinst",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:listemptyinst",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	list, err := instRepo.ListByBinary(binary.ID)
	if err != nil {
		t.Fatalf("Failed to list installations: %v", err)
	}

	if len(list) != 0 {
		t.Errorf("Expected 0 installations, got %d", len(list))
	}
}

func TestInstallationsRepository_GetLatest(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)

	binary := &database.Binary{
		UserID:        "test-latest-inst",
		Name:          "TestLatestInst",
		Provider:      "github",
		ProviderPath:  "test/latestinst",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:latestinst",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installations with distinct timestamps
	versions := []string{"1.0.0", "2.0.0", "3.0.0"}
	var lastInstallation *database.Installation

	for _, version := range versions {
		installation := &database.Installation{
			BinaryID:          binary.ID,
			Version:           version,
			InstalledPath:     "/usr/local/bin/test-" + version,
			SourceURL:         "https://example.com/test-" + version + ".tar.gz",
			FileSize:          1024,
			Checksum:          "checksum-" + version,
			ChecksumAlgorithm: "sha256",
		}

		err = instRepo.Create(installation)
		if err != nil {
			t.Fatalf("Failed to create installation %s: %v", version, err)
		}

		lastInstallation = installation
		// Ensure distinct timestamps (Unix timestamps are in seconds)
		time.Sleep(1100 * time.Millisecond)
	}

	latest, err := instRepo.GetLatest(binary.ID)
	if err != nil {
		t.Fatalf("Failed to get latest installation: %v", err)
	}

	if latest.Version != "3.0.0" {
		t.Errorf("Expected latest version 3.0.0, got %s", latest.Version)
	}

	if latest.ID != lastInstallation.ID {
		t.Errorf("Expected latest ID %d, got %d", lastInstallation.ID, latest.ID)
	}
}

func TestInstallationsRepository_GetLatestNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instRepo := NewInstallationsRepository(db)

	_, err := instRepo.GetLatest(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestInstallationsRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)

	binary := &database.Binary{
		UserID:        "test-delete-inst",
		Name:          "TestDeleteInst",
		Provider:      "github",
		ProviderPath:  "test/deleteinst",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:deleteinst",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	err = instRepo.Delete(installation.ID)
	if err != nil {
		t.Fatalf("Failed to delete installation: %v", err)
	}

	_, err = instRepo.GetByID(installation.ID)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}
}

func TestInstallationsRepository_DeleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instRepo := NewInstallationsRepository(db)

	err := instRepo.Delete(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestInstallationsRepository_VerifyChecksum(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)

	binary := &database.Binary{
		UserID:        "test-verify-inst",
		Name:          "TestVerifyInst",
		Provider:      "github",
		ProviderPath:  "test/verifyinst",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:verifyinst",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "correctchecksum",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	// Test with correct checksum
	valid, err := instRepo.VerifyChecksum(installation.ID, "correctchecksum")
	if err != nil {
		t.Fatalf("Failed to verify checksum: %v", err)
	}

	if !valid {
		t.Error("Expected checksum to be valid")
	}

	// Test with incorrect checksum
	valid, err = instRepo.VerifyChecksum(installation.ID, "wrongchecksum")
	if err != nil {
		t.Fatalf("Failed to verify checksum: %v", err)
	}

	if valid {
		t.Error("Expected checksum to be invalid")
	}
}

func TestInstallationsRepository_VerifyChecksumNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instRepo := NewInstallationsRepository(db)

	_, err := instRepo.VerifyChecksum(999, "checksum")
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
