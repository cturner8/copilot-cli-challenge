/*
Copyright © 2026 cturner8
*/
package cmd

import (
	"os"

	"cturner8/binmate/internal/cli/install"
	"cturner8/binmate/internal/cli/root"
	"cturner8/binmate/internal/core/config"
)

var rootCmd = root.NewCommand()

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	config := config.ReadConfig()

	rootCmd.AddCommand(install.NewCommand(config))
}
