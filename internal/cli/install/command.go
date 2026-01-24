package install

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install",
		Short:   "Install a new binary version",
		Aliases: []string{"i", "add"},
		Run: func(cmd *cobra.Command, args []string) {
			binary, err := cmd.Flags().GetString("binary")
			if err != nil {
				panic("binary is required")
			}

			version, err := cmd.Flags().GetString("version")
			if err != nil {
				panic("version is required")
			}

			fmt.Printf("installing binary: %s version: %s", binary, version)
		},
	}

	cmd.Flags().String("binary", "", "binary to be installed")
	cmd.Flags().String("version", "", "version of the binary to be installed")

	return cmd
}
