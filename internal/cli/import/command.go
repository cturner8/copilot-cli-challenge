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
		name         string
		version      string
		keepLocation bool
	)

	cmd := &cobra.Command{
		Use:   "import <path>",
		Short: "Import an existing binary from the file system",
		Long: `Import an existing binary that is already installed on your system.

This command will register the binary with binmate and create the necessary database records.
By default, the binary is copied to a managed location. Use --keep-location to use the original path.

Example:
  binmate import /usr/local/bin/gh --name gh
  binmate import /usr/local/bin/gh --name gh --version 2.0.0
  binmate import /usr/local/bin/gh --name gh --keep-location`,
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			if name == "" {
				return fmt.Errorf("please provide a name for the binary using --name flag")
			}

			_, err := binarySvc.ImportBinaryWithOptions(path, name, version, keepLocation, DBService)
			if err != nil {
				return fmt.Errorf("failed to import binary: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Binary %s imported successfully\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name for the imported binary (required)")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Version string (default: auto-generated)")
	cmd.Flags().BoolVarP(&keepLocation, "keep-location", "k", false, "Keep binary in original location instead of copying")
	cmd.MarkFlagRequired("name")

	return cmd
}
