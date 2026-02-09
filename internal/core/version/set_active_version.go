package version

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func SetActiveVersion(versionPath string, initialInstallPath string, binaryName string) (string, error) {
	installPath, err := getInstallBinPath(initialInstallPath)
	if err != nil {
		return "", fmt.Errorf("unable to resolve install path: %s", err)

	}

	mode := os.FileMode(0o755)
	err = os.MkdirAll(installPath, mode)
	if err != nil {
		return "", fmt.Errorf("unable to create install path: %s", err)
	}

	targetInstallPath := path.Join(installPath, binaryName)

	// Remove existing symlink/file if present
	_, err = os.Lstat(targetInstallPath)
	if err == nil {
		// File exists, remove it
		if err := os.Remove(targetInstallPath); err != nil {
			return "", fmt.Errorf("unable to remove existing symlink: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		// Error other than "file not found" - could be permission issue, etc.
		return "", fmt.Errorf("unable to check existing symlink: %w", err)
	}

	err = os.Symlink(versionPath, targetInstallPath)
	if err != nil {
		return "", fmt.Errorf("unable to create symlink: %w", err)
	}

	return targetInstallPath, nil
}
