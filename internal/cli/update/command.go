package update

import (
	"fmt"

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
	var (
		binaryID string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a binary to the latest version",
		Long: `Update a binary to the latest available version.

This will install the latest version and set it as the active version.

Example:
  binmate update --binary gh              # Update gh to latest version`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := installSvc.UpdateToLatest(binaryID, DBService)
			if err != nil {
				return fmt.Errorf("failed to update binary: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated %s to version %s\n", binaryID, result.Version)
			return nil
		},
	}

	cmd.Flags().StringVarP(&binaryID, "binary", "b", "", "Binary ID to update (required)")
	cmd.MarkFlagRequired("binary")

	return cmd
}
