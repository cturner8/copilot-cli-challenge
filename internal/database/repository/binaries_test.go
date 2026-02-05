package repository

import (
	"path/filepath"
	"testing"
	"time"

	"cturner8/binmate/internal/database"
)

// setupTestDB creates an in-memory database for testing
func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	return db
}

func TestBinariesRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	binary := &database.Binary{
		UserID:       "test-binary",
		Name:         "TestBinary",
		Provider:     "github",
		ProviderPath: "test/binary",
		Format:       ".tar.gz",
		ConfigDigest: "sha256:abc123",
		ConfigVersion: 1,
	}

	err := repo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	if binary.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	if binary.CreatedAt == 0 {
		t.Error("Expected CreatedAt to be set")
	}

	if binary.UpdatedAt == 0 {
		t.Error("Expected UpdatedAt to be set")
	}

	if binary.CreatedAt != binary.UpdatedAt {
		t.Error("Expected CreatedAt to equal UpdatedAt on creation")
	}
}

func TestBinariesRepository_CreateWithOptionalFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	alias := "test-alias"
	installPath := "/usr/local/bin"
	assetRegex := ".*linux.*"
	releaseRegex := "v.*"

	binary := &database.Binary{
		UserID:        "test-binary-optional",
		Name:          "TestBinaryOptional",
		Alias:         &alias,
		Provider:      "github",
		ProviderPath:  "test/binary",
		InstallPath:   &installPath,
		Format:        ".tar.gz",
		AssetRegex:    &assetRegex,
		ReleaseRegex:  &releaseRegex,
		ConfigDigest:  "sha256:def456",
		ConfigVersion: 1,
	}

	err := repo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary with optional fields: %v", err)
	}

	// Retrieve and verify
	retrieved, err := repo.Get(binary.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve binary: %v", err)
	}

	if retrieved.Alias == nil || *retrieved.Alias != alias {
		t.Errorf("Expected alias %s, got %v", alias, retrieved.Alias)
	}

	if retrieved.InstallPath == nil || *retrieved.InstallPath != installPath {
		t.Errorf("Expected install path %s, got %v", installPath, retrieved.InstallPath)
	}

	if retrieved.AssetRegex == nil || *retrieved.AssetRegex != assetRegex {
		t.Errorf("Expected asset regex %s, got %v", assetRegex, retrieved.AssetRegex)
	}

	if retrieved.ReleaseRegex == nil || *retrieved.ReleaseRegex != releaseRegex {
		t.Errorf("Expected release regex %s, got %v", releaseRegex, retrieved.ReleaseRegex)
	}
}

func TestBinariesRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	binary := &database.Binary{
		UserID:        "test-get",
		Name:          "TestGet",
		Provider:      "github",
		ProviderPath:  "test/get",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:get123",
		ConfigVersion: 1,
	}

	err := repo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	retrieved, err := repo.Get(binary.ID)
	if err != nil {
		t.Fatalf("Failed to get binary: %v", err)
	}

	if retrieved.ID != binary.ID {
		t.Errorf("Expected ID %d, got %d", binary.ID, retrieved.ID)
	}

	if retrieved.UserID != binary.UserID {
		t.Errorf("Expected UserID %s, got %s", binary.UserID, retrieved.UserID)
	}

	if retrieved.Name != binary.Name {
		t.Errorf("Expected Name %s, got %s", binary.Name, retrieved.Name)
	}
}

func TestBinariesRepository_GetNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	_, err := repo.Get(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBinariesRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	binary := &database.Binary{
		UserID:        "unique-user-id",
		Name:          "TestByUserID",
		Provider:      "github",
		ProviderPath:  "test/userid",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:userid123",
		ConfigVersion: 1,
	}

	err := repo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	retrieved, err := repo.GetByUserID("unique-user-id")
	if err != nil {
		t.Fatalf("Failed to get binary by user ID: %v", err)
	}

	if retrieved.UserID != "unique-user-id" {
		t.Errorf("Expected UserID unique-user-id, got %s", retrieved.UserID)
	}
}

