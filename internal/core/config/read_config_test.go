package config

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	return path
}

func TestReadConfig_ConfigPathOverridesEnv(t *testing.T) {
	envPath := writeTempConfig(t, `{
"version": 10,
"logLevel": "error",
"binaries": [{"id":"env-bin","name":"env","provider":"github","path":"owner/env","format":".zip"}]
}`)
	flagPath := writeTempConfig(t, `{
"version": 20,
"logLevel": "warn",
"binaries": [{"id":"flag-bin","name":"flag","provider":"github","path":"owner/flag","format":".tar.gz"}]
}`)

	t.Setenv("BINMATE_CONFIG_PATH", envPath)
	cfg := ReadConfig(ConfigFlags{ConfigPath: flagPath})

	if cfg.Version != 20 {
		t.Fatalf("expected version 20 from --config path, got %d", cfg.Version)
	}
	if len(cfg.Binaries) != 1 || cfg.Binaries[0].Id != "flag-bin" {
		t.Fatalf("expected binary from --config path, got %#v", cfg.Binaries)
	}
}

func TestReadConfig_UsesEnvPathWhenNoFlagPath(t *testing.T) {
	envPath := writeTempConfig(t, `{
"version": 30,
"logLevel": "debug",
"binaries": [{"id":"env-only","name":"env","provider":"github","path":"owner/env","format":".zip"}]
}`)

	t.Setenv("BINMATE_CONFIG_PATH", envPath)
	cfg := ReadConfig(ConfigFlags{})

	if cfg.Version != 30 {
		t.Fatalf("expected version 30 from BINMATE_CONFIG_PATH, got %d", cfg.Version)
	}
	if len(cfg.Binaries) != 1 || cfg.Binaries[0].Id != "env-only" {
		t.Fatalf("expected env binary, got %#v", cfg.Binaries)
	}
}

func TestReadConfig_DefaultsWhenConfigMissingOrMalformed(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		cfg := ReadConfig(ConfigFlags{
			ConfigPath: filepath.Join(t.TempDir(), "missing.json"),
			LogLevel:   "warn",
		})

		if cfg.Version != 1 {
			t.Fatalf("expected default version 1, got %d", cfg.Version)
		}
		if cfg.LogLevel != "warn" {
			t.Fatalf("expected logLevel from flags, got %q", cfg.LogLevel)
		}
		if len(cfg.Binaries) != 0 {
			t.Fatalf("expected default empty binaries, got %d", len(cfg.Binaries))
		}
	})

	t.Run("malformed file", func(t *testing.T) {
		path := writeTempConfig(t, `{"version": 2,`)
		cfg := ReadConfig(ConfigFlags{ConfigPath: path})

		if cfg.Version != 1 {
			t.Fatalf("expected fallback default version 1, got %d", cfg.Version)
		}
		if cfg.LogLevel != "silent" {
			t.Fatalf("expected default silent log level, got %q", cfg.LogLevel)
		}
	})
}

func TestGetBinary(t *testing.T) {
	binaries := []Binary{
		{Id: "gh", Name: "GitHub CLI"},
		{Id: "fzf", Name: "fzf"},
	}

	got, err := GetBinary("fzf", binaries)
	if err != nil {
		t.Fatalf("expected binary to be found, got error: %v", err)
	}
	if got.Name != "fzf" {
		t.Fatalf("expected fzf, got %q", got.Name)
	}

	if _, err := GetBinary("missing", binaries); err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  slog.Level
	}{
		{"silent", "silent", LevelSilent},
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"warning alias", "warning", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"empty defaults to info", "", slog.LevelInfo},
		{"invalid defaults to info", "nope", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseLogLevel(tt.input); got != tt.want {
				t.Fatalf("ParseLogLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestReadAndSync(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("failed to initialise database: %v", err)
	}
	defer db.Close()

	svc := repository.NewService(db)
	cfg := Config{
		Version: 1,
		Binaries: []Binary{
			{
				Id:       "test-binary",
				Name:     "test",
				Provider: "github",
				Path:     "owner/repo",
				Format:   ".tar.gz",
			},
		},
	}

	returnedCfg, err := ReadAndSync(cfg, svc)
	if err != nil {
		t.Fatalf("ReadAndSync failed: %v", err)
	}
	if returnedCfg.Version != cfg.Version {
		t.Fatalf("expected returned config version %d, got %d", cfg.Version, returnedCfg.Version)
	}

	bin, err := svc.Binaries.GetByUserID("test-binary")
	if err != nil {
		t.Fatalf("expected synced binary, got error: %v", err)
	}
	if bin.UserID != "test-binary" {
		t.Fatalf("expected user_id test-binary, got %q", bin.UserID)
	}
}

func TestSyncBinary(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("failed to initialise database: %v", err)
	}
	defer db.Close()

	svc := repository.NewService(db)
	cfg := Config{
		Version: 1,
		Global: GlobalConfig{
			InstallPath: "/usr/local/bin",
		},
		Binaries: []Binary{
			{
				Id:       "sync-bin",
				Name:     "sync",
				Provider: "github",
				Path:     "owner/repo",
				Format:   ".zip",
			},
		},
	}

	if err := SyncBinary("sync-bin", cfg, svc); err != nil {
		t.Fatalf("SyncBinary create failed: %v", err)
	}

	created, err := svc.Binaries.GetByUserID("sync-bin")
	if err != nil {
		t.Fatalf("expected created binary, got error: %v", err)
	}
	if created.Source != "config" {
		t.Fatalf("expected source=config, got %q", created.Source)
	}
	if created.InstallPath == nil || *created.InstallPath != "/usr/local/bin" {
		t.Fatalf("expected merged install path from global config, got %#v", created.InstallPath)
	}

	cfg.Version = 2
	cfg.Binaries[0].Name = "sync-updated"
	if err := SyncBinary("sync-bin", cfg, svc); err != nil {
		t.Fatalf("SyncBinary update failed: %v", err)
	}

	updated, err := svc.Binaries.GetByUserID("sync-bin")
	if err != nil {
		t.Fatalf("expected updated binary, got error: %v", err)
	}
	if updated.Name != "sync-updated" {
		t.Fatalf("expected updated name, got %q", updated.Name)
	}
	if updated.ConfigVersion != 2 {
		t.Fatalf("expected config version 2, got %d", updated.ConfigVersion)
	}

	if err := SyncBinary("missing", cfg, svc); err == nil {
		t.Fatal("expected error for missing binary in config")
	}
}

func TestConfigureLogger(t *testing.T) {
	ConfigureLogger("silent")
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("silent logger should not enable debug level")
	}

	ConfigureLogger("debug")
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("debug logger should enable debug level")
	}
}
