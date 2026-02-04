package install

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/core/config"
	installSvc "cturner8/binmate/internal/core/install"
	"cturner8/binmate/internal/database/repository"
)

// Package variable will be set by cmd package
var (
	Config    *config.Config
	DBService *repository.Service
)

func NewCommand() *cobra.Command {
	var (
		binary  string
		version string
	)

	cmd := &cobra.Command{
		Use:           "install",
		Short:         "Install a new binary version",
		Aliases:       []string{"i"},
		SilenceUsage:  true,  // Don't show usage on runtime errors
		SilenceErrors: false, // Still print errors
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Sync the specific binary from config to database
			if err := config.SyncBinary(binary, *Config, DBService); err != nil {
				return fmt.Errorf("failed to sync binary config: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			id, err := DBService.Logs.LogStart("install", "", "", "start install process")
			if err != nil {
				return fmt.Errorf("sync start error: %w", err)
			}

			log.Printf("installing binary: %s version: %s", binary, version)

			DBService.Logs.LogEntity(id, binary, version)

			// Use the service layer to install the binary
			result, err := installSvc.InstallBinary(binary, version, DBService)
			if err != nil {
				msg := "installation failed"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s: %w", msg, err)
			}

			log.Printf("downloaded binary: %s version: %s", binary, result.Version)

			DBService.Logs.LogSuccess(id, int64(time.Since(start)))
			return nil
		},
	}

	cmd.Flags().StringVar(&binary, "binary", "", "binary to be installed")
	cmd.Flags().StringVar(&version, "version", "latest", "version of the binary to be installed")

	// Mark required flags
	cmd.MarkFlagRequired("binary")

	return cmd
}
