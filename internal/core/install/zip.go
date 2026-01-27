package install

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// extractZip extracts the specified binary from a ZIP archive into destDir.
// It searches for the binary by name within the archive (including subdirectories)
// and extracts only that file to destDir/binaryName.
func extractZip(srcZip string, destDir string, binaryName string) (string, error) {
	if destDir == "" {
		return "", fmt.Errorf("destination directory is required")
	}
	if binaryName == "" {
		return "", fmt.Errorf("binary name is required")
	}

	destDir = filepath.Clean(destDir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("create destination: %w", err)
	}

	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	// Search for the binary in the archive
	for _, f := range r.File {
		// Check if this is the binary we're looking for (could be in subdirectories like bin/gh)
		if filepath.Base(f.Name) == binaryName && !f.FileInfo().IsDir() {
			targetPath := filepath.Join(destDir, binaryName)
			if err := extractZipBinary(f, targetPath); err != nil {
				return "", err
			}
			return targetPath, nil
		}
	}

	return "", fmt.Errorf("binary %s not found in archive", binaryName)
}

func extractZipBinary(f *zip.File, targetPath string) error {
	// We disallow symlinks to avoid unexpected filesystem writes.
	if f.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symlinks are not supported in archive: %s", f.Name)
	}

	reader, err := f.Open()
	if err != nil {
		return fmt.Errorf("open file %s: %w", f.Name, err)
	}
	defer reader.Close()

	mode := f.Mode()
	if mode == 0 || mode.Perm() == 0 {
		mode = 0o755 // Binaries should be executable
	}

	dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return fmt.Errorf("create file %s: %w", targetPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return fmt.Errorf("write file %s: %w", targetPath, err)
	}

	return nil
}
