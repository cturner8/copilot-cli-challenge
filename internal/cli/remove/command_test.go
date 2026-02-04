package remove

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

func setupTestEnv(t *testing.T) (*repository.Service, *config.Config, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	dbService := repository.NewService(db)
	
	cfg := &config.Config{
		Version:  1,
		Binaries: []config.Binary{},
	}

	cleanup := func() {
		db.Close()
	}

	return dbService, cfg, cleanup
}

func TestRemoveCommand_Success(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	// Add a test binary
	binary := &database.Binary{
		UserID:       "testbin",
		Name:         "testbin",
		Provider:     "github",
		ProviderPath: "owner/repo",
		Format:       ".tar.gz",
	}
	if err := dbService.Binaries.Create(binary); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "testbin"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "removed") {
		t.Errorf("Expected success message, got: %s", output)
	}
}

func TestRemoveCommand_WithFiles(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	// Add a test binary
	binary := &database.Binary{
		UserID:       "testbin",
		Name:         "testbin",
		Provider:     "github",
		ProviderPath: "owner/repo",
		Format:       ".tar.gz",
	}
	if err := dbService.Binaries.Create(binary); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	cmd := NewCommand()
	cmd.SetArgs([]string{"-b", "testbin", "-f"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "including files") {
		t.Errorf("Expected 'including files' message, got: %s", output)
	}
}

func TestRemoveCommand_NonExistent(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "nonexistent"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for non-existent binary, got none")
	}
}

func TestRemoveCommand_MissingBinaryFlag(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	cmd := NewCommand()
	cmd.SetArgs([]string{})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when binary flag missing, got none")
	}
}

func TestRemoveCommand_Help(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"--help"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Help command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "Remove a binary") {
		t.Error("Help output missing expected text")
	}
	if !strings.Contains(output, "-b, --binary") {
		t.Error("Help output missing binary flag")
	}
	if !strings.Contains(output, "-f, --files") {
		t.Error("Help output missing files flag")
	}
}
