package update

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/core/config"
	installSvc "cturner8/binmate/internal/core/install"
	"cturner8/binmate/internal/database/repository"
)

// Package variables will be set by cmd package
var (
	Config    *config.Config
	DBService *repository.Service
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <binary-id>",
		Short: "Update a binary to the latest version",
		Long: `Update a binary to the latest available version.

This will install the latest version and set it as the active version.

Example:
  binmate update gh              # Update gh to latest version`,
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			binaryID := args[0]

			result, err := installSvc.UpdateToLatest(binaryID, DBService)
			if err != nil {
				return fmt.Errorf("failed to update binary: %w", err)
			}

			log.Printf("✓ Updated %s to version %s", binaryID, result.Version)
			return nil
		},
	}

	return cmd
}
