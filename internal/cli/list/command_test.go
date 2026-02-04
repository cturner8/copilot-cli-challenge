package list

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
		Version: 1,
		Binaries: []config.Binary{},
	}

	cleanup := func() {
		db.Close()
	}

	return dbService, cfg, cleanup
}

func TestListCommand_Empty(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	cmd := NewCommand()
	cmd.SetArgs([]string{})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "No binaries") {
		t.Errorf("Expected 'No binaries' message, got: %s", output)
	}
}

func TestListCommand_WithBinaries(t *testing.T) {
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
	cmd.SetArgs([]string{})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "testbin") {
		t.Errorf("Expected binary name in output, got: %s", output)
	}
	if !strings.Contains(output, "Binary") && !strings.Contains(output, "Active Version") {
		t.Errorf("Expected table headers in output, got: %s", output)
	}
}

func TestListCommand_BinaryFlag(t *testing.T) {
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
	if !strings.Contains(output, "Versions for testbin") {
		t.Errorf("Expected version list header, got: %s", output)
	}
}

func TestListCommand_BinaryShorthand(t *testing.T) {
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
	cmd.SetArgs([]string{"-b", "testbin"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "Versions for testbin") {
		t.Errorf("Expected version list header, got: %s", output)
	}
}

func TestListCommand_NonExistentBinary(t *testing.T) {
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

func TestListCommand_Help(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"--help"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Help command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "List") {
		t.Error("Help output missing expected text")
	}
	if !strings.Contains(output, "-b, --binary") {
		t.Error("Help output missing flag information")
	}
}
