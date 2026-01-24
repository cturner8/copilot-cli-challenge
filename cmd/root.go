/*
Copyright © 2026 cturner8
*/
package cmd

import (
	"os"

	"cturner8/binmate/internal/cli/install"
	"cturner8/binmate/internal/cli/root"
)

var rootCmd = root.NewCommand()

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(install.NewCommand())
}
