package install

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/core/crypto"
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
	var (
		binary  string
		version string
	)

	cmd := &cobra.Command{
		Use:           "install",
		Short:         "Install a new binary version",
		Aliases:       []string{"i", "add"},
		SilenceUsage:  true,  // Don't show usage on runtime errors
		SilenceErrors: false, // Still print errors
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Sync the specific binary from config to database
			if err := config.SyncBinary(binary, *Config, DBService); err != nil {
				return fmt.Errorf("failed to sync binary config: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			id, err := DBService.Logs.LogStart("install", "", "", "start install process")
			if err != nil {
				return fmt.Errorf("sync start error: %w", err)
			}

			log.Printf("installing binary: %s version: %s", binary, version)

			DBService.Logs.LogEntity(id, binary, version)

			binaryConfig, err := DBService.Binaries.GetByUserID(binary)
			if err != nil {
				msg := "unable to find requested binary config"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s: %w", msg, err)
			}

			if binaryConfig.Provider != "github" {
				msg := "only github provider is currently supported"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s", msg)
			}

			release, asset, err := github.FetchReleaseAsset(binaryConfig, version)
			if err != nil {
				msg := "fetch failed"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s: %w", msg, err)
			}

			downloadPath, err := github.DownloadAsset(asset.BrowserDownloadUrl, asset.Name)
			if err != nil {
				msg := "download failed"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s: %w", msg, err)
			}

			// Verify downloaded archive checksum if digest is provided
			if asset.Digest != "" {
				if err := crypto.VerifyDigest(downloadPath, asset.Digest); err != nil {
					msg := fmt.Sprintf("checksum verification failed: %v", err)
					DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
					return fmt.Errorf("%s", msg)
				}
				log.Printf("✓ archive checksum verified")
			}

			resolvedVersion := version
			if version == "latest" {
				resolvedVersion = release.TagName
			}

			destPath, err := install.ExtractAsset(downloadPath, binaryConfig, resolvedVersion)
			if err != nil {
				msg := "error extracting asset"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s: %w", msg, err)
			}

			// Compute checksum of extracted binary
			binaryChecksum, err := crypto.ComputeSHA256(destPath)
			if err != nil {
				msg := fmt.Sprintf("failed to compute binary checksum: %v", err)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s", msg)
			}

			// Get file size of extracted binary
			fileInfo, err := os.Stat(destPath)
			if err != nil {
				msg := fmt.Sprintf("failed to get file info: %v", err)
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s", msg)
			}

			// Handle optional InstallPath
			customInstallPath := ""
			if binaryConfig.InstallPath != nil {
				customInstallPath = *binaryConfig.InstallPath
			}

			installPath, err := v.SetActiveVersion(destPath, customInstallPath, binaryConfig.Name)
			if err != nil {
				msg := "error setting active version"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s: %w", msg, err)
			}

			log.Printf("downloaded binary: %s version: %s", binary, resolvedVersion)

			installation := &database.Installation{
				ID:                0,
				Version:           resolvedVersion,
				InstalledPath:     installPath,
				InstalledAt:       time.Now().Unix(),
				BinaryID:          binaryConfig.ID,
				SourceURL:         asset.BrowserDownloadUrl,
				Checksum:          binaryChecksum,
				ChecksumAlgorithm: "SHA256",
				FileSize:          fileInfo.Size(),
			}

			if err := DBService.Installations.Create(installation); err != nil {
				msg := "error saving installation"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s: %w", msg, err)
			}

			if err := DBService.Versions.Set(binaryConfig.ID, installation.ID, installPath); err != nil {
				msg := "error saving version"
				DBService.Logs.LogFailure(id, msg, int64(time.Since(start)))
				return fmt.Errorf("%s: %w", msg, err)
			}

			DBService.Logs.LogSuccess(id, int64(time.Since(start)))
			return nil
		},
	}

	cmd.Flags().StringVarP(&binary, "binary", "b", "", "binary to be installed")
	cmd.Flags().StringVarP(&version, "version", "v", "latest", "version of the binary to be installed")

	// Mark required flags
	cmd.MarkFlagRequired("binary")

	return cmd
}
