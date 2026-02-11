package check

import (
	"fmt"

	"github.com/spf13/cobra"

	binarySvc "cturner8/binmate/internal/core/binary"
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database/repository"
	"cturner8/binmate/internal/providers/github"
)

// Package variables will be set by cmd package
var (
	Config    *config.Config
	DBService *repository.Service
)

func NewCommand() *cobra.Command {
	var (
		binaryID string
		checkAll bool
	)

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check for available updates without installing",
		Long: `Check if there are newer versions available for binaries without installing them.

This will query the provider for the latest version and compare it with the installed version.

Examples:
  binmate check --binary gh              # Check if gh has updates
  binmate check --all                    # Check all binaries for updates`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if checkAll {
				// Check all binaries
				binaries, err := binarySvc.ListBinariesWithDetails(DBService)
				if err != nil {
					return fmt.Errorf("failed to list binaries: %w", err)
				}

				if len(binaries) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No binaries to check")
					return nil
				}

				updatesAvailable := 0
				for _, b := range binaries {
					binaryConfig, err := DBService.Binaries.GetByUserID(b.Binary.UserID)
					if err != nil {
						fmt.Fprintf(cmd.OutOrStdout(), "⚠ Failed to check %s: %v\n", b.Binary.Name, err)
						continue
					}

					if binaryConfig.Provider != "github" {
						fmt.Fprintf(cmd.OutOrStdout(), "⚠ Skipping %s: only github provider is supported\n", b.Binary.Name)
						continue
					}

					release, _, err := github.FetchReleaseAsset(binaryConfig, "latest")
					if err != nil {
						fmt.Fprintf(cmd.OutOrStdout(), "⚠ Failed to check %s: %v\n", b.Binary.Name, err)
						continue
					}

					if b.ActiveVersion == "none" {
						fmt.Fprintf(cmd.OutOrStdout(), "ℹ %s: No version installed (latest: %s)\n", b.Binary.Name, release.TagName)
						updatesAvailable++
					} else if b.ActiveVersion != release.TagName {
						fmt.Fprintf(cmd.OutOrStdout(), "⬆ %s: Update available %s → %s\n", b.Binary.Name, b.ActiveVersion, release.TagName)
						updatesAvailable++
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "✓ %s: Up to date (%s)\n", b.Binary.Name, b.ActiveVersion)
					}
				}

				if updatesAvailable > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "\n%d update(s) available\n", updatesAvailable)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "\nAll binaries are up to date")
				}
				return nil
			}

			// Check single binary
			binaryConfig, err := DBService.Binaries.GetByUserID(binaryID)
			if err != nil {
				return fmt.Errorf("binary not found: %w", err)
			}

			if binaryConfig.Provider != "github" {
				return fmt.Errorf("only github provider is currently supported")
			}

			release, _, err := github.FetchReleaseAsset(binaryConfig, "latest")
			if err != nil {
				return fmt.Errorf("failed to fetch latest release: %w", err)
			}

			// Get current active version
			activeVersion, err := DBService.Versions.Get(binaryConfig.ID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "No version currently installed. Latest version: %s\n", release.TagName)
				return nil
			}

			installation, err := DBService.Installations.GetByID(activeVersion.InstallationID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "No version currently installed. Latest version: %s\n", release.TagName)
				return nil
			}

			if installation.Version == release.TagName {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ %s is up to date (version %s)\n", binaryID, installation.Version)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "⬆ Update available for %s: %s → %s\n", binaryID, installation.Version, release.TagName)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&binaryID, "binary", "b", "", "Binary ID to check for updates")
	cmd.Flags().BoolVarP(&checkAll, "all", "a", false, "Check all binaries for updates")

	// Make binary required unless --all is specified
	cmd.MarkFlagsOneRequired("binary", "all")
	cmd.MarkFlagsMutuallyExclusive("binary", "all")

	return cmd
}
