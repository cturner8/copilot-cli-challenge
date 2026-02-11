package tui

import (
	"fmt"
	"regexp"
	"strings"

	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

const (
	// noActiveVersion is the display value when a binary has no active version
	noActiveVersion = "none"
)

// BinaryWithMetadata represents a binary with additional metadata
type BinaryWithMetadata struct {
	Binary             *database.Binary
	ActiveVersion      string
	InstallCount       int
	ActiveInstallation *database.Installation
}

// getBinariesWithMetadata fetches all binaries with their metadata using the repository method
func getBinariesWithMetadata(dbService *repository.Service) ([]BinaryWithMetadata, error) {
	// Use the repository method to fetch binaries with version details
	details, err := dbService.Binaries.ListWithVersionDetails(noActiveVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get binaries with metadata: %w", err)
	}

	// Convert to TUI metadata type
	result := make([]BinaryWithMetadata, len(details))
	for i, detail := range details {
		result[i] = BinaryWithMetadata{
			Binary:             detail.Binary,
			ActiveVersion:      detail.ActiveVersion,
			InstallCount:       detail.InstallCount,
			ActiveInstallation: detail.ActiveInstallation,
		}
	}

	return result, nil
}

// getVersionsForBinary fetches all installations for a binary ordered by date
func getVersionsForBinary(dbService *repository.Service, binaryID int64) ([]*database.Installation, error) {
	installations, err := dbService.Installations.ListByBinary(binaryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list installations: %w", err)
	}
	return installations, nil
}

// getActiveVersion gets the currently active version for a binary
func getActiveVersion(dbService *repository.Service, binaryID int64) (*database.Installation, error) {
	_, installation, err := dbService.Versions.GetWithInstallation(binaryID)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}
	return installation, nil
}

// filterBinaries filters binaries by name using regex pattern
func filterBinaries(binaries []BinaryWithMetadata, pattern string) []BinaryWithMetadata {
	if pattern == "" {
		return binaries
	}

	// Try to compile pattern as regex
	// If it fails, treat as literal string
	var filtered []BinaryWithMetadata
	for _, binary := range binaries {
		// Try regex match first
		matched := false

		// For simple patterns without regex special chars, do case-insensitive substring match
		if !containsRegexChars(pattern) {
			matched = strings.Contains(strings.ToLower(binary.Binary.Name), strings.ToLower(pattern)) ||
				strings.Contains(strings.ToLower(binary.Binary.UserID), strings.ToLower(pattern))
		} else {
			// Try regex matching
			if re, err := regexp.Compile("(?i)" + pattern); err == nil {
				matched = re.MatchString(binary.Binary.Name) || re.MatchString(binary.Binary.UserID)
			}
		}

		if matched {
			filtered = append(filtered, binary)
		}
	}
	return filtered
}

// containsRegexChars checks if string contains regex special characters
func containsRegexChars(s string) bool {
	specialChars := []string{".", "*", "+", "?", "^", "$", "[", "]", "(", ")", "{", "}", "|", "\\"}
	for _, char := range specialChars {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}
