package install

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func getExtractPath(binaryId string, version string) (string, error) {
	baseDir, err := getLocalDataDir()
	if err != nil {
		return "", fmt.Errorf("unable to locate asset extract dir: %w", err)
	}

	destDir := filepath.Join(baseDir, "binmate", "versions", binaryId, version)
	return destDir, nil
}

func getLocalDataDir() (string, error) {
	if runtime.GOOS == "windows" {
		return os.UserCacheDir()
	}

	xdgDataHome, xdgDataHomeSet := os.LookupEnv("XDG_DATA_HOME")
	if xdgDataHomeSet {
		return xdgDataHome, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}

	return filepath.Join(homeDir, ".local", "share"), nil
}
