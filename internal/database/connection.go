package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection
type DB struct {
	*sql.DB
}

// Open creates and initialises a new database connection
func Open(dbPath string) (*DB, error) {
	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite database
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{sqlDB}

	// Configure connection pool
	db.SetMaxOpenConns(1) // SQLite works best with single writer
	db.SetMaxIdleConns(1)

	// Apply performance pragmas
	if err := db.configurePragmas(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure pragmas: %w", err)
	}

	return db, nil
}

// configurePragmas sets SQLite pragmas for performance and safety
func (db *DB) configurePragmas() error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -64000", // 64MB
		"PRAGMA temp_store = MEMORY",
		"PRAGMA busy_timeout = 5000", // 5 seconds
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}

	return nil
}

// GetDefaultDBPath returns the default database path based on OS
func GetDefaultDBPath() (string, error) {
	var baseDir string
	var err error

	if runtime.GOOS == "windows" {
		// On Windows, use %LOCALAPPDATA%
		// os.UserCacheDir() returns %LOCALAPPDATA% on Windows
		baseDir, err = os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("failed to get cache directory: %w", err)
		}
	} else {
		// On Linux/macOS, use ~/.local/share
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".local", "share")
	}

	dbPath := filepath.Join(baseDir, "binmate", "user.db")
	return dbPath, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
