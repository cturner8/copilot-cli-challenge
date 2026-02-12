package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestInstallCommand_ConfigBinaryReSyncedFromConfig(t *testing.T) {
	// Test that config-managed binaries ARE re-synced from config during PreRun
	// This ensures config changes are picked up even if binary exists in DB
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)
	defer dbService.Close()

	// Create a config-managed binary in the database with old name
	configBinary := &database.Binary{
		UserID:        "config-binary",
		Name:          "oldname",
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

	// Setup config with updated binary name
	DBService = dbService
	Config = &config.Config{
		Binaries: []config.Binary{
			{
				Id:       "config-binary",
				Name:     "newname", // Updated name in config
				Provider: "github",
				Path:     "test/config-binary",
				Format:   ".tar.gz",
			},
		},
		Version: 2,
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

	// Verify binary was updated from config
	binary, err := dbService.Binaries.GetByUserID("config-binary")
	if err != nil {
		t.Errorf("Binary should still exist: %v", err)
		return
	}

	if binary.Source != "config" {
		t.Errorf("Expected source='config', got '%s'", binary.Source)
	}

	if binary.Name != "newname" {
		t.Errorf("Binary should be updated from config. Expected name 'newname', got '%s'", binary.Name)
	}

	if binary.ConfigVersion != 2 {
		t.Errorf("Expected ConfigVersion=2, got %d", binary.ConfigVersion)
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

type githubRewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (t *githubRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.target.Scheme
	clone.URL.Host = t.target.Host
	return t.base.RoundTrip(clone)
}

func routeGitHubAPIToServer(t *testing.T, srv *httptest.Server) {
	t.Helper()
	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("Failed to parse test server URL: %v", err)
	}

	original := http.DefaultTransport
	http.DefaultTransport = &githubRewriteTransport{
		target: target,
		base:   srv.Client().Transport,
	}

	t.Cleanup(func() {
		http.DefaultTransport = original
	})
}

func createTarGzArchive(t *testing.T, fileName string, content []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	hdr := &tar.Header{
		Name: fileName,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("Failed to close tar writer: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %v", err)
	}

	return buf.Bytes()
}

func TestInstallCommand_RunE_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)
	defer dbService.Close()

	// Existing manual binary ensures PreRunE skips config sync.
	testBinary := &database.Binary{
		UserID:       "test-binary",
		Name:         "testbin",
		Provider:     "github",
		ProviderPath: "test/test-binary",
		Format:       ".tar.gz",
	}
	if err := dbService.Binaries.Create(testBinary); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	archiveBytes := createTarGzArchive(t, "testbin", []byte("#!/bin/sh\necho ok\n"))
	assetName := fmt.Sprintf("testbin-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/test/test-binary/releases/latest":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"tag_name":"v1.2.3","assets":[{"id":1,"name":"%s","browser_download_url":"https://api.github.com/download/%s"}]}`, assetName, assetName)
		case "/download/" + assetName:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	routeGitHubAPIToServer(t, server)

	envHome := filepath.Join(tmpDir, "home")
	t.Setenv("HOME", envHome)
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmpDir, "cache"))

	DBService = dbService
	Config = &config.Config{Version: 1}

	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "test-binary", "--version", "latest"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Command should succeed, got error: %v\nOutput: %s", err, buf.String())
	}

	out := buf.String()
	if !strings.Contains(out, "Installing test-binary version latest...") {
		t.Fatalf("Expected install start output, got: %s", out)
	}
	if !strings.Contains(out, "Successfully installed test-binary version v1.2.3") {
		t.Fatalf("Expected install success output, got: %s", out)
	}

	installation, err := dbService.Installations.Get(testBinary.ID, "v1.2.3")
	if err != nil {
		t.Fatalf("Expected installation record for v1.2.3, got error: %v", err)
	}
	if installation.BinaryID != testBinary.ID {
		t.Fatalf("Expected binary ID %d, got %d", testBinary.ID, installation.BinaryID)
	}
}

func TestInstallCommand_RunE_FailureLogsAndReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	dbService := repository.NewService(db)
	defer dbService.Close()

	unsupportedBinary := &database.Binary{
		UserID:       "unsupported-binary",
		Name:         "badbin",
		Provider:     "unsupported",
		ProviderPath: "owner/repo",
		Format:       ".tar.gz",
	}
	if err := dbService.Binaries.Create(unsupportedBinary); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	DBService = dbService
	Config = &config.Config{Version: 1}

	cmd := NewCommand()
	cmd.SetArgs([]string{"--binary", "unsupported-binary", "--version", "v1.0.0"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Fatal("Expected command to fail for unsupported provider")
	}
	if !strings.Contains(err.Error(), "installation failed") {
		t.Fatalf("Expected wrapped installation error, got: %v", err)
	}

	logs, logErr := dbService.Logs.GetFailures(10)
	if logErr != nil {
		t.Fatalf("Failed to read failure logs: %v", logErr)
	}
	if len(logs) == 0 {
		t.Fatal("Expected at least one failed log entry")
	}
}

func TestInstallCommand_RunE_LogStartError(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	dbService := repository.NewService(db)
	DBService = dbService
	Config = &config.Config{Version: 1}

	// Force LogStart to fail.
	if err := dbService.Close(); err != nil {
		t.Fatalf("Failed to close db service: %v", err)
	}

	cmd := NewCommand()
	if err := cmd.ParseFlags([]string{"--binary", "test-binary", "--version", "v1.0.0"}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	err = cmd.RunE(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "sync start error") {
		t.Fatalf("Expected sync start error, got: %v", err)
	}
}

// TestMain sets up and tears down test environment
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
