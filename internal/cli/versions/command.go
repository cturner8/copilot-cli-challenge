package versions

import (
	"fmt"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/core/format"
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
	)

	cmd := &cobra.Command{
		Use:   "versions",
		Short: "List installed versions of a binary",
		Long: `List all installed versions of a specific binary.

Example:
  binmate versions --binary gh              # List versions of gh binary`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// List versions for specific binary
			versions, err := versionSvc.ListVersions(binaryID, DBService)
			if err != nil {
				return fmt.Errorf("failed to list versions: %w", err)
			}

			// Get active version
			activeVersion, _ := versionSvc.GetActiveVersion(binaryID, DBService)

			fmt.Fprintf(cmd.OutOrStdout(), "Versions for %s:\n", binaryID)
			fmt.Fprintln(cmd.OutOrStdout(), "---")

			// Get date format from config or use default
			dateFormat := ""
			if Config != nil && Config.DateFormat != "" {
				dateFormat = Config.DateFormat
			}

			for _, v := range versions {
				activeMarker := " "
				if activeVersion != nil && activeVersion.InstallationID == v.ID {
					activeMarker = "*"
				}
				formattedDate := format.FormatTimestamp(v.InstalledAt, dateFormat)
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s (installed: %s)\n", activeMarker, v.Version, formattedDate)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&binaryID, "binary", "b", "", "Binary ID to list versions for (required)")
	cmd.MarkFlagRequired("binary")

	return cmd
}
