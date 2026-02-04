/*
Copyright © 2026 cturner8
*/
package cmd

import (
	"log"
	"os"

	"cturner8/binmate/internal/cli/add"
	configcmd "cturner8/binmate/internal/cli/config"
	importcmd "cturner8/binmate/internal/cli/import"
	"cturner8/binmate/internal/cli/install"
	"cturner8/binmate/internal/cli/list"
	"cturner8/binmate/internal/cli/remove"
	"cturner8/binmate/internal/cli/root"
	switchcmd "cturner8/binmate/internal/cli/switch"
	"cturner8/binmate/internal/cli/sync"
	"cturner8/binmate/internal/cli/update"
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
	// set global flags
	// TODO: use this to override config path
	rootCmd.PersistentFlags().String("config", "", "(optional) path to the config file to use")

	// Setup database lifecycle hooks
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Read config file
		cfg = config.ReadConfig()

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
		add.Config = &cfg
		add.DBService = dbService
		list.Config = &cfg
		list.DBService = dbService
		remove.Config = &cfg
		remove.DBService = dbService
		switchcmd.Config = &cfg
		switchcmd.DBService = dbService
		update.Config = &cfg
		update.DBService = dbService
		importcmd.Config = &cfg
		importcmd.DBService = dbService
		configcmd.Config = &cfg
		configcmd.DBService = dbService
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
	rootCmd.AddCommand(add.NewCommand())
	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(remove.NewCommand())
	rootCmd.AddCommand(switchcmd.NewCommand())
	rootCmd.AddCommand(update.NewCommand())
	rootCmd.AddCommand(importcmd.NewCommand())
	rootCmd.AddCommand(configcmd.NewCommand())
}
