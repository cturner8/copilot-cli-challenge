package switchcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/core/config"
	versionSvc "cturner8/binmate/internal/core/version"
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
		version  string
	)

	cmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch to a different installed version",
		Long: `Switch the active version of a binary to a different installed version.

The version must already be installed. Use 'binmate list --binary <binary-id>' to see available versions.

Example:
  binmate switch --binary gh --version v2.30.0      # Switch gh to version v2.30.0`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if binaryID == "" {
				return fmt.Errorf("please provide a binary ID using --binary flag")
			}
			if version == "" {
				return fmt.Errorf("please provide a version using --version flag")
			}

			if err := versionSvc.SwitchVersion(binaryID, version, DBService); err != nil {
				return fmt.Errorf("failed to switch version: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Switched %s to version %s\n", binaryID, version)
			return nil
		},
	}

	cmd.Flags().StringVarP(&binaryID, "binary", "b", "", "Binary ID (required)")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Version to switch to (required)")
	cmd.MarkFlagRequired("binary")
	cmd.MarkFlagRequired("version")

	return cmd
}
