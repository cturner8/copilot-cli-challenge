package database

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOpen(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

func TestGetDefaultDBPath(t *testing.T) {
	path, err := GetDefaultDBPath()
	if err != nil {
		t.Fatalf("Failed to get default DB path: %v", err)
	}

	// Verify path is absolute
	if !filepath.IsAbs(path) {
		t.Error("DB path is not absolute")
	}

	// Verify path contains 'binmate' directory
	if !contains(path, "binmate") {
		t.Errorf("DB path should contain 'binmate' directory: %s", path)
	}

	// Verify path ends with user.db
	if filepath.Base(path) != "user.db" {
		t.Errorf("DB path should end with 'user.db', got: %s", filepath.Base(path))
	}

	// Platform-specific checks
	switch runtime.GOOS {
	case "windows":
		// On Windows, should be in LocalAppData
		if !contains(path, "AppData") && !contains(path, "Local") {
			t.Logf("Note: Expected Windows path in LocalAppData, got: %s", path)
		}
	case "darwin":
		// On macOS, path can vary but should be valid
		t.Logf("macOS DB path: %s", path)
	default:
		// On Linux, should contain .local/share
		if !contains(path, ".local") || !contains(path, "share") {
			t.Errorf("Linux DB path should contain '.local/share': %s", path)
		}
	}
}

func TestPragmas(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Verify foreign keys are enabled
	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Fatalf("Failed to check foreign keys pragma: %v", err)
	}
	if fkEnabled != 1 {
		t.Error("Foreign keys are not enabled")
	}

	// Verify journal mode is WAL
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to check journal mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("Expected journal mode 'wal', got '%s'", journalMode)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
