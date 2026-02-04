package remove

import (
	"fmt"
	"log"

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
		removeFiles bool
	)

	cmd := &cobra.Command{
		Use:   "remove <binary-id>",
		Short: "Remove a binary and its installations",
		Long: `Remove a binary from binmate.

This will remove the binary from the database along with all installation records.
Use --files to also remove the physical binary files and symlinks.

Example:
  binmate remove gh              # Remove binary from database only
  binmate remove gh --files      # Remove binary and all files`,
		Aliases:       []string{"rm", "delete"},
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			binaryID := args[0]

			if err := binarySvc.RemoveBinary(binaryID, DBService, removeFiles); err != nil {
				return fmt.Errorf("failed to remove binary: %w", err)
			}

			if removeFiles {
				log.Printf("✓ Binary %s removed (including files)", binaryID)
			} else {
				log.Printf("✓ Binary %s removed from database", binaryID)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&removeFiles, "files", false, "Also remove physical binary files and symlinks")

	return cmd
}
