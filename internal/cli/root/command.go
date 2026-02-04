package root

import (
	"fmt"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database/repository"
	"cturner8/binmate/internal/tui"
)

// Package variables will be set by cmd package
var (
	Config    *config.Config
	DBService *repository.Service
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "binmate",
		Short:         "An application for managing binary installations from remote repositories.",
		Long:          `An application for managing binary installations from remote repositories.`,
		SilenceUsage:  true,
		SilenceErrors: false,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Sync config to database before launching TUI
			if err := config.SyncToDatabase(*Config, DBService); err != nil {
				return fmt.Errorf("sync error: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			p := tui.InitProgram(DBService, Config)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}
			return nil
		},
	}

	return cmd
}
