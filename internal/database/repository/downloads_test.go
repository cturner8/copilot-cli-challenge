package repository

import (
	"testing"
	"time"

	"cturner8/binmate/internal/database"
)

func TestDownloadsRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	dlRepo := NewDownloadsRepository(db)

	// Create binary
	binary := &database.Binary{
		UserID:        "test-dl",
		Name:          "TestDL",
		Provider:      "github",
		ProviderPath:  "test/dl",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:dl",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	download := &database.Download{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		CachePath:         "/tmp/cache/test-1.0.0.tar.gz",
		SourceURL:         "https://example.com/test-1.0.0.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
		IsComplete:        true,
	}

	err = dlRepo.Create(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}

	if download.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	if download.DownloadedAt == 0 {
		t.Error("Expected DownloadedAt to be set")
	}

	if download.LastAccessedAt == 0 {
		t.Error("Expected LastAccessedAt to be set")
	}

	if download.DownloadedAt != download.LastAccessedAt {
		t.Error("Expected DownloadedAt to equal LastAccessedAt on creation")
	}
}

func TestDownloadsRepository_CreateForeignKeyConstraint(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dlRepo := NewDownloadsRepository(db)

	// Try to create download with non-existent binary ID
	download := &database.Download{
		BinaryID:          999,
		Version:           "1.0.0",
		CachePath:         "/tmp/cache/test.tar.gz",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
		IsComplete:        true,
	}

	err := dlRepo.Create(download)
	if err == nil {
		t.Error("Expected foreign key constraint error, got nil")
	}
}

func TestDownloadsRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	dlRepo := NewDownloadsRepository(db)

	binary := &database.Binary{
		UserID:        "test-get-dl",
		Name:          "TestGetDL",
		Provider:      "github",
		ProviderPath:  "test/getdl",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:getdl",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	download := &database.Download{
		BinaryID:          binary.ID,
		Version:           "2.0.0",
		CachePath:         "/tmp/cache/test-2.0.0.tar.gz",
		SourceURL:         "https://example.com/test-2.0.0.tar.gz",
		FileSize:          2048,
		Checksum:          "def456",
		ChecksumAlgorithm: "sha256",
		IsComplete:        false,
	}

	err = dlRepo.Create(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}

	retrieved, err := dlRepo.Get(binary.ID, "2.0.0")
	if err != nil {
		t.Fatalf("Failed to get download: %v", err)
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

	if retrieved.IsComplete {
		t.Error("Expected IsComplete to be false")
	}
}

func TestDownloadsRepository_GetNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dlRepo := NewDownloadsRepository(db)

	_, err := dlRepo.Get(999, "1.0.0")
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestDownloadsRepository_UpdateLastAccessed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	dlRepo := NewDownloadsRepository(db)

	binary := &database.Binary{
		UserID:        "test-update-accessed-dl",
		Name:          "TestUpdateAccessedDL",
		Provider:      "github",
		ProviderPath:  "test/updateaccesseddl",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:updateaccesseddl",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

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

	err = dlRepo.Create(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}

	originalLastAccessed := download.LastAccessedAt

	// Wait a moment to ensure timestamp changes
	time.Sleep(1100 * time.Millisecond)

	err = dlRepo.UpdateLastAccessed(download.ID)
	if err != nil {
		t.Fatalf("Failed to update last accessed: %v", err)
	}

	// Retrieve and verify
	retrieved, err := dlRepo.Get(binary.ID, "1.0.0")
	if err != nil {
		t.Fatalf("Failed to get download: %v", err)
	}

	if retrieved.LastAccessedAt < originalLastAccessed {
		t.Errorf("Expected LastAccessedAt to be updated, got %d <= %d", retrieved.LastAccessedAt, originalLastAccessed)
	}
}

