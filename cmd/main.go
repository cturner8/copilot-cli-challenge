/*
Copyright © 2026 cturner8
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"cturner8/binmate/internal/cli/add"
	"cturner8/binmate/internal/cli/check"
	configcmd "cturner8/binmate/internal/cli/config"
	importcmd "cturner8/binmate/internal/cli/import"
	"cturner8/binmate/internal/cli/install"
	"cturner8/binmate/internal/cli/list"
	"cturner8/binmate/internal/cli/remove"
	"cturner8/binmate/internal/cli/root"
	switchcmd "cturner8/binmate/internal/cli/switch"
	"cturner8/binmate/internal/cli/sync"
	"cturner8/binmate/internal/cli/update"
	versioncmd "cturner8/binmate/internal/cli/version"
	"cturner8/binmate/internal/cli/versions"
	"cturner8/binmate/internal/core/buildinfo"
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"

	"github.com/spf13/cobra"
)

var (
	rootCmd      = root.NewCommand()
	dbService    *repository.Service
	cfg          config.Config
	showVersion  bool
	buildVersion = buildinfo.DefaultVersion
	buildCommit  = buildinfo.DefaultUnknown
	buildDate    = buildinfo.DefaultUnknown
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func SetBuildMetadata(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date

	versioncmd.BuildVersion = version
	versioncmd.BuildCommit = commit
	versioncmd.BuildDate = date
}

func init() {
	var (
		configPath string
		logLevel   string
	)

	SetBuildMetadata(buildVersion, buildCommit, buildDate)

	originalRootPreRunE := rootCmd.PreRunE
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if showVersion {
			return nil
		}
		if originalRootPreRunE != nil {
			return originalRootPreRunE(cmd, args)
		}
		return nil
	}

	originalRootRunE := rootCmd.RunE
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if showVersion {
			info := buildinfo.Resolve(buildVersion, buildCommit, buildDate)
			fmt.Fprintf(cmd.OutOrStdout(), "binmate %s\n", info.Version)
			return nil
		}
		if originalRootRunE != nil {
			return originalRootRunE(cmd, args)
		}
		return nil
	}

	// set global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "(optional) path to the config file to use")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "(optional) controls verbosity of application logging")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version")

	// Setup database lifecycle hooks
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if cmd == rootCmd && showVersion {
			return
		}

		// Read config file
		cfg = config.ReadConfig(config.ConfigFlags{
			ConfigPath: configPath,
			LogLevel:   logLevel,
		})

		// Configure logger with appropriate level (handles silent mode)
		config.ConfigureLogger(cfg.LogLevel)

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
		versions.Config = &cfg
		versions.DBService = dbService
		versioncmd.BuildVersion = buildVersion
		versioncmd.BuildCommit = buildCommit
		versioncmd.BuildDate = buildDate
		check.Config = &cfg
		check.DBService = dbService
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
	rootCmd.AddCommand(versions.NewCommand())
	rootCmd.AddCommand(versioncmd.NewCommand())
	rootCmd.AddCommand(check.NewCommand())
}
