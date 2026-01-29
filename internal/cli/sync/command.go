package sync

import (
	"log"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database/repository"

	"github.com/spf13/cobra"
)

// Package variables will be set by cmd package
var (
	Config    *config.Config
	DBService *repository.Service
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync the local configuration file with the database.",
		Run: func(cmd *cobra.Command, args []string) {
			// Access package variables set by cmd
			if DBService == nil {
				log.Fatal("Database service not initialized")
			}

			err := config.SyncToDatabase(*Config, DBService)
			if err != nil {
				log.Fatalf("Error syncing to database: %v", err)
			}

			log.Println("Sync complete")
		},
	}

	return cmd
}
