package update

import (
	"fmt"

	"github.com/spf13/cobra"

	binarySvc "cturner8/binmate/internal/core/binary"
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
		binaryID  string
		updateAll bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a binary to the latest version",
		Long: `Update a binary to the latest available version.

This will install the latest version and set it as the active version.

Examples:
  binmate update --binary gh              # Update gh to latest version
  binmate update --all                    # Update all binaries to latest versions`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if updateAll {
				// Update all binaries
				binaries, err := binarySvc.ListBinariesWithDetails(DBService)
				if err != nil {
					return fmt.Errorf("failed to list binaries: %w", err)
				}

				if len(binaries) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No binaries to update")
					return nil
				}

				updatedCount := 0
				for _, b := range binaries {
					result, err := installSvc.UpdateToLatest(b.Binary.UserID, DBService)
					if err != nil {
						fmt.Fprintf(cmd.OutOrStdout(), "⚠ Failed to update %s: %v\n", b.Binary.Name, err)
						continue
					}
					fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated %s to version %s\n", b.Binary.Name, result.Version)
					updatedCount++
				}

				fmt.Fprintf(cmd.OutOrStdout(), "\n✓ Updated %d/%d binaries\n", updatedCount, len(binaries))
				return nil
			}

			// Update single binary
			result, err := installSvc.UpdateToLatest(binaryID, DBService)
			if err != nil {
				return fmt.Errorf("failed to update binary: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated %s to version %s\n", binaryID, result.Version)
			return nil
		},
	}

	cmd.Flags().StringVarP(&binaryID, "binary", "b", "", "Binary ID to update")
	cmd.Flags().BoolVarP(&updateAll, "all", "a", false, "Update all binaries to their latest versions")

	// Make binary required unless --all is specified
	cmd.MarkFlagsOneRequired("binary", "all")
	cmd.MarkFlagsMutuallyExclusive("binary", "all")

	return cmd
}
