package root

import (
	"path/filepath"
	"strings"
	"testing"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

func setupRootDB(t *testing.T) (*repository.Service, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "root.db")
	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	svc := repository.NewService(db)
	return svc, func() { _ = svc.Close() }
}

func TestRootCommand_PreRunESync(t *testing.T) {
	svc, cleanup := setupRootDB(t)
	defer cleanup()

	Config = &config.Config{
		Version: 1,
		Binaries: []config.Binary{
			{Id: "gh", Name: "gh", Provider: "github", Path: "cli/cli", Format: ".tar.gz"},
		},
	}
	DBService = svc

	cmd := NewCommand()
	if cmd.Use != "binmate" {
		t.Fatalf("expected use binmate, got %q", cmd.Use)
	}

	if err := cmd.PreRunE(cmd, []string{}); err != nil {
		t.Fatalf("expected pre-run sync success, got: %v", err)
	}
	if _, err := svc.Binaries.GetByUserID("gh"); err != nil {
		t.Fatalf("expected synced binary in DB, got: %v", err)
	}
}

func TestRootCommand_PreRunESyncError(t *testing.T) {
	svc, cleanup := setupRootDB(t)
	defer cleanup()

	Config = &config.Config{
		Version: 1,
		Binaries: []config.Binary{
			{Id: "bad", Name: "bad", Provider: "github", Path: "owner/repo", Format: ".tar.gz"},
		},
	}
	DBService = svc
	_ = svc.Close()

	cmd := NewCommand()
	err := cmd.PreRunE(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "sync error") {
		t.Fatalf("expected sync error, got: %v", err)
	}
}
