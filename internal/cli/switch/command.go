package switchcmd

import (
	"fmt"
	"log"

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
	cmd := &cobra.Command{
		Use:   "switch <binary-id> <version>",
		Short: "Switch to a different installed version",
		Long: `Switch the active version of a binary to a different installed version.

The version must already be installed. Use 'binmate list <binary-id>' to see available versions.

Example:
  binmate switch gh v2.30.0      # Switch gh to version v2.30.0`,
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			binaryID := args[0]
			version := args[1]

			if err := versionSvc.SwitchVersion(binaryID, version, DBService); err != nil {
				return fmt.Errorf("failed to switch version: %w", err)
			}

			log.Printf("✓ Switched %s to version %s", binaryID, version)
			return nil
		},
	}

	return cmd
}
