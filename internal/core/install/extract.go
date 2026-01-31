package install

import (
	"cturner8/binmate/internal/database"
	"fmt"
)

func ExtractAsset(srcPath string, binary *database.Binary, version string) (string, error) {
	destDir, err := getExtractPath(binary.UserID, version)
	if err != nil {
		return "", fmt.Errorf("unable to locate asset extract dir")
	}

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
