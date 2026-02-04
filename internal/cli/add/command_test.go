package add

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
		Binaries: []config.Binary{
			{
				Id:       "testbin",
				Name:     "testbin",
				Provider: "github",
				Path:     "owner/repo",
				Format:   ".tar.gz",
			},
		},
	}

	cleanup := func() {
		db.Close()
	}

	return dbService, cfg, cleanup
}

func TestAddCommand_URL(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	cmd := NewCommand()
	
	// Test with URL flag
	cmd.SetArgs([]string{"--url", "https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz"})
	
	// Capture output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "added successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}
}

func TestAddCommand_URLShorthand(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	cmd := NewCommand()
	
	// Test with -u shorthand
	cmd.SetArgs([]string{"-u", "https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "added successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}
}

func TestAddCommand_InvalidURL(t *testing.T) {
	dbService, cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	Config = cfg
	DBService = dbService

	cmd := NewCommand()
	cmd.SetArgs([]string{"--url", "not-a-github-url"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid URL, got none")
	}
}

func TestAddCommand_NoArgs(t *testing.T) {
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
		t.Error("Expected error when no args provided, got none")
	}
}

func TestAddCommand_Help(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"--help"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Help command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "Add a new binary") {
		t.Error("Help output missing expected text")
	}
	if !strings.Contains(output, "-u, --url") {
		t.Error("Help output missing flag information")
	}
}
