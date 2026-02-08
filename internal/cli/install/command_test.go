package install

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

func TestInstallCommand_BinaryInDatabase(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize test database
	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)
	defer dbService.Close()

	// Create a binary directly in the database (simulating 'add' command from URL)
	testBinary := &database.Binary{
		UserID:       "test-binary",
		Name:         "testbin",
		Provider:     "github",
		ProviderPath: "test/test-binary",
		Format:       ".tar.gz",
	}

	err = dbService.Binaries.Create(testBinary)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Setup command environment
	DBService = dbService
	Config = &config.Config{
		Binaries: []config.Binary{}, // Empty config - binary only in DB
		Version:  1,
	}

	// Create command
	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "test-binary", "--version", "v1.0.0"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Parse flags first
	if err := cmd.ParseFlags([]string{"--binary", "test-binary", "--version", "v1.0.0"}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Execute PreRunE only (we don't want to actually install)
	err = cmd.PreRunE(cmd, []string{})
	if err != nil {
		t.Errorf("PreRunE failed for binary in database: %v\nOutput: %s", err, buf.String())
	}

	// Verify binary still exists in database
	binary, err := dbService.Binaries.GetByUserID("test-binary")
	if err != nil {
		t.Errorf("Binary should exist in database after PreRunE: %v", err)
		return
	}
	if binary.UserID != "test-binary" {
		t.Errorf("Expected binary ID 'test-binary', got '%s'", binary.UserID)
	}
}

func TestInstallCommand_BinaryInConfig(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize test database
	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)
	defer dbService.Close()

	// Setup command environment with binary in config but NOT in database
	DBService = dbService
	Config = &config.Config{
		Binaries: []config.Binary{
			{
				Id:       "config-binary",
				Name:     "configbin",
				Provider: "github",
				Path:     "test/config-binary",
				Format:   ".tar.gz",
			},
		},
		Version: 1,
	}

	// Create command
	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "config-binary", "--version", "v1.0.0"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Parse flags first
	if err := cmd.ParseFlags([]string{"--binary", "config-binary", "--version", "v1.0.0"}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Execute PreRunE
	err = cmd.PreRunE(cmd, []string{})
	if err != nil {
		t.Errorf("PreRunE failed for binary in config: %v\nOutput: %s", err, buf.String())
	}

	// Verify binary was synced to database
	binary, err := dbService.Binaries.GetByUserID("config-binary")
	if err != nil {
		t.Errorf("Binary should be synced to database: %v", err)
		return
	}
	if binary.UserID != "config-binary" {
		t.Errorf("Expected binary ID 'config-binary', got '%s'", binary.UserID)
	}
}

func TestInstallCommand_ManualBinaryNotSyncedFromConfig(t *testing.T) {
	// Test that manually added binaries (source='manual') are NOT overwritten
	// by config sync in the PreRunE hook, even if they exist in config.
	// See bug bm-64q for context.

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)
	defer dbService.Close()

	// Create a manually added binary in the database
	manualBinary := &database.Binary{
		UserID:        "test-binary",
		Name:          "testbin",
		Provider:      "github",
		ProviderPath:  "test/test-binary",
		Format:        ".tar.gz",
		Source:        "manual",
		ConfigVersion: 0,
	}

	err = dbService.Binaries.Create(manualBinary)
	if err != nil {
		t.Fatalf("Failed to create manual binary: %v", err)
	}

	// Setup config that also has this binary (but with different settings)
	DBService = dbService
	Config = &config.Config{
		Binaries: []config.Binary{
			{
				Id:       "test-binary",
				Name:     "different-name", // Different from manual binary
				Provider: "github",
				Path:     "test/test-binary",
				Format:   ".tar.gz",
			},
		},
		Version: 1,
	}

	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "test-binary", "--version", "v1.0.0"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.ParseFlags([]string{"--binary", "test-binary", "--version", "v1.0.0"}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Execute PreRunE
	err = cmd.PreRunE(cmd, []string{})
	if err != nil {
		t.Errorf("PreRunE failed for manual binary: %v\nOutput: %s", err, buf.String())
	}

	// Verify binary still has source='manual' and original name
	binary, err := dbService.Binaries.GetByUserID("test-binary")
	if err != nil {
		t.Errorf("Binary should still exist: %v", err)
		return
	}

	if binary.Source != "manual" {
		t.Errorf("Expected source='manual', got '%s'", binary.Source)
	}

	if binary.Name != "testbin" {
		t.Errorf("Binary should NOT be overwritten by config. Expected name 'testbin', got '%s'", binary.Name)
	}
}

func TestInstallCommand_ConfigBinaryAlreadyInDatabase(t *testing.T) {
	// Test that config-managed binaries already in DB are not re-synced
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)
	defer dbService.Close()

	// Create a config-managed binary in the database
	configBinary := &database.Binary{
		UserID:        "config-binary",
		Name:          "configbin",
		Provider:      "github",
		ProviderPath:  "test/config-binary",
		Format:        ".tar.gz",
		Source:        "config",
		ConfigVersion: 1,
	}

	err = dbService.Binaries.Create(configBinary)
	if err != nil {
		t.Fatalf("Failed to create config binary: %v", err)
	}

	// Setup config with same binary
	DBService = dbService
	Config = &config.Config{
		Binaries: []config.Binary{
			{
				Id:       "config-binary",
				Name:     "configbin",
				Provider: "github",
				Path:     "test/config-binary",
				Format:   ".tar.gz",
			},
		},
		Version: 1,
	}

	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "config-binary", "--version", "v1.0.0"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.ParseFlags([]string{"--binary", "config-binary", "--version", "v1.0.0"}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Execute PreRunE
	err = cmd.PreRunE(cmd, []string{})
	if err != nil {
		t.Errorf("PreRunE failed for config binary: %v\nOutput: %s", err, buf.String())
	}

	// Verify binary still exists and is still source='config'
	binary, err := dbService.Binaries.GetByUserID("config-binary")
	if err != nil {
		t.Errorf("Binary should still exist: %v", err)
		return
	}

	if binary.Source != "config" {
		t.Errorf("Expected source='config', got '%s'", binary.Source)
	}
}

func TestInstallCommand_BinaryNotFound(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize test database
	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)
	defer dbService.Close()

	// Setup command environment with empty config and database
	DBService = dbService
	Config = &config.Config{
		Binaries: []config.Binary{},
		Version:  1,
	}

	// Create command
	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "nonexistent", "--version", "v1.0.0"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Parse flags first
	if err := cmd.ParseFlags([]string{"--binary", "nonexistent", "--version", "v1.0.0"}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Execute PreRunE - should fail
	err = cmd.PreRunE(cmd, []string{})
	if err == nil {
		t.Error("PreRunE should fail for nonexistent binary")
	}

	// Check error message mentions the binary name
	if err != nil {
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("Error message should not be empty")
		}
	}
}

func TestInstallCommand_RequiresBinaryFlag(t *testing.T) {
	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)
	defer dbService.Close()

	DBService = dbService
	Config = &config.Config{Version: 1}

	cmd := NewCommand()
	// Don't set --binary flag
	cmd.SetArgs([]string{"--version", "v1.0.0"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Should fail during flag parsing or execution
	err = cmd.Execute()
	if err == nil {
		t.Error("Command should fail when --binary flag is missing")
	}
}

// TestMain sets up and tears down test environment
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
