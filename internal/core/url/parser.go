package url

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// ParsedGitHubRelease represents a parsed GitHub release URL
type ParsedGitHubRelease struct {
	Owner     string
	Repo      string
	Version   string
	AssetName string
	Format    string
}

// ParseGitHubReleaseURL parses a GitHub release URL and extracts metadata
// Expected format: https://github.com/owner/repo/releases/download/version/asset-name.tar.gz
func ParseGitHubReleaseURL(rawURL string) (*ParsedGitHubRelease, error) {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Validate it's a GitHub URL
	if parsedURL.Host != "github.com" {
		return nil, fmt.Errorf("not a GitHub URL: %s", parsedURL.Host)
	}

	// Split the path into segments
	pathSegments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	// Expected format: owner/repo/releases/download/version/asset-name
	if len(pathSegments) < 6 {
		return nil, fmt.Errorf("invalid GitHub release URL format: expected at least 6 path segments, got %d", len(pathSegments))
	}

	// Validate the URL structure
	if pathSegments[2] != "releases" || pathSegments[3] != "download" {
		return nil, fmt.Errorf("invalid GitHub release URL: expected /releases/download/ in path")
	}

	owner := pathSegments[0]
	repo := pathSegments[1]
	version := pathSegments[4]
	assetName := pathSegments[5]

	// Detect format from file extension
	format := detectFormat(assetName)
	if format == "" {
		return nil, fmt.Errorf("unsupported file format for asset: %s", assetName)
	}

	return &ParsedGitHubRelease{
		Owner:     owner,
		Repo:      repo,
		Version:   version,
		AssetName: assetName,
		Format:    format,
	}, nil
}

// detectFormat detects the archive format from the file name
func detectFormat(fileName string) string {
	ext := path.Ext(fileName)
	
	// Check for .tar.gz
	if strings.HasSuffix(fileName, ".tar.gz") {
		return ".tar.gz"
	}
	
	// Check for other common formats
	switch ext {
	case ".zip":
		return ".zip"
	case ".tgz":
		return ".tar.gz"
	default:
		return ""
	}
}

// GenerateBinaryID generates a user ID from the asset name
// Extracts the prefix before platform/arch identifiers
func GenerateBinaryID(assetName string) string {
	// Remove extension
	name := strings.TrimSuffix(assetName, ".tar.gz")
	name = strings.TrimSuffix(name, ".zip")
	name = strings.TrimSuffix(name, ".tgz")
	
	// Split by common separators
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_'
	})
	
	if len(parts) > 0 {
		return parts[0]
	}
	
	return name
}

// GenerateBinaryName generates a binary name from the asset name
// Uses the same logic as GenerateBinaryID for consistency
func GenerateBinaryName(assetName string) string {
	return GenerateBinaryID(assetName)
}
