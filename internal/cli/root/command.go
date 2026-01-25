package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/tui"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "binmate",
		Short: "An application for managing binary installations from remote repositories.",
		Long:  `An application for managing binary installations from remote repositories.`,
		Run: func(cmd *cobra.Command, args []string) {
			p := tui.InitProgram()
			if _, err := p.Run(); err != nil {
				fmt.Printf("error: %v", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
