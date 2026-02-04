package database

import (
	"database/sql"
	"fmt"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

var migrations = []Migration{
	{
		Version:     1,
		Description: "Initial schema",
		SQL:         InitialSchema,
	},
}

// Migrate runs all pending migrations
func (db *DB) Migrate() error {
	// Get current schema version
	currentVersion, err := db.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			if err := db.runMigration(migration); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

// getCurrentVersion returns the current schema version
func (db *DB) getCurrentVersion() (int, error) {
	// Check if migrations table exists
	var tableName string
	err := db.QueryRow(`
SELECT name FROM sqlite_master 
WHERE type='table' AND name='migrations'
`).Scan(&tableName)

	if err == sql.ErrNoRows {
		return 0, nil // No migrations table, version 0
	}
	if err != nil {
		return 0, err
	}

	// Get latest version
	var version int
	err = db.QueryRow(`
SELECT COALESCE(MAX(version), 0) FROM migrations
`).Scan(&version)

	return version, err
}

// runMigration executes a single migration
func (db *DB) runMigration(migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// Initialize creates the database and runs all migrations
func Initialize(dbPath string) (*DB, error) {
	db, err := Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
