package sync

import (
	"fmt"
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
		Use:           "sync",
		Short:         "Sync the local configuration file with the database.",
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			id, err := DBService.Logs.LogStart("sync", "", "", "start sync process")
			if err != nil {
				return fmt.Errorf("sync start error: %w", err)
			}

			// Access package variables set by cmd
			if DBService == nil {
				msg := "database service not initialised"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s", msg)
			}

			if err := config.SyncToDatabase(*Config, DBService); err != nil {
				msg := "error syncing to database"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s: %w", msg, err)
			}

			log.Println("Sync complete")

			DBService.Logs.LogSuccess(id, int64(time.Since(start)))
			return nil
		},
	}

	return cmd
}
