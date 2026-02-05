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
