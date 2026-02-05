package github

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// AssetFilter represents criteria for filtering release assets
type AssetFilter struct {
	OS         string // e.g., "linux", "darwin", "windows"
	Arch       string // e.g., "amd64", "arm64"
	Extension  string // e.g., ".tar.gz", ".zip"
	Prefix     string // optional prefix to match
	AssetRegex string // optional regex pattern from config
}

// NewAssetFilter creates a filter with current platform defaults
func NewAssetFilter() AssetFilter {
	return AssetFilter{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// FilterAssets filters release assets based on the provided criteria
func FilterAssets(assets []ReleaseAsset, filter AssetFilter) ([]ReleaseAsset, error) {
	// If custom regex is provided, use it
	if filter.AssetRegex != "" {
		return filterByRegex(assets, filter.AssetRegex)
	}

	// Otherwise, use standard filtering
	filtered := assets

	// Filter by OS/platform
	if filter.OS != "" {
		filtered = filterByOS(filtered, filter.OS)
	}

	// Filter by architecture
	if filter.Arch != "" {
		filtered = filterByArch(filtered, filter.Arch)
	}

	// Filter by extension
	if filter.Extension != "" {
		filtered = filterByExtension(filtered, filter.Extension)
	}

	// Filter by prefix (optional)
	if filter.Prefix != "" {
		filtered = filterByPrefix(filtered, filter.Prefix)
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no assets matched filter criteria")
	}

	return filtered, nil
}

// filterByRegex filters assets using a custom regex pattern
func filterByRegex(assets []ReleaseAsset, pattern string) ([]ReleaseAsset, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	filtered := make([]ReleaseAsset, 0, len(assets))
	for _, asset := range assets {
		if re.MatchString(asset.Name) {
			filtered = append(filtered, asset)
		}
	}

	return filtered, nil
}

// filterByOS filters assets by operating system
func filterByOS(assets []ReleaseAsset, os string) []ReleaseAsset {
	filtered := make([]ReleaseAsset, 0, len(assets))

	// Common OS name variations in release asset names
	osPatterns := map[string][]string{
		"linux":   {"linux"},
		"darwin":  {"darwin", "macos", "osx"},
		"windows": {"windows", "win"},
	}

	patterns, exists := osPatterns[os]
	if !exists {
		patterns = []string{os}
	}

	for _, asset := range assets {
		nameLower := strings.ToLower(asset.Name)
		for _, pattern := range patterns {
			// Use case-insensitive matching
			if strings.Contains(nameLower, strings.ToLower(pattern)) {
				// edge case to prevent "win" matching on "darwin"
				if pattern == "win" && strings.Contains(nameLower, "darwin") {
					continue
				}
				filtered = append(filtered, asset)
				break
			}
		}
	}

	return filtered
}

// filterByArch filters assets by architecture
func filterByArch(assets []ReleaseAsset, arch string) []ReleaseAsset {
	filtered := make([]ReleaseAsset, 0, len(assets))

	// Common architecture name variations
	archPatterns := map[string][]string{
		"amd64": {"amd64", "x86_64", "x64"},
		"386":   {"i386", "i686", "386"},
		"arm64": {"arm64", "aarch64"},
		"arm":   {"arm", "armv7"},
	}

	patterns, exists := archPatterns[arch]
	if !exists {
		patterns = []string{arch}
	}

	for _, asset := range assets {
		nameLower := strings.ToLower(asset.Name)
		for _, pattern := range patterns {
			if strings.Contains(nameLower, strings.ToLower(pattern)) {
				filtered = append(filtered, asset)
				break
			}
		}
	}

	return filtered
}

// filterByExtension filters assets by file extension
func filterByExtension(assets []ReleaseAsset, ext string) []ReleaseAsset {
	filtered := make([]ReleaseAsset, 0, len(assets))

	// Ensure extension starts with a dot
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	for _, asset := range assets {
		// Handle multi-part extensions like .tar.gz
		assetName := asset.Name
		if strings.HasSuffix(assetName, ext) {
			filtered = append(filtered, asset)
			continue
		}

		// Also check standard extension
		if filepath.Ext(assetName) == ext {
			filtered = append(filtered, asset)
		}
	}

	return filtered
}

// filterByPrefix filters assets by filename prefix
func filterByPrefix(assets []ReleaseAsset, prefix string) []ReleaseAsset {
	filtered := make([]ReleaseAsset, 0, len(assets))

	for _, asset := range assets {
		if strings.HasPrefix(asset.Name, prefix) {
			filtered = append(filtered, asset)
		}
	}

	return filtered
}

// SelectBestAsset returns the most suitable asset from filtered results
// Prefers smaller files and more common archive formats
func SelectBestAsset(assets []ReleaseAsset) (ReleaseAsset, error) {
	if len(assets) == 0 {
		return ReleaseAsset{}, fmt.Errorf("no assets to select from")
	}

	if len(assets) == 1 {
		return assets[0], nil
	}

	// Prefer certain extensions in order
	preferredExts := []string{".tar.gz", ".tgz", ".zip", ".tar.xz", ".tar.bz2"}

	for _, ext := range preferredExts {
		var matching []ReleaseAsset
		for _, asset := range assets {
			if strings.HasSuffix(asset.Name, ext) {
				matching = append(matching, asset)
			}
		}

		if len(matching) > 0 {
			// Sort by name length to prefer simpler names
			// (assets without additional suffixes come first)
			best := matching[0]
			for _, asset := range matching[1:] {
				if len(asset.Name) < len(best.Name) {
					best = asset
				}
			}
			return best, nil
		}
	}

	// If no preferred extension found, return the shortest name
	best := assets[0]
	for _, asset := range assets[1:] {
		if len(asset.Name) < len(best.Name) {
			best = asset
		}
	}
	return best, nil
}
