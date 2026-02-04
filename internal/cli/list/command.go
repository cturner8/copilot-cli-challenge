package list

import (
	"fmt"

	"github.com/spf13/cobra"

	binarySvc "cturner8/binmate/internal/core/binary"
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
		showVersions bool
		binaryID     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed binaries and their versions",
		Long: `List all installed binaries or versions of a specific binary.

Examples:
  binmate list                          # List all binaries
  binmate list --binary gh              # List versions of gh binary
  binmate list --versions --binary gh   # List versions of gh binary (explicit flag)`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if we should list versions for a specific binary
			if binaryID != "" || showVersions {
				if binaryID == "" {
					return fmt.Errorf("please provide a binary ID using --binary flag when using --versions")
				}

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
			}

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

	cmd.Flags().BoolVarP(&showVersions, "versions", "v", false, "Show versions for a specific binary")
	cmd.Flags().StringVarP(&binaryID, "binary", "b", "", "Binary ID to list versions for")

	return cmd
}
