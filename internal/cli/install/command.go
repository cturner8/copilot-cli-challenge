package install

import (
	"log"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/core/install"
	"cturner8/binmate/internal/providers/github"
)

func NewCommand(c config.Config) *cobra.Command {
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

			log.Printf("installing binary: %s version: %s", binary, version)

			binaryConfig, err := config.GetBinary(binary, c.Binaries)
			if err != nil {
				log.Panicf("unable to find requested binary config")
			}

			assetName, downloadUrl, err := github.FetchReleaseAsset(binaryConfig, version)
			if err != nil {
				log.Panicf("fetch failed: %s", err)
			}

			downloadPath, err := github.DownloadAsset(downloadUrl, assetName)
			if err != nil {
				log.Panicf("download failed: %s", err)
			}

			destPath, err := install.ExtractAsset(downloadPath, binaryConfig.Id, version)
			if err != nil {
				log.Panicf("error extracting asset: %s", err)
			}

			log.Printf("downloaded binary: %s version: %s", binary, version)
			log.Printf(destPath)
		},
	}

	cmd.Flags().String("binary", "", "binary to be installed")
	cmd.Flags().String("version", "latest", "version of the binary to be installed")

	return cmd
}
