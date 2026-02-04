package switchcmd

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

func TestSwitchCommand_NonExistent(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "nonexistent", "--version", "v1.0.0"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for non-existent binary, got none")
	}
}

func TestSwitchCommand_MissingBinaryFlag(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	cmd := NewCommand()
	cmd.SetArgs([]string{"--version", "v1.0.0"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when binary flag missing, got none")
	}
}

func TestSwitchCommand_MissingVersionFlag(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "testbin"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when version flag missing, got none")
	}
}

func TestSwitchCommand_Shorthands(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	cmd := NewCommand()
	cmd.SetArgs([]string{"-b", "testbin", "-v", "v1.0.0"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	// This will error because testbin doesn't exist, but it validates flag parsing
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error (binary not found), got none")
	}
}

func TestSwitchCommand_Help(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"--help"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Help command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "Switch") {
		t.Error("Help output missing expected text")
	}
	if !strings.Contains(output, "-b, --binary") {
		t.Error("Help output missing binary flag")
	}
	if !strings.Contains(output, "-v, --version") {
		t.Error("Help output missing version flag")
	}
}
