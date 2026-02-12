package repository

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"cturner8/binmate/internal/database"
)

func setupRepositoryService(t *testing.T) *Service {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "repository.db")
	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	svc := NewService(db)
	t.Cleanup(func() {
		_ = svc.Close()
	})
	return svc
}

func createRepoBinary(t *testing.T, svc *Service, userID, name string) *database.Binary {
	t.Helper()
	b := &database.Binary{
		UserID:       userID,
		Name:         name,
		Provider:     "github",
		ProviderPath: "owner/repo",
		Format:       ".tar.gz",
	}
	if err := svc.Binaries.Create(b); err != nil {
		t.Fatalf("failed to create binary: %v", err)
	}
	return b
}

func createRepoInstallation(t *testing.T, svc *Service, binaryID int64, version string) *database.Installation {
	t.Helper()
	inst := &database.Installation{
		BinaryID:          binaryID,
		Version:           version,
		InstalledPath:     "/tmp/" + version,
		SourceURL:         "https://example.com/" + version + ".tar.gz",
		FileSize:          1234,
		Checksum:          "sum-" + version,
		ChecksumAlgorithm: "SHA256",
	}
	if err := svc.Installations.Create(inst); err != nil {
		t.Fatalf("failed to create installation: %v", err)
	}
	return inst
}

func TestServiceClose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "service-close.db")
	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	svc := NewService(db)
	if err := svc.Close(); err != nil {
		t.Fatalf("expected service close to succeed, got: %v", err)
	}
}

func TestBinariesRepositoryCRUD(t *testing.T) {
	svc := setupRepositoryService(t)

	binary := &database.Binary{
		UserID:       "gh",
		Name:         "gh",
		Provider:     "github",
		ProviderPath: "cli/cli",
		Format:       ".tar.gz",
	}
	if err := svc.Binaries.Create(binary); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if binary.Source != "manual" {
		t.Fatalf("expected default source manual, got %q", binary.Source)
	}

	gotByID, err := svc.Binaries.Get(binary.ID)
	if err != nil {
		t.Fatalf("get by ID failed: %v", err)
	}
	if gotByID.UserID != "gh" {
		t.Fatalf("expected user_id gh, got %q", gotByID.UserID)
	}

	gotByUserID, err := svc.Binaries.GetByUserID("gh")
	if err != nil {
		t.Fatalf("get by user ID failed: %v", err)
	}
	gotByName, err := svc.Binaries.GetByName("gh")
	if err != nil {
		t.Fatalf("get by name failed: %v", err)
	}
	if gotByUserID.ID != gotByName.ID {
		t.Fatalf("expected same binary from user ID and name lookups")
	}

	list, err := svc.Binaries.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 binary, got %d", len(list))
	}

	gotByID.Name = "gh-updated"
	gotByID.Authenticated = true
	if err := svc.Binaries.Update(gotByID); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, err := svc.Binaries.Get(gotByID.ID)
	if err != nil {
		t.Fatalf("get after update failed: %v", err)
	}
	if updated.Name != "gh-updated" || !updated.Authenticated {
		t.Fatalf("expected updated binary fields, got name=%q auth=%v", updated.Name, updated.Authenticated)
	}

	if err := svc.Binaries.Delete(updated.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := svc.Binaries.Get(updated.ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got: %v", err)
	}
	if err := svc.Binaries.Delete(updated.ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on second delete, got: %v", err)
	}
}

