package add

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	binarySvc "cturner8/binmate/internal/core/binary"
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database/repository"
)

// Package variables will be set by cmd package
var (
	Config    *config.Config
	DBService *repository.Service
)

func NewCommand() *cobra.Command {
	var (
		url string
	)

	cmd := &cobra.Command{
		Use:           "add [binary-id] [url]",
		Short:         "Add a new binary from GitHub release URL or config",
		Long: `Add a new binary to binmate.

You can add a binary in two ways:
1. Provide a GitHub release URL: binmate add https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz
2. Reference a binary from config: binmate add <binary-id>

The binary will be registered in the database but not installed until you run 'binmate install'.`,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Case 1: URL provided as flag
			if url != "" {
				binary, err := binarySvc.AddBinaryFromURL(url, DBService)
				if err != nil {
					return fmt.Errorf("failed to add binary from URL: %w", err)
				}
				log.Printf("✓ Binary %s added successfully", binary.UserID)
				return nil
			}

			// Case 2: URL provided as first argument
			if len(args) > 0 && len(args[0]) > 0 {
				// Check if it's a URL (starts with http)
				if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") {
					binary, err := binarySvc.AddBinaryFromURL(args[0], DBService)
					if err != nil {
						return fmt.Errorf("failed to add binary from URL: %w", err)
					}
					log.Printf("✓ Binary %s added successfully", binary.UserID)
					return nil
				}

				// Otherwise treat as binary ID from config
				binaryID := args[0]
				if err := config.SyncBinary(binaryID, *Config, DBService); err != nil {
					return fmt.Errorf("failed to add binary from config: %w", err)
				}
				log.Printf("✓ Binary %s added from config", binaryID)
				return nil
			}

			return fmt.Errorf("please provide either a GitHub release URL or a binary ID from config")
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "GitHub release URL for the binary")

	return cmd
}
