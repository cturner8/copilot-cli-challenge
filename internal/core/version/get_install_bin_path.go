package version

import (
	"fmt"
	"os"
	"path"
	"runtime"
)

func getInstallBinPath(userInstallPath string) (string, error) {
	installPath := userInstallPath
	if installPath == "" {
		if runtime.GOOS == "windows" {
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				return "", fmt.Errorf("unable to find user cache directory: %s", err)
			}

			installPath = path.Join(cacheDir, "binmate", "bin")
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("unable to find user home directory: %s", err)
			}

			installPath = path.Join(homeDir, ".local", "bin")
		}
	}

	return installPath, nil
}
