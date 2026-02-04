package remove

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
	var (
		binaryID    string
		removeFiles bool
	)

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a binary and its installations",
		Long: `Remove a binary from binmate.

This will remove the binary from the database along with all installation records.
Use --files to also remove the physical binary files and symlinks.

Example:
  binmate remove --binary gh              # Remove binary from database only
  binmate remove --binary gh --files      # Remove binary and all files`,
		Aliases:       []string{"rm", "delete"},
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if binaryID == "" {
				return fmt.Errorf("please provide a binary ID using --binary flag")
			}

			if err := binarySvc.RemoveBinary(binaryID, DBService, removeFiles); err != nil {
				return fmt.Errorf("failed to remove binary: %w", err)
			}

			if removeFiles {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Binary %s removed (including files)\n", binaryID)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Binary %s removed from database\n", binaryID)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&binaryID, "binary", "b", "", "Binary ID to remove (required)")
	cmd.Flags().BoolVarP(&removeFiles, "files", "f", false, "Also remove physical binary files and symlinks")
	cmd.MarkFlagRequired("binary")

	return cmd
}
