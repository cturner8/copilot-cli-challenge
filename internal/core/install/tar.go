package install

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// extractTar extracts the specified binary from a tar.gz archive into destDir.
// It searches for the binary by name within the archive (including subdirectories)
// and extracts only that file to destDir/binaryName.
func extractTar(srcTar string, destDir string, binaryName string) (string, error) {
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

	f, err := os.Open(srcTar)
	if err != nil {
		return "", fmt.Errorf("open tar: %w", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Search for the binary in the archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar entry: %w", err)
		}

		// Check if this is the binary we're looking for (could be in subdirectories like bin/gh)
		if filepath.Base(header.Name) == binaryName && header.Typeflag == tar.TypeReg {
			targetPath := filepath.Join(destDir, binaryName)
			if err := extractTarBinary(tr, header, targetPath); err != nil {
				return "", err
			}
			return targetPath, nil
		}
	}

	return "", fmt.Errorf("binary %s not found in archive", binaryName)
}

func extractTarBinary(tr *tar.Reader, header *tar.Header, targetPath string) error {
	// We disallow symlinks to avoid unexpected filesystem writes.
	if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
		return fmt.Errorf("symlinks are not supported in archive: %s", header.Name)
	}

	mode := os.FileMode(header.Mode)
	if mode == 0 || mode.Perm() == 0 {
		mode = 0o755 // Binaries should be executable
	}

	dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return fmt.Errorf("create file %s: %w", targetPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, tr); err != nil {
		return fmt.Errorf("write file %s: %w", targetPath, err)
	}

	return nil
}
