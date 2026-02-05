package repository

import (
	"testing"

	"cturner8/binmate/internal/database"
)

func TestVersionsRepository_Set(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)
	verRepo := NewVersionsRepository(db)

	// Create binary
	binary := &database.Binary{
		UserID:        "test-ver-set",
		Name:          "TestVerSet",
		Provider:      "github",
		ProviderPath:  "test/verset",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:verset",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installation
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test-1.0.0",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	// Set active version
	symlinkPath := "/usr/local/bin/test"
	err = verRepo.Set(binary.ID, installation.ID, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Verify
	version, err := verRepo.Get(binary.ID)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if version.BinaryID != binary.ID {
		t.Errorf("Expected BinaryID %d, got %d", binary.ID, version.BinaryID)
	}

	if version.InstallationID != installation.ID {
		t.Errorf("Expected InstallationID %d, got %d", installation.ID, version.InstallationID)
	}

	if version.SymlinkPath != symlinkPath {
		t.Errorf("Expected SymlinkPath %s, got %s", symlinkPath, version.SymlinkPath)
	}

	if version.ActivatedAt == 0 {
		t.Error("Expected ActivatedAt to be set")
	}
}

func TestVersionsRepository_SetUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)
	verRepo := NewVersionsRepository(db)

	// Create binary
	binary := &database.Binary{
		UserID:        "test-ver-update",
		Name:          "TestVerUpdate",
		Provider:      "github",
		ProviderPath:  "test/verupdate",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:verupdate",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create first installation
	inst1 := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test-1.0.0",
		SourceURL:         "https://example.com/test-1.0.0.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(inst1)
	if err != nil {
		t.Fatalf("Failed to create installation 1: %v", err)
	}

	// Set initial version
	err = verRepo.Set(binary.ID, inst1.ID, "/usr/local/bin/test")
	if err != nil {
		t.Fatalf("Failed to set initial version: %v", err)
	}

	// Create second installation
	inst2 := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "2.0.0",
		InstalledPath:     "/usr/local/bin/test-2.0.0",
		SourceURL:         "https://example.com/test-2.0.0.tar.gz",
		FileSize:          2048,
		Checksum:          "def456",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(inst2)
	if err != nil {
		t.Fatalf("Failed to create installation 2: %v", err)
	}

	// Update to second version
	err = verRepo.Set(binary.ID, inst2.ID, "/usr/local/bin/test")
	if err != nil {
		t.Fatalf("Failed to update version: %v", err)
	}

	// Verify update
	version, err := verRepo.Get(binary.ID)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if version.InstallationID != inst2.ID {
		t.Errorf("Expected InstallationID %d, got %d", inst2.ID, version.InstallationID)
	}
}

func TestVersionsRepository_SetForeignKeyConstraint(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	verRepo := NewVersionsRepository(db)

	// Try to set version with non-existent binary ID
	err := verRepo.Set(999, 999, "/usr/local/bin/test")
	if err == nil {
		t.Error("Expected foreign key constraint error, got nil")
	}
}

func TestVersionsRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)
	verRepo := NewVersionsRepository(db)

	// Create binary
	binary := &database.Binary{
		UserID:        "test-ver-get",
		Name:          "TestVerGet",
		Provider:      "github",
		ProviderPath:  "test/verget",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:verget",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installation
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test-1.0.0",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	// Set version
	err = verRepo.Set(binary.ID, installation.ID, "/usr/local/bin/test")
	if err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Get version
	version, err := verRepo.Get(binary.ID)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if version.BinaryID != binary.ID {
		t.Errorf("Expected BinaryID %d, got %d", binary.ID, version.BinaryID)
	}

	if version.InstallationID != installation.ID {
		t.Errorf("Expected InstallationID %d, got %d", installation.ID, version.InstallationID)
	}
}

func TestVersionsRepository_GetNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	verRepo := NewVersionsRepository(db)

	_, err := verRepo.Get(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestVersionsRepository_Unset(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)
	verRepo := NewVersionsRepository(db)

	// Create binary
	binary := &database.Binary{
		UserID:        "test-ver-unset",
		Name:          "TestVerUnset",
		Provider:      "github",
		ProviderPath:  "test/verunset",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:verunset",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installation
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test-1.0.0",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	// Set version
	err = verRepo.Set(binary.ID, installation.ID, "/usr/local/bin/test")
	if err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Unset version
	err = verRepo.Unset(binary.ID)
	if err != nil {
		t.Fatalf("Failed to unset version: %v", err)
	}

	// Verify unset
	_, err = verRepo.Get(binary.ID)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound after unset, got %v", err)
	}
}

