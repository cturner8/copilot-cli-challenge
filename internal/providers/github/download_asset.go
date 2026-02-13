package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadAsset(providerPath string, assetId int, assetName string, authenticated bool) (string, error) {
	// Create HTTP client with optional authentication
	client, err := CreateHTTPClient(authenticated)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Get the asset via the GitHub API rather than the `BrowserDownloadUrl` to support authentication.
	// BrowserDownloadUrl is a `github.com` URL which does not accept a bearer token.
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%d", providerPath, assetId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set to `application/octet-stream` to return asset content directly
	req.Header.Set("Accept", "application/octet-stream")

	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download asset: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download asset: unexpected status %s", response.Status)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("unable to locate cache directory: %w", err)
	}

	basePath := filepath.Join(cacheDir, "binmate")
	destPath := filepath.Join(basePath, assetName)
	tmp, err := os.CreateTemp(os.TempDir(), assetName+".*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	if _, err := io.Copy(tmp, response.Body); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", fmt.Errorf("write asset: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("close temp file: %w", err)
	}

	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return "", fmt.Errorf("create destination path: %w", err)
	}
	if err := os.Rename(tmp.Name(), destPath); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("finalise asset: %w", err)
	}

	return destPath, nil
}