func TestBinariesRepositorySyncAndVersionDetails(t *testing.T) {
	svc := setupRepositoryService(t)

	manual := &database.Binary{
		UserID:       "manual-bin",
		Name:         "manual",
		Provider:     "github",
		ProviderPath: "owner/manual",
		Format:       ".tar.gz",
		Source:       "manual",
	}
	if err := svc.Binaries.Create(manual); err != nil {
		t.Fatalf("failed to create manual binary: %v", err)
	}

	initialConfig := []ConfigBinary{
		{ID: "cfg-1", Name: "cfg-one", Provider: "github", Path: "owner/cfg1", Format: ".tar.gz"},
		{ID: "cfg-2", Name: "cfg-two", Provider: "github", Path: "owner/cfg2", Format: ".zip"},
	}
	if err := svc.Binaries.SyncFromConfig(initialConfig, 1); err != nil {
		t.Fatalf("initial sync failed: %v", err)
	}

	updatedConfig := []ConfigBinary{
		{ID: "cfg-1", Name: "cfg-one-updated", Provider: "github", Path: "owner/cfg1", Format: ".tar.gz", Authenticated: true},
	}
	if err := svc.Binaries.SyncFromConfig(updatedConfig, 2); err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	cfg1, err := svc.Binaries.GetByUserID("cfg-1")
	if err != nil {
		t.Fatalf("expected cfg-1 to exist: %v", err)
	}
	if cfg1.Name != "cfg-one-updated" || cfg1.ConfigVersion != 2 || !cfg1.Authenticated {
		t.Fatalf("cfg-1 was not updated correctly")
	}
	if _, err := svc.Binaries.GetByUserID("cfg-2"); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected cfg-2 removal, got: %v", err)
	}
	if _, err := svc.Binaries.GetByUserID("manual-bin"); err != nil {
		t.Fatalf("expected manual binary retention, got: %v", err)
	}

	// Cover SyncBinary create/update/no-op paths.
	cfg3 := ConfigBinary{ID: "cfg-3", Name: "cfg-three", Provider: "github", Path: "owner/cfg3", Format: ".zip"}
	if err := svc.Binaries.SyncBinary(cfg3, 1); err != nil {
		t.Fatalf("sync binary create failed: %v", err)
	}
	if err := svc.Binaries.SyncBinary(cfg3, 1); err != nil {
		t.Fatalf("sync binary no-op failed: %v", err)
	}
	cfg3.Name = "cfg-three-updated"
	if err := svc.Binaries.SyncBinary(cfg3, 3); err != nil {
		t.Fatalf("sync binary update failed: %v", err)
	}

	inst := createRepoInstallation(t, svc, cfg1.ID, "v1.0.0")
	if err := svc.Versions.Set(cfg1.ID, inst.ID, "/tmp/cfg1-link"); err != nil {
		t.Fatalf("set active version failed: %v", err)
	}

	details, err := svc.Binaries.ListWithVersionDetails("No active version")
	if err != nil {
		t.Fatalf("list with version details failed: %v", err)
	}
	if len(details) < 2 {
		t.Fatalf("expected at least 2 binaries in details, got %d", len(details))
	}

	var sawActive bool
	for _, d := range details {
		if d.Binary.UserID == "cfg-1" {
			sawActive = true
			if d.ActiveVersion != "v1.0.0" || d.ActiveInstallation == nil {
				t.Fatalf("expected active installation details for cfg-1")
			}
		}
	}
	if !sawActive {
		t.Fatalf("expected cfg-1 in version details")
	}
}

