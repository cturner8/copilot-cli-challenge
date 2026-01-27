package install

import (
	"cturner8/binmate/internal/core/config"
	"fmt"
	"os"
	"path/filepath"
)

func ExtractAsset(srcPath string, binary config.Binary, version string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("unable to locate asset cache dir")
	}

	destDir := filepath.Join(cacheDir, ".binmate", binary.Id, version)

	switch binary.Format {
	case ".zip":
		{
			return extractZip(srcPath, destDir, binary.Name)
		}
	case ".tar.gz":
		{
			return extractTar(srcPath, destDir, binary.Name)
		}
	}

	return "", fmt.Errorf("unsupported asset format: %s", binary.Format)
}