func TestDownloadsRepository_UpdateLastAccessedNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dlRepo := NewDownloadsRepository(db)

	err := dlRepo.UpdateLastAccessed(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestDownloadsRepository_MarkComplete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	dlRepo := NewDownloadsRepository(db)

	binary := &database.Binary{
		UserID:        "test-mark-complete-dl",
		Name:          "TestMarkCompleteDL",
		Provider:      "github",
		ProviderPath:  "test/markcompletedl",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:markcompletedl",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	download := &database.Download{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		CachePath:         "/tmp/cache/test.tar.gz",
		SourceURL:         "https://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
		IsComplete:        false,
	}

	err = dlRepo.Create(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}

	err = dlRepo.MarkComplete(download.ID)
	if err != nil {
		t.Fatalf("Failed to mark download complete: %v", err)
	}

	// Retrieve and verify
	retrieved, err := dlRepo.Get(binary.ID, "1.0.0")
	if err != nil {
		t.Fatalf("Failed to get download: %v", err)
	}

	if !retrieved.IsComplete {
		t.Error("Expected IsComplete to be true")
	}
}

func TestDownloadsRepository_MarkCompleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dlRepo := NewDownloadsRepository(db)

	err := dlRepo.MarkComplete(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestDownloadsRepository_ListForCleanup(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	dlRepo := NewDownloadsRepository(db)

	binary := &database.Binary{
		UserID:        "test-cleanup-dl",
		Name:          "TestCleanupDL",
		Provider:      "github",
		ProviderPath:  "test/cleanupdl",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:cleanupdl",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create old download
	oldDownload := &database.Download{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		CachePath:         "/tmp/cache/test-old.tar.gz",
		SourceURL:         "https://example.com/test-old.tar.gz",
		FileSize:          1024,
		Checksum:          "old123",
		ChecksumAlgorithm: "sha256",
		IsComplete:        true,
	}

	err = dlRepo.Create(oldDownload)
	if err != nil {
		t.Fatalf("Failed to create old download: %v", err)
	}

	// Manually set old last accessed time
	_, err = db.Exec("UPDATE downloads SET last_accessed_at = ? WHERE id = ?",
		time.Now().Add(-48*time.Hour).Unix(), oldDownload.ID)
	if err != nil {
		t.Fatalf("Failed to update last accessed time: %v", err)
	}

	// Create recent download
	recentDownload := &database.Download{
		BinaryID:          binary.ID,
		Version:           "2.0.0",
		CachePath:         "/tmp/cache/test-recent.tar.gz",
		SourceURL:         "https://example.com/test-recent.tar.gz",
		FileSize:          2048,
		Checksum:          "recent456",
		ChecksumAlgorithm: "sha256",
		IsComplete:        true,
	}

	err = dlRepo.Create(recentDownload)
	if err != nil {
		t.Fatalf("Failed to create recent download: %v", err)
	}

	// List for cleanup with cutoff time
	cutoffTime := time.Now().Add(-24 * time.Hour).Unix()
	list, err := dlRepo.ListForCleanup(cutoffTime, 10)
	if err != nil {
		t.Fatalf("Failed to list for cleanup: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("Expected 1 download for cleanup, got %d", len(list))
	}

	if list[0].ID != oldDownload.ID {
		t.Errorf("Expected old download ID %d, got %d", oldDownload.ID, list[0].ID)
	}
}

func TestDownloadsRepository_ListForCleanupEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dlRepo := NewDownloadsRepository(db)

	// List with future cutoff time
	cutoffTime := time.Now().Add(24 * time.Hour).Unix()
	list, err := dlRepo.ListForCleanup(cutoffTime, 10)
	if err != nil {
		t.Fatalf("Failed to list for cleanup: %v", err)
	}

	if len(list) != 0 {
		t.Errorf("Expected 0 downloads for cleanup, got %d", len(list))
	}
}

func TestDownloadsRepository_GetIncomplete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	dlRepo := NewDownloadsRepository(db)

	binary := &database.Binary{
		UserID:        "test-incomplete-dl",
		Name:          "TestIncompleteDL",
		Provider:      "github",
		ProviderPath:  "test/incompletedl",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:incompletedl",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create incomplete download
	incompleteDownload := &database.Download{
		BinaryID:          binary.ID,
		Version:           "1.0.0",
		CachePath:         "/tmp/cache/test-incomplete.tar.gz",
		SourceURL:         "https://example.com/test-incomplete.tar.gz",
		FileSize:          1024,
		Checksum:          "incomplete123",
		ChecksumAlgorithm: "sha256",
		IsComplete:        false,
	}

	err = dlRepo.Create(incompleteDownload)
	if err != nil {
		t.Fatalf("Failed to create incomplete download: %v", err)
	}

	// Create complete download
	completeDownload := &database.Download{
		BinaryID:          binary.ID,
		Version:           "2.0.0",
		CachePath:         "/tmp/cache/test-complete.tar.gz",
		SourceURL:         "https://example.com/test-complete.tar.gz",
		FileSize:          2048,
		Checksum:          "complete456",
		ChecksumAlgorithm: "sha256",
		IsComplete:        true,
	}

	err = dlRepo.Create(completeDownload)
	if err != nil {
		t.Fatalf("Failed to create complete download: %v", err)
	}

	// Get incomplete downloads
	list, err := dlRepo.GetIncomplete()
	if err != nil {
		t.Fatalf("Failed to get incomplete downloads: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("Expected 1 incomplete download, got %d", len(list))
	}

	if list[0].ID != incompleteDownload.ID {
		t.Errorf("Expected incomplete download ID %d, got %d", incompleteDownload.ID, list[0].ID)
	}

	if list[0].IsComplete {
		t.Error("Expected IsComplete to be false")
	}
}

func TestDownloadsRepository_GetIncompleteEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dlRepo := NewDownloadsRepository(db)

	list, err := dlRepo.GetIncomplete()
	if err != nil {
		t.Fatalf("Failed to get incomplete downloads: %v", err)
	}

	if len(list) != 0 {
		t.Errorf("Expected 0 incomplete downloads, got %d", len(list))
	}
}

func TestDownloadsRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	dlRepo := NewDownloadsRepository(db)

	binary := &database.Binary{
		UserID:        "test-delete-dl",
		Name:          "TestDeleteDL",
		Provider:      "github",
		ProviderPath:  "test/deletedl",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:deletedl",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

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

	err = dlRepo.Create(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}

	err = dlRepo.Delete(download.ID)
	if err != nil {
		t.Fatalf("Failed to delete download: %v", err)
	}

	_, err = dlRepo.Get(binary.ID, "1.0.0")
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}
}

func TestDownloadsRepository_DeleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dlRepo := NewDownloadsRepository(db)

	err := dlRepo.Delete(999)
	if err != database.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestDownloadsRepository_ListByBinary(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	dlRepo := NewDownloadsRepository(db)

	binary := &database.Binary{
		UserID:        "test-list-binary-dl",
		Name:          "TestListBinaryDL",
		Provider:      "github",
		ProviderPath:  "test/listbinarydl",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:listbinarydl",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create multiple downloads
	versions := []string{"1.0.0", "1.1.0", "1.2.0"}
	for _, version := range versions {
		download := &database.Download{
			BinaryID:          binary.ID,
			Version:           version,
			CachePath:         "/tmp/cache/test-" + version + ".tar.gz",
			SourceURL:         "https://example.com/test-" + version + ".tar.gz",
			FileSize:          1024,
			Checksum:          "checksum-" + version,
			ChecksumAlgorithm: "sha256",
			IsComplete:        true,
		}

		err = dlRepo.Create(download)
		if err != nil {
			t.Fatalf("Failed to create download %s: %v", version, err)
		}
	}

	list, err := dlRepo.ListByBinary(binary.ID)
	if err != nil {
		t.Fatalf("Failed to list downloads by binary: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 downloads, got %d", len(list))
	}

	// Verify ordered by downloaded_at DESC
	for i := 0; i < len(list)-1; i++ {
		if list[i].DownloadedAt < list[i+1].DownloadedAt {
			t.Error("Downloads are not sorted by DownloadedAt DESC")
		}
	}
}

func TestDownloadsRepository_ListByBinaryEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	binRepo := NewBinariesRepository(db)
	dlRepo := NewDownloadsRepository(db)

	binary := &database.Binary{
		UserID:        "test-list-empty-binary-dl",
		Name:          "TestListEmptyBinaryDL",
		Provider:      "github",
		ProviderPath:  "test/listemptybinarydl",
		Format:        ".tar.gz",
		ConfigDigest:  "sha256:listemptybinarydl",
		ConfigVersion: 1,
	}

	err := binRepo.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	list, err := dlRepo.ListByBinary(binary.ID)
	if err != nil {
		t.Fatalf("Failed to list downloads by binary: %v", err)
	}

	if len(list) != 0 {
		t.Errorf("Expected 0 downloads, got %d", len(list))
	}
}