func TestInstallationsRepositoryMethods(t *testing.T) {
	svc := setupRepositoryService(t)
	binary := createRepoBinary(t, svc, "inst-bin", "instbin")

	inst1 := createRepoInstallation(t, svc, binary.ID, "v1.0.0")
	inst2 := createRepoInstallation(t, svc, binary.ID, "v2.0.0")

	// Ensure deterministic latest ordering.
	_, _ = svc.DB.Exec(`UPDATE installations SET installed_at = ? WHERE id = ?`, 1, inst1.ID)
	_, _ = svc.DB.Exec(`UPDATE installations SET installed_at = ? WHERE id = ?`, 2, inst2.ID)

	if _, err := svc.Installations.Get(binary.ID, "v1.0.0"); err != nil {
		t.Fatalf("get by binary/version failed: %v", err)
	}
	if _, err := svc.Installations.GetByID(inst1.ID); err != nil {
		t.Fatalf("get by ID failed: %v", err)
	}

	list, err := svc.Installations.ListByBinary(binary.ID)
	if err != nil {
		t.Fatalf("list by binary failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 installations, got %d", len(list))
	}

	latest, err := svc.Installations.GetLatest(binary.ID)
	if err != nil {
		t.Fatalf("get latest failed: %v", err)
	}
	if latest.Version != "v2.0.0" {
		t.Fatalf("expected v2.0.0 latest, got %s", latest.Version)
	}

	if err := svc.Installations.Delete(inst1.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := svc.Installations.Delete(inst1.ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on second delete, got: %v", err)
	}
}

func TestVersionsRepositoryMethods(t *testing.T) {
	svc := setupRepositoryService(t)
	binary := createRepoBinary(t, svc, "ver-bin", "verbin")
	inst1 := createRepoInstallation(t, svc, binary.ID, "v1.0.0")
	inst2 := createRepoInstallation(t, svc, binary.ID, "v2.0.0")

	if err := svc.Versions.Set(binary.ID, inst1.ID, "/tmp/verbin-v1"); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	got, err := svc.Versions.Get(binary.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.InstallationID != inst1.ID {
		t.Fatalf("expected installation %d, got %d", inst1.ID, got.InstallationID)
	}

	if err := svc.Versions.Switch(binary.ID, inst2.ID, "/tmp/verbin-v2"); err != nil {
		t.Fatalf("switch failed: %v", err)
	}
	version, installation, err := svc.Versions.GetWithInstallation(binary.ID)
	if err != nil {
		t.Fatalf("get with installation failed: %v", err)
	}
	if version.InstallationID != inst2.ID || installation.Version != "v2.0.0" {
		t.Fatalf("expected switched version data")
	}

	if err := svc.Versions.Unset(binary.ID); err != nil {
		t.Fatalf("unset failed: %v", err)
	}
	if _, err := svc.Versions.Get(binary.ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after unset, got: %v", err)
	}
	if _, _, err := svc.Versions.GetWithInstallation(binary.ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for get-with-installation after unset, got: %v", err)
	}
	if err := svc.Versions.Unset(binary.ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on second unset, got: %v", err)
	}
}

func TestDownloadsRepositoryMethods(t *testing.T) {
	svc := setupRepositoryService(t)
	binary := createRepoBinary(t, svc, "dl-bin", "dlbin")

	dl := &database.Download{
		BinaryID:          binary.ID,
		Version:           "v1.0.0",
		CachePath:         "/tmp/cache-v1",
		SourceURL:         "https://example.com/v1.tar.gz",
		FileSize:          100,
		Checksum:          "checksum-v1",
		ChecksumAlgorithm: "SHA256",
		IsComplete:        false,
	}
	if err := svc.Downloads.Create(dl); err != nil {
		t.Fatalf("create download failed: %v", err)
	}

	got, err := svc.Downloads.Get(binary.ID, "v1.0.0")
	if err != nil {
		t.Fatalf("get download failed: %v", err)
	}
	if got.IsComplete {
		t.Fatalf("expected initial incomplete download")
	}

	incomplete, err := svc.Downloads.GetIncomplete()
	if err != nil {
		t.Fatalf("get incomplete failed: %v", err)
	}
	if len(incomplete) == 0 {
		t.Fatalf("expected at least one incomplete download")
	}

	beforeAccess := got.LastAccessedAt
	time.Sleep(10 * time.Millisecond)
	if err := svc.Downloads.UpdateLastAccessed(dl.ID); err != nil {
		t.Fatalf("update last accessed failed: %v", err)
	}
	updated, err := svc.Downloads.Get(binary.ID, "v1.0.0")
	if err != nil {
		t.Fatalf("get after access update failed: %v", err)
	}
	if updated.LastAccessedAt < beforeAccess {
		t.Fatalf("expected last_accessed_at to be updated")
	}

	if err := svc.Downloads.MarkComplete(dl.ID); err != nil {
		t.Fatalf("mark complete failed: %v", err)
	}
	complete, err := svc.Downloads.Get(binary.ID, "v1.0.0")
	if err != nil {
		t.Fatalf("get after mark complete failed: %v", err)
	}
	if !complete.IsComplete {
		t.Fatalf("expected completed download")
	}

	byBinary, err := svc.Downloads.ListByBinary(binary.ID)
	if err != nil {
		t.Fatalf("list by binary failed: %v", err)
	}
	if len(byBinary) != 1 {
		t.Fatalf("expected 1 binary download, got %d", len(byBinary))
	}

	forCleanup, err := svc.Downloads.ListForCleanup(time.Now().Unix()+1000, 10)
	if err != nil {
		t.Fatalf("list for cleanup failed: %v", err)
	}
	if len(forCleanup) == 0 {
		t.Fatalf("expected cleanup candidate")
	}

	if err := svc.Downloads.Delete(dl.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := svc.Downloads.Delete(dl.ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on second delete, got: %v", err)
	}
	if _, err := svc.Downloads.Get(binary.ID, "v1.0.0"); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got: %v", err)
	}
	if err := svc.Downloads.UpdateLastAccessed(dl.ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound update last accessed, got: %v", err)
	}
	if err := svc.Downloads.MarkComplete(dl.ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound mark complete, got: %v", err)
	}
}

func TestLogsRepositoryMethods(t *testing.T) {
	svc := setupRepositoryService(t)

	startID, err := svc.Logs.LogStart("install", "", "", "starting install")
	if err != nil {
		t.Fatalf("log start failed: %v", err)
	}
	if err := svc.Logs.LogEntity(startID, "binary", "gh"); err != nil {
		t.Fatalf("log entity failed: %v", err)
	}
	if err := svc.Logs.LogSuccess(startID, 100); err != nil {
		t.Fatalf("log success failed: %v", err)
	}

	failID, err := svc.Logs.LogStart("sync", "binary", "gh", "starting sync")
	if err != nil {
		t.Fatalf("second log start failed: %v", err)
	}
	if err := svc.Logs.LogFailure(failID, "boom", 50); err != nil {
		t.Fatalf("log failure failed: %v", err)
	}

	recent, err := svc.Logs.GetRecent(10)
	if err != nil || len(recent) < 2 {
		t.Fatalf("expected recent logs, got len=%d err=%v", len(recent), err)
	}

	byType, err := svc.Logs.GetByType("install", 10)
	if err != nil || len(byType) == 0 {
		t.Fatalf("expected logs by type, got len=%d err=%v", len(byType), err)
	}

	failures, err := svc.Logs.GetFailures(10)
	if err != nil || len(failures) == 0 {
		t.Fatalf("expected failures, got len=%d err=%v", len(failures), err)
	}
	if failures[0].OperationStatus != "failed" {
		t.Fatalf("expected failed operation status, got %q", failures[0].OperationStatus)
	}

	byEntity, err := svc.Logs.GetByEntity("binary", "gh", 10)
	if err != nil || len(byEntity) == 0 {
		t.Fatalf("expected logs by entity, got len=%d err=%v", len(byEntity), err)
	}
}
