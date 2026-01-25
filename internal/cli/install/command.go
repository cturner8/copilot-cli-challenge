package install

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/core/install"
	"cturner8/binmate/internal/providers/github"
)

func NewCommand(config config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install",
		Short:   "Install a new binary version",
		Aliases: []string{"i", "add"},
		Run: func(cmd *cobra.Command, args []string) {
			binary, err := cmd.Flags().GetString("binary")
			if err != nil || binary == "" {
				log.Panicf("binary is required")
			}

			version, err := cmd.Flags().GetString("version")
			if err != nil || version == "" {
				log.Panicf("version is required")
			}

			fmt.Printf("installing binary: %s version: %s", binary, version)

			downloadPath, err := github.DownloadAsset(binary, version, "bun-linux-aarch64.zip")
			if err != nil {
				log.Panicf("download failed: %s", err)
			}

			if err := install.ExtractAsset(downloadPath, fmt.Sprintf("/tmp/binmate/%s", binary)); err != nil {
				log.Panicf("error extracting asset: %s", err)
			}

			fmt.Printf("\ndownloaded binary: %s version: %s", binary, version)
		},
	}

	cmd.Flags().String("binary", "", "binary to be installed")
	cmd.Flags().String("version", "", "version of the binary to be installed")

	return cmd
}
