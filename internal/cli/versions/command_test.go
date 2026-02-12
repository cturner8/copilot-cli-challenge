package versions

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

func setupVersionsDB(t *testing.T) (*repository.Service, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "versions.db")
	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	svc := repository.NewService(db)
	return svc, func() { _ = svc.Close() }
}

func TestVersionsCommand_ListWithActiveMarker(t *testing.T) {
	svc, cleanup := setupVersionsDB(t)
	defer cleanup()

	b := &database.Binary{
		UserID:       "gh",
		Name:         "gh",
		Provider:     "github",
		ProviderPath: "cli/cli",
		Format:       ".tar.gz",
	}
	if err := svc.Binaries.Create(b); err != nil {
		t.Fatalf("failed to create binary: %v", err)
	}

	inst := &database.Installation{
		BinaryID:          b.ID,
		Version:           "v1.2.3",
		InstalledPath:     "/tmp/gh",
		SourceURL:         "https://example.com/gh.tar.gz",
		FileSize:          1000,
		Checksum:          "abc123",
		ChecksumAlgorithm: "SHA256",
	}
	if err := svc.Installations.Create(inst); err != nil {
		t.Fatalf("failed to create installation: %v", err)
	}
	if err := svc.Versions.Set(b.ID, inst.ID, "/tmp/bin/gh"); err != nil {
		t.Fatalf("failed to set active version: %v", err)
	}

	DBService = svc
	Config = &config.Config{DateFormat: "2006-01-02"}

	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "gh"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Versions for gh:") || !strings.Contains(output, "* v1.2.3") {
		t.Fatalf("expected versions output with active marker, got: %s", output)
	}
}

func TestVersionsCommand_BinaryNotFound(t *testing.T) {
	svc, cleanup := setupVersionsDB(t)
	defer cleanup()

	DBService = svc
	Config = &config.Config{}

	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "missing"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "failed to list versions") {
		t.Fatalf("expected list versions error, got: %v", err)
	}
}
