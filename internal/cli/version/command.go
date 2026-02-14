package versioncmd

import (
	"fmt"

	"cturner8/binmate/internal/core/buildinfo"

	"github.com/spf13/cobra"
)

var (
	BuildVersion = buildinfo.DefaultVersion
	BuildCommit  = buildinfo.DefaultUnknown
	BuildDate    = buildinfo.DefaultUnknown
)

func NewCommand() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Show binmate version information",
		Long: `Show binmate version information.

Examples:
  binmate version
  binmate version --verbose`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := buildinfo.Resolve(BuildVersion, BuildCommit, BuildDate)

			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "version: %s\n", info.Version)
				fmt.Fprintf(cmd.OutOrStdout(), "commit: %s\n", info.Commit)
				fmt.Fprintf(cmd.OutOrStdout(), "date: %s\n", info.Date)
				fmt.Fprintf(cmd.OutOrStdout(), "modified: %t\n", info.Modified)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "binmate %s\n", info.Version)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "V", false, "Show detailed build metadata")

	return cmd
}