func TestBinariesRepository_GetByUserIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	_, err := repo.GetByUserID("nonexistent")
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBinariesRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	binary := &database.Binary{
		UserID:        "test-name-id",
		Name:          "UniqueBinaryName",
		Provider:      "github",
		ProviderPath:  "test/name",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:name123",
		ConfigVersion: 1,
	}

	err := repo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	retrieved, err := repo.GetByName("UniqueBinaryName")
	if err != nil {
		t.Fatalf("Failed to get binary by name: %v", err)
	}

	if retrieved.Name != "UniqueBinaryName" {
		t.Errorf("Expected Name UniqueBinaryName, got %s", retrieved.Name)
	}
}

func TestBinariesRepository_GetByNameNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	_, err := repo.GetByName("NonexistentName")
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBinariesRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	binary := &database.Binary{
		UserID:        "test-update",
		Name:          "OriginalName",
		Provider:      "github",
		ProviderPath:  "test/update",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:original",
		ConfigVersion: 1,
	}

	err := repo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	originalUpdatedAt := binary.UpdatedAt

	// Wait a moment to ensure timestamp can change
	time.Sleep(10 * time.Millisecond)

	// Update the binary
	binary.Name = "UpdatedName"
	binary.ConfigDigest = "sha256:updated"
	binary.ConfigVersion = 2

	err = repo.Update(binary)
	if err != nil {
		t.Fatalf("Failed to update binary: %v", err)
	}

	if binary.UpdatedAt < originalUpdatedAt {
		t.Error("Expected UpdatedAt to be updated")
	}

	// Retrieve and verify
	retrieved, err := repo.Get(binary.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated binary: %v", err)
	}

	if retrieved.Name != "UpdatedName" {
		t.Errorf("Expected Name UpdatedName, got %s", retrieved.Name)
	}

	if retrieved.ConfigDigest != "sha256:updated" {
		t.Errorf("Expected ConfigDigest sha256:updated, got %s", retrieved.ConfigDigest)
	}

	if retrieved.ConfigVersion != 2 {
		t.Errorf("Expected ConfigVersion 2, got %d", retrieved.ConfigVersion)
	}
}

func TestBinariesRepository_UpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	binary := &database.Binary{
		ID:            999,
		UserID:        "test-update-notfound",
		Name:          "Test",
		Provider:      "github",
		ProviderPath:  "test/test",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:test",
		ConfigVersion: 1,
	}

	err := repo.Update(binary)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBinariesRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	// Create multiple binaries
	binaries := []*database.Binary{
		{
			UserID:        "test-list-1",
			Name:          "Binary1",
			Provider:      "github",
			ProviderPath:  "test/1",
			Format:        ".tar.gz",
			ConfigDigest:  "sha256:1",
			ConfigVersion: 1,
		},
		{
			UserID:        "test-list-2",
			Name:          "Binary2",
			Provider:      "github",
			ProviderPath:  "test/2",
			Format:        ".tar.gz",
			ConfigDigest:  "sha256:2",
			ConfigVersion: 1,
		},
		{
			UserID:        "test-list-3",
			Name:          "Binary3",
			Provider:      "github",
			ProviderPath:  "test/3",
			Format:        ".tar.gz",
			ConfigDigest:  "sha256:3",
			ConfigVersion: 1,
		},
	}

	for _, b := range binaries {
		err := repo.Create(b)
		if err != nil {
			t.Fatalf("Failed to create binary: %v", err)
		}
	}

	// List all binaries
	list, err := repo.List()
	if err != nil {
		t.Fatalf("Failed to list binaries: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 binaries, got %d", len(list))
	}

	// Verify they are sorted by name
	for i := 0; i < len(list)-1; i++ {
		if list[i].Name > list[i+1].Name {
			t.Error("Binaries are not sorted by name")
		}
	}
}

func TestBinariesRepository_ListEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	list, err := repo.List()
	if err != nil {
		t.Fatalf("Failed to list binaries: %v", err)
	}

	if len(list) != 0 {
		t.Errorf("Expected 0 binaries, got %d", len(list))
	}
}

func TestBinariesRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	binary := &database.Binary{
		UserID:        "test-delete",
		Name:          "ToDelete",
		Provider:      "github",
		ProviderPath:  "test/delete",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:delete",
		ConfigVersion: 1,
	}

	err := repo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	err = repo.Delete(binary.ID)
	if err != nil {
		t.Fatalf("Failed to delete binary: %v", err)
	}

	_, err = repo.Get(binary.ID)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}
}

func TestBinariesRepository_DeleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBinariesRepository(db)

	err := repo.Delete(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBinariesRepository_DeleteCascade(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)

	// Create binary
	binary := &database.Binary{
		UserID:        "test-cascade",
		Name:          "CascadeTest",
		Provider:      "github",
		ProviderPath:  "test/cascade",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:cascade",
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

	// Delete binary should cascade to installation
	err = binRepo.Delete(binary.ID)
	if err != nil {
		t.Fatalf("Failed to delete binary: %v", err)
	}

	// Verify installation was deleted
	_, err = instRepo.GetByID(installation.ID)
	if err != database.ErrNotFound {
		t.Errorf("Expected installation to be cascade deleted, got %v", err)
	}
}

func TestBinariesRepository_ListWithVersionDetails(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	instRepo := NewInstallationsRepository(db)
	verRepo := NewVersionsRepository(db)

	// Create binary
	binary := &database.Binary{
		UserID:        "test-version-details",
		Name:          "VersionDetailsTest",
		Provider:      "github",
		ProviderPath:  "test/versiondetails",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:versiondetails",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create installations
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
		Version:           "1.1.0",
		InstalledPath:     "/usr/local/bin/test-1.1.0",
		SourceURL:         "https://example.com/test-1.1.0.tar.gz",
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

	// Set active version
	err = verRepo.Set(binary.ID, inst2.ID, "/usr/local/bin/test")
	if err != nil {
		t.Fatalf("Failed to set active version: %v", err)
	}

	// List with version details
	details, err := binRepo.ListWithVersionDetails("No active version")
	if err != nil {
		t.Fatalf("Failed to list with version details: %v", err)
	}

	if len(details) != 1 {
		t.Fatalf("Expected 1 binary with details, got %d", len(details))
	}

	detail := details[0]

	if detail.Binary.ID != binary.ID {
		t.Errorf("Expected binary ID %d, got %d", binary.ID, detail.Binary.ID)
	}

	if detail.ActiveVersion != "1.1.0" {
		t.Errorf("Expected active version 1.1.0, got %s", detail.ActiveVersion)
	}

	if detail.InstallCount != 2 {
		t.Errorf("Expected install count 2, got %d", detail.InstallCount)
	}

	if detail.ActiveInstallation == nil {
		t.Fatal("Expected active installation to be populated")
	}

	if detail.ActiveInstallation.Version != "1.1.0" {
		t.Errorf("Expected active installation version 1.1.0, got %s", detail.ActiveInstallation.Version)
	}
}

func TestBinariesRepository_ListWithVersionDetailsNoActiveVersion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)

	// Create binary without any installations
	binary := &database.Binary{
		UserID:        "test-no-active",
		Name:          "NoActiveVersion",
		Provider:      "github",
		ProviderPath:  "test/noactive",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:noactive",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	details, err := binRepo.ListWithVersionDetails("No active version")
	if err != nil {
		t.Fatalf("Failed to list with version details: %v", err)
	}

	if len(details) != 1 {
		t.Fatalf("Expected 1 binary with details, got %d", len(details))
	}

	detail := details[0]

	if detail.ActiveVersion != "No active version" {
		t.Errorf("Expected 'No active version', got %s", detail.ActiveVersion)
	}

	if detail.InstallCount != 0 {
		t.Errorf("Expected install count 0, got %d", detail.InstallCount)
	}

	if detail.ActiveInstallation != nil {
		t.Error("Expected active installation to be nil")
	}
}
