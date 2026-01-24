package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "binmate",
		Short: "An application for managing binary installations from remote repositories.",
		Long:  `An application for managing binary installations from remote repositories.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("starting tui")
		},
	}

	return cmd
}
