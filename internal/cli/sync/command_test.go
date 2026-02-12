package sync

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

func setupSyncDB(t *testing.T) (*repository.Service, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "sync.db")
	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	svc := repository.NewService(db)
	return svc, func() { _ = svc.Close() }
}

func TestSyncCommand_Success(t *testing.T) {
	svc, cleanup := setupSyncDB(t)
	defer cleanup()

	DBService = svc
	Config = &config.Config{
		Version: 1,
		Binaries: []config.Binary{
			{Id: "gh", Name: "gh", Provider: "github", Path: "cli/cli", Format: ".tar.gz"},
		},
	}

	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !strings.Contains(out.String(), "Sync complete") {
		t.Fatalf("expected sync output, got: %s", out.String())
	}
	if _, err := svc.Binaries.GetByUserID("gh"); err != nil {
		t.Fatalf("expected synced binary in DB, got: %v", err)
	}
}

func TestSyncCommand_LogStartError(t *testing.T) {
	svc, cleanup := setupSyncDB(t)
	defer cleanup()

	DBService = svc
	Config = &config.Config{Version: 1}
	_ = svc.Close()

	cmd := NewCommand()
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "sync start error") {
		t.Fatalf("expected sync start error, got: %v", err)
	}
}
