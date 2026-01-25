package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadAsset(path string, version string, assetName string) (string, error) {
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", path, version, assetName)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download asset: unexpected status %s", resp.Status)
	}

	destPath := filepath.Join(os.TempDir(), assetName)
	tmp, err := os.CreateTemp(os.TempDir(), assetName+".*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", fmt.Errorf("write asset: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmp.Name(), destPath); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("finalise asset: %w", err)
	}

	return destPath, nil
}
