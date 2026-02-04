package configcmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

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
		showJSON bool
	)

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show or edit configuration",
		Long: `Display the current binmate configuration.

This shows the configuration file contents and database location.

Example:
  binmate config                 # Show config as table
  binmate config --json          # Show config as JSON`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showJSON {
				// Output config as JSON
				jsonData, err := json.MarshalIndent(Config, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal config: %w", err)
				}
				fmt.Println(string(jsonData))
				return nil
			}

			// Output config as human-readable format
			fmt.Println("binmate Configuration")
			fmt.Println("====================")
			fmt.Printf("Binaries: %d\n", len(Config.Binaries))
			fmt.Println()

			if len(Config.Binaries) > 0 {
				fmt.Printf("%-20s %-15s %s\n", "ID", "Name", "Provider")
				fmt.Println("---")
				for _, b := range Config.Binaries {
					fmt.Printf("%-20s %-15s %s:%s\n", b.Id, b.Name, b.Provider, b.Path)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showJSON, "json", false, "Output configuration as JSON")

	return cmd
}
