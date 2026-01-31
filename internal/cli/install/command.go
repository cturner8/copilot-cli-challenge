package install

import (
	"log"
	"time"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/core/install"
	v "cturner8/binmate/internal/core/version"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
	"cturner8/binmate/internal/providers/github"
)

// Package variable will be set by cmd package
var (
	Config    *config.Config
	DBService *repository.Service
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install",
		Short:   "Install a new binary version",
		Aliases: []string{"i", "add"},
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()

			id, err := DBService.Logs.LogStart("install", "", "", "start install process")
			if err != nil {
				log.Fatalf("sync start error: %s", err)
			}

			binary, err := cmd.Flags().GetString("binary")
			if err != nil || binary == "" {
				msg := "binary is required"
				log.Panic(msg)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			version, err := cmd.Flags().GetString("version")
			if err != nil || version == "" {
				msg := "version is required"
				log.Panic(msg)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			log.Printf("installing binary: %s version: %s", binary, version)

			DBService.Logs.LogEntity(id, binary, version)

			binaryConfig, err := DBService.Binaries.GetByUserID(binary)

			if err != nil {
				msg := "unable to find requested binary config"
				log.Panic(msg)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			if binaryConfig.Provider != "github" {
				msg := "only github provider is currently supported"
				log.Panic(msg)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			release, asset, err := github.FetchReleaseAsset(binaryConfig, version)
			if err != nil {
				msg := "fetch failed"
				log.Panic(msg, err)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			downloadPath, err := github.DownloadAsset(asset.BrowserDownloadUrl, asset.Name)
			if err != nil {
				msg := "download failed"
				log.Panic(msg, err)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			resolvedVersion := version
			if version == "latest" {
				resolvedVersion = release.TagName
			}

			destPath, err := install.ExtractAsset(downloadPath, binaryConfig, resolvedVersion)
			if err != nil {
				msg := "error extracting asset"
				log.Panic(msg, err)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			// Handle optional InstallPath
			customInstallPath := ""
			if binaryConfig.InstallPath != nil {
				customInstallPath = *binaryConfig.InstallPath
			}

			installPath, err := v.SetActiveVersion(destPath, customInstallPath, binaryConfig.Name)
			if err != nil {
				msg := "error setting active version"
				log.Panic(msg, err)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			log.Printf("downloaded binary: %s version: %s", binary, resolvedVersion)

			installation := &database.Installation{
				ID:                0,
				Version:           version,
				InstalledPath:     installPath,
				InstalledAt:       time.Now().Unix(),
				BinaryID:          binaryConfig.ID,
				SourceURL:         asset.BrowserDownloadUrl,
				Checksum:          asset.Digest, // TODO
				ChecksumAlgorithm: asset.Digest, // TODO
				FileSize:          int64(asset.Size),
			}

			err = DBService.Installations.Create(installation)
			if err != nil {
				msg := "error saving installation"
				log.Panic(msg, err)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			err = DBService.Versions.Set(binaryConfig.ID, installation.ID, installPath)
			if err != nil {
				msg := "error saving version"
				log.Panic(msg, err)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			if err != nil {
				msg := "error saving installation"
				log.Panic(msg, err)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
			}

			DBService.Logs.LogSuccess(id, int64(time.Since(start)))
		},
	}

	cmd.Flags().String("binary", "", "binary to be installed")
	cmd.Flags().String("version", "latest", "version of the binary to be installed")

	return cmd
}
