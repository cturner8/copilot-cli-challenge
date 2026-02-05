package importcmd

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
		name string
	)

	cmd := &cobra.Command{
		Use:   "import <path>",
		Short: "Import an existing binary from the file system",
		Long: `Import an existing binary that is already installed on your system.

This command will register the binary with binmate and create the necessary database records.

Example:
  binmate import /usr/local/bin/gh --name gh`,
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			if name == "" {
				return fmt.Errorf("please provide a name for the binary using --name flag")
			}

			_, err := binarySvc.ImportBinary(path, name, DBService)
			if err != nil {
				return fmt.Errorf("failed to import binary: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Binary %s imported successfully\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name for the imported binary (required)")
	cmd.MarkFlagRequired("name")

	return cmd
}
