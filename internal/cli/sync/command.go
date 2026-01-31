package sync

import (
	"log"
	"time"

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
			start := time.Now()

			id, err := DBService.Logs.LogStart("sync", "", "", "start sync process")
			if err != nil {
				log.Fatalf("sync start error: %s", err)
			}

			// Access package variables set by cmd
			if DBService == nil {
				msg := "Database service not initialized"
				log.Fatal(msg)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			err = config.SyncToDatabase(*Config, DBService)
			if err != nil {
				log.Fatalf("Error syncing to database: %v", err)
			}

			log.Println("Sync complete")

			DBService.Logs.LogSuccess(id, int64(time.Since(start)))
		},
	}

	return cmd
}
