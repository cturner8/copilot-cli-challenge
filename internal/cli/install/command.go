package install

import (
	"fmt"
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
			// Check if binary exists in database first
			existingBinary, err := DBService.Binaries.GetByUserID(binary)
			if err == nil {
				// Binary exists - check if it's manually added
				if existingBinary.Source == "manual" {
					// Manually added binary, don't sync from config
					return nil
				}
				// Config-managed binary already in database, no need to sync again
				return nil
			}

			// Binary not in database, try to sync from config
			if err := config.SyncBinary(binary, *Config, DBService); err != nil {
				return fmt.Errorf("binary '%s' not found in database or config: %w", binary, err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			id, err := DBService.Logs.LogStart("install", "", "", "start install process")
			if err != nil {
				return fmt.Errorf("sync start error: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Installing %s version %s...\n", binary, version)

			DBService.Logs.LogEntity(id, binary, version)

			// Use the service layer to install the binary
			result, err := installSvc.InstallBinary(binary, version, DBService)
			if err != nil {
				msg := "installation failed"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s: %w", msg, err)
			}

			DBService.Logs.LogSuccess(id, int64(time.Since(start)))

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Successfully installed %s version %s\n", binary, result.Version)
			return nil
		},
	}

	cmd.Flags().StringVarP(&binary, "binary", "b", "", "binary to be installed")
	cmd.Flags().StringVarP(&version, "version", "v", "latest", "version of the binary to be installed")

	// Mark required flags
	cmd.MarkFlagRequired("binary")

	return cmd
}
