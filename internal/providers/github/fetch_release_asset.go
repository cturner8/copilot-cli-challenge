package github

import (
	"cturner8/binmate/internal/database"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type ReleaseAsset struct {
	Id                 int    `json:"id"`
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	Size               int    `json:"size"`
	Digest             string `json:"digest"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type Release struct {
	Name    string         `json:"name"`
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

func FetchReleaseAsset(binary *database.Binary, version string) (Release, ReleaseAsset, error) {
	if binary.ProviderPath == "" {
		log.Panicln("path is required for binary config")
	}

	// default to latest release
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", binary.ProviderPath)
	if version != "latest" {
		tag := version
		// TODO: handle as actual regex expression?
		if binary.ReleaseRegex != nil && *binary.ReleaseRegex != "" {
			tag = *binary.ReleaseRegex + version
		}

		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", binary.ProviderPath, tag)
	}

	// Create HTTP client with optional authentication
	client, err := CreateHTTPClient(binary.Authenticated)
	if err != nil {
		return Release{}, ReleaseAsset{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	response, err := client.Get(url)
	if err != nil {
		return Release{}, ReleaseAsset{}, fmt.Errorf("download asset: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return Release{}, ReleaseAsset{}, fmt.Errorf("download asset: unexpected status %s", response.Status)
	}

	contentType := response.Header.Get("content-type")
	if !strings.Contains(contentType, "application/json") {
		return Release{}, ReleaseAsset{}, fmt.Errorf("invalid release response content: %s", contentType)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Release{}, ReleaseAsset{}, fmt.Errorf("failed to read body: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return Release{}, ReleaseAsset{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(release.Assets) == 0 {
		return Release{}, ReleaseAsset{}, fmt.Errorf("failed to find requested binary, no release assets")
	}

	// Create filter based on binary config
	filter := NewAssetFilter()
	filter.Extension = binary.Format // e.g., ".tar.gz", ".zip"
	if binary.AssetRegex != nil {
		filter.AssetRegex = *binary.AssetRegex // custom regex if provided
	}

	// Filter assets based on platform, architecture, and format
	filteredAssets, err := FilterAssets(release.Assets, filter)
	if err != nil {
		return Release{}, ReleaseAsset{}, fmt.Errorf("no matching assets found: %w", err)
	}

	// Select the best asset from filtered results
	selectedAsset, err := SelectBestAsset(filteredAssets)
	if err != nil {
		return Release{}, ReleaseAsset{}, fmt.Errorf("failed to select asset: %w", err)
	}

	return release, selectedAsset, nil
}
