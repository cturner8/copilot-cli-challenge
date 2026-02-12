package check

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

func setupCheckDB(t *testing.T) (*repository.Service, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "check.db")
	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	svc := repository.NewService(db)
	return svc, func() { _ = svc.Close() }
}

func TestCheckCommand_AllNoBinaries(t *testing.T) {
	svc, cleanup := setupCheckDB(t)
	defer cleanup()

	DBService = svc
	cmd := NewCommand()
	cmd.SetArgs([]string{"--all"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !strings.Contains(out.String(), "No binaries to check") {
		t.Fatalf("expected no binaries output, got: %s", out.String())
	}
}

func TestCheckCommand_SingleBinaryErrors(t *testing.T) {
	t.Run("binary not found", func(t *testing.T) {
		svc, cleanup := setupCheckDB(t)
		defer cleanup()

		DBService = svc
		cmd := NewCommand()
		cmd.SetArgs([]string{"--binary", "missing"})

		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "binary not found") {
			t.Fatalf("expected binary not found error, got: %v", err)
		}
	})

	t.Run("unsupported provider", func(t *testing.T) {
		svc, cleanup := setupCheckDB(t)
		defer cleanup()

		b := &database.Binary{
			UserID:       "custom",
			Name:         "custom",
			Provider:     "custom",
			ProviderPath: "owner/repo",
			Format:       ".zip",
		}
		if err := svc.Binaries.Create(b); err != nil {
			t.Fatalf("failed to create binary: %v", err)
		}

		DBService = svc
		cmd := NewCommand()
		cmd.SetArgs([]string{"--binary", "custom"})

		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "only github provider is currently supported") {
			t.Fatalf("expected unsupported provider error, got: %v", err)
		}
	})
}
