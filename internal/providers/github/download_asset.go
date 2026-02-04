package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadAsset(url string, assetName string) (string, error) {
	response, err := http.Get(url)
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
