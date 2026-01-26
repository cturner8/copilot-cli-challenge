package install

import (
	"log"
	"os"
	"path/filepath"
)

func ExtractAsset(srcPath string, id string, version string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Panicf("unable to locate asset cache dir")
	}

	destDir := filepath.Join(cacheDir, "binmate", id, version)
	// TODO: check file type
	return extractZip(srcPath, destDir)
}
