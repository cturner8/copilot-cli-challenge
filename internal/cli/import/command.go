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
		name          string
		version       string
		url           string
		authenticated bool
		keepLocation  bool
	)

	cmd := &cobra.Command{
		Use:   "import <path>",
		Short: "Import an existing binary from the file system",
		Long: `Import an existing binary that is already installed on your system.

This command will register the binary with binmate and create the necessary database records.
By default, the binary is copied to a managed location. Use --keep-location to use the original path.

You can optionally associate the imported binary with a GitHub release URL to enable future
install and update functionality. The version will be automatically extracted from the URL:
  binmate import /usr/local/bin/gh --url https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz

Example:
  binmate import /usr/local/bin/gh --name gh
  binmate import /usr/local/bin/gh --name gh --version 2.0.0
  binmate import /usr/local/bin/gh --name gh --keep-location
  binmate import /usr/local/bin/gh --url https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz
  binmate import /usr/local/bin/gh --url https://github.com/cli/cli/releases/download/v2.30.0/gh_2.30.0_linux_amd64.tar.gz --version v2.30.0-custom`,
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			if name == "" && url == "" {
				return fmt.Errorf("please provide either a name (--name) or GitHub URL (--url)")
			}

			_, err := binarySvc.ImportBinaryWithOptions(path, name, url, version, authenticated, keepLocation, DBService)
			if err != nil {
				return fmt.Errorf("failed to import binary: %w", err)
			}

			if url != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Binary imported and associated with GitHub repository\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Binary %s imported successfully\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name for the imported binary")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Version string (default: auto-extracted from URL or auto-generated)")
	cmd.Flags().StringVarP(&url, "url", "u", "", "GitHub release URL to associate with the binary (version auto-extracted)")
	cmd.Flags().BoolVarP(&authenticated, "authenticated", "a", false, "Use GitHub token authentication for private repos")
	cmd.Flags().BoolVarP(&keepLocation, "keep-location", "k", false, "Keep binary in original location instead of copying")

	return cmd
}
