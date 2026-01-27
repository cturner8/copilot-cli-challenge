package install

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExtractAsset(srcPath string, id string, version string, format string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("unable to locate asset cache dir")
	}

	destDir := filepath.Join(cacheDir, ".binmate", id, version)

	switch format {
	case ".zip":
		{
			return extractZip(srcPath, destDir)
		}
	case ".tar.gz":
		{
			return extractTar(srcPath, destDir)
		}
	}

	return "", fmt.Errorf("unsupported asset format: %s", format)
}
