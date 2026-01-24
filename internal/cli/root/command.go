package root

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "binmate",
		Short: "An application for managing binary installations from remote repositories.",
		Long:  `An application for managing binary installations from remote repositories.`,
		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(initialModel())
			if _, err := p.Run(); err != nil {
				fmt.Printf("error: %v", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
