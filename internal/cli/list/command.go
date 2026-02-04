package list

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	binarySvc "cturner8/binmate/internal/core/binary"
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
		showVersions bool
		binaryID     string
	)

	cmd := &cobra.Command{
		Use:   "list [binary-id]",
		Short: "List installed binaries and their versions",
		Long: `List all installed binaries or versions of a specific binary.

Examples:
  binmate list                    # List all binaries
  binmate list gh                 # List versions of gh binary
  binmate list --versions gh      # List versions of gh binary (explicit flag)`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If binary ID is provided (either as arg or flag), list its versions
			if len(args) > 0 {
				binaryID = args[0]
			}

			if binaryID != "" || showVersions {
				if binaryID == "" {
					return fmt.Errorf("please provide a binary ID when using --versions")
				}

				// List versions for specific binary
				versions, err := versionSvc.ListVersions(binaryID, DBService)
				if err != nil {
					return fmt.Errorf("failed to list versions: %w", err)
				}

				// Get active version
				activeVersion, err := versionSvc.GetActiveVersion(binaryID, DBService)
				if err != nil {
					log.Printf("No active version set for %s", binaryID)
				}

				fmt.Printf("Versions for %s:\n", binaryID)
				fmt.Println("---")
				for _, v := range versions {
					activeMarker := " "
					if activeVersion != nil && activeVersion.InstallationID == v.ID {
						activeMarker = "*"
					}
					fmt.Printf("%s %s (installed: %s)\n", activeMarker, v.Version, formatTimestamp(v.InstalledAt))
				}
				return nil
			}

			// List all binaries
			binaries, err := binarySvc.ListBinariesWithDetails(DBService)
			if err != nil {
				return fmt.Errorf("failed to list binaries: %w", err)
			}

			if len(binaries) == 0 {
				fmt.Println("No binaries installed")
				return nil
			}

			fmt.Printf("%-20s %-15s %-10s %s\n", "Binary", "Active Version", "Installed", "Provider")
			fmt.Println("---")
			for _, b := range binaries {
				fmt.Printf("%-20s %-15s %-10d %s:%s\n",
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
	cmd.Flags().StringVar(&binaryID, "binary", "", "Binary ID to list versions for")

	return cmd
}

func formatTimestamp(timestamp int64) string {
	// Simple formatting - just return as string for now
	return fmt.Sprintf("%d", timestamp)
}
