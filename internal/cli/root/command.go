package root

import (
	"fmt"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/tui"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "binmate",
		Short:         "An application for managing binary installations from remote repositories.",
		Long:          `An application for managing binary installations from remote repositories.`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := tui.InitProgram()
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}
			return nil
		},
	}

	return cmd
}
