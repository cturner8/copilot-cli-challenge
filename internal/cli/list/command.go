package list

import (
	"fmt"

	"github.com/spf13/cobra"

	binarySvc "cturner8/binmate/internal/core/binary"
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database/repository"
)

// Package variables will be set by cmd package
var (
	Config    *config.Config
	DBService *repository.Service
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all installed binaries",
		Long: `List all installed binaries.

Example:
  binmate list                          # List all binaries
  binmate ls                            # Same as above (alias)`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// List all binaries
			binaries, err := binarySvc.ListBinariesWithDetails(DBService)
			if err != nil {
				return fmt.Errorf("failed to list binaries: %w", err)
			}

			if len(binaries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No binaries installed")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-15s %-10s %s\n", "Binary", "Active Version", "Installed", "Provider")
			fmt.Fprintln(cmd.OutOrStdout(), "---")
			for _, b := range binaries {
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-15s %-10d %s:%s\n",
					b.Binary.Name,
					b.ActiveVersion,
					b.InstallCount,
					b.Binary.Provider,
					b.Binary.ProviderPath,
				)
			}

			return nil
		},
	}

	return cmd
}
