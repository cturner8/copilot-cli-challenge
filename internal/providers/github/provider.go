package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadAsset(path string, version string, assetName string) string {
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", path, version, assetName)

	resp, err := http.Get(url)
	if err != nil {
		panic("error downloading asset")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	fileName := fmt.Sprintf("/tmp/%s", assetName)
	os.WriteFile(fileName, body, 0755)

	return fileName
}