func TestVersionsRepository_UnsetNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	verRepo := NewVersionsRepository(db)

	err := verRepo.Unset(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestVersionsRepository_Switch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)
	verRepo := NewVersionsRepository(db)

	// Create binary
	binary := &database.Binary{
		UserID:        "test-ver-switch",
		Name:          "TestVerSwitch",
		Provider:      "github",
		ProviderPath:  "test/verswitch",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:verswitch",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create two installations
	inst1 := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test-1.0.0",
		SourceURL:         "https://example.com/test-1.0.0.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	inst2 := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "2.0.0",
		InstalledPath:     "/usr/local/bin/test-2.0.0",
		SourceURL:         "https://example.com/test-2.0.0.tar.gz",
		FileSize:          2048,
		Checksum:          "def456",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(inst1)
	if err != nil {
		t.Fatalf("Failed to create installation 1: %v", err)
	}

	err = instRepo.Create(inst2)
	if err != nil {
		t.Fatalf("Failed to create installation 2: %v", err)
	}

	// Set initial version
	err = verRepo.Set(binary.ID, inst1.ID, "/usr/local/bin/test")
	if err != nil {
		t.Fatalf("Failed to set initial version: %v", err)
	}

	// Switch to second version
	err = verRepo.Switch(binary.ID, inst2.ID, "/usr/local/bin/test")
	if err != nil {
		t.Fatalf("Failed to switch version: %v", err)
	}

	// Verify switch
	version, err := verRepo.Get(binary.ID)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if version.InstallationID != inst2.ID {
		t.Errorf("Expected InstallationID %d after switch, got %d", inst2.ID, version.InstallationID)
	}
}

func TestVersionsRepository_GetWithInstallation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)
	verRepo := NewVersionsRepository(db)

	// Create binary
	binary := &database.Binary{
		UserID:        "test-ver-getwithinst",
		Name:          "TestVerGetWithInst",
		Provider:      "github",
		ProviderPath:  "test/vergetwithinst",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:vergetwithinst",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installation
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "3.0.0",
		InstalledPath:     "/usr/local/bin/test-3.0.0",
		SourceURL:         "https://example.com/test-3.0.0.tar.gz",
		FileSize:          3072,
		Checksum:          "ghi789",
		ChecksumAlgorithm: "sha512",
	}

	err = instRepo.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	// Set version
	symlinkPath := "/usr/local/bin/test"
	err = verRepo.Set(binary.ID, installation.ID, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Get with installation
	version, inst, err := verRepo.GetWithInstallation(binary.ID)
	if err != nil {
		t.Fatalf("Failed to get version with installation: %v", err)
	}

	// Verify version
	if version.BinaryID != binary.ID {
		t.Errorf("Expected BinaryID %d, got %d", binary.ID, version.BinaryID)
	}

	if version.InstallationID != installation.ID {
		t.Errorf("Expected InstallationID %d, got %d", installation.ID, version.InstallationID)
	}

	if version.SymlinkPath != symlinkPath {
		t.Errorf("Expected SymlinkPath %s, got %s", symlinkPath, version.SymlinkPath)
	}

	// Verify installation
	if inst.ID != installation.ID {
		t.Errorf("Expected installation ID %d, got %d", installation.ID, inst.ID)
	}

	if inst.Version != "3.0.0" {
		t.Errorf("Expected version 3.0.0, got %s", inst.Version)
	}

	if inst.InstalledPath != "/usr/local/bin/test-3.0.0" {
		t.Errorf("Expected InstalledPath /usr/local/bin/test-3.0.0, got %s", inst.InstalledPath)
	}

	if inst.SourceURL != "https://example.com/test-3.0.0.tar.gz" {
		t.Errorf("Expected SourceURL, got %s", inst.SourceURL)
	}

	if inst.FileSize != 3072 {
		t.Errorf("Expected FileSize 3072, got %d", inst.FileSize)
	}

	if inst.Checksum != "ghi789" {
		t.Errorf("Expected Checksum ghi789, got %s", inst.Checksum)
	}

	if inst.ChecksumAlgorithm != "sha512" {
		t.Errorf("Expected ChecksumAlgorithm sha512, got %s", inst.ChecksumAlgorithm)
	}
}

func TestVersionsRepository_GetWithInstallationNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	verRepo := NewVersionsRepository(db)

	_, _, err := verRepo.GetWithInstallation(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestVersionsRepository_CascadeDeleteOnInstallation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)
	verRepo := NewVersionsRepository(db)

	// Create binary
	binary := &database.Binary{
		UserID:        "test-ver-cascade",
		Name:          "TestVerCascade",
		Provider:      "github",
		ProviderPath:  "test/vercascade",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:vercascade",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installation
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		InstalledPath:     "/usr/local/bin/test-1.0.0",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}

	err = instRepo.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	// Set version
	err = verRepo.Set(binary.ID, installation.ID, "/usr/local/bin/test")
	if err != nil {
		t.Fatalf("Failed to set version: %v", err)
	}

	// Delete installation should cascade to version
	err = instRepo.Delete(installation.ID)
	if err != nil {
		t.Fatalf("Failed to delete installation: %v", err)
	}

	// Verify version was deleted
	_, err = verRepo.Get(binary.ID)
	if err != database.ErrNotFound {
		t.Errorf("Expected version to be cascade deleted, got %v", err)
	}
}
