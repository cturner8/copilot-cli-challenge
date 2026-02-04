/*
Copyright © 2026 cturner8
*/
package cmd

import (
	"log"
	"os"

	"cturner8/binmate/internal/cli/install"
	"cturner8/binmate/internal/cli/root"
	"cturner8/binmate/internal/cli/sync"
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"

	"github.com/spf13/cobra"
)

var (
	rootCmd   = root.NewCommand()
	dbService *repository.Service
	cfg       config.Config
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var (
		configPath string
	)

	// set global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "(optional) path to the config file to use")

	// Setup database lifecycle hooks
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Read config file
		cfg = config.ReadConfig(configPath)

		// Resolve database path
		dbPath, err := database.GetDefaultDBPath()
		if err != nil {
			log.Fatalf("Unable to locate database path: %v", err)
		}

		// Initialize database with migrations
		db, err := database.Initialize(dbPath)
		if err != nil {
			log.Fatalf("Unable to initialize database: %v", err)
		}

		// Create database service
		dbService = repository.NewService(db)

		// Set package variables for commands
		root.Config = &cfg
		root.DBService = dbService
		install.Config = &cfg
		install.DBService = dbService
		sync.Config = &cfg
		sync.DBService = dbService
	}

	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if dbService != nil {
			if err := dbService.Close(); err != nil {
				log.Printf("Warning: failed to close database: %v", err)
			}
		}
	}

	// Register subcommands
	rootCmd.AddCommand(install.NewCommand())
	rootCmd.AddCommand(sync.NewCommand())
}
