package install

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extractTar extracts the given tar.gz archive into destDir using only the Go standard library.
func extractTar(srcTar string, destDir string) (string, error) {
	if destDir == "" {
		return "", fmt.Errorf("destination directory is required")
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

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar entry: %w", err)
		}

		if err := extractTarEntry(tr, header, destDir); err != nil {
			return "", err
		}
	}

	return destDir, nil
}

func extractTarEntry(tr *tar.Reader, header *tar.Header, destDir string) error {
	cleanName := filepath.Clean(header.Name)
	if filepath.IsAbs(cleanName) {
		return fmt.Errorf("absolute paths are not allowed in archive: %s", header.Name)
	}

	targetPath := filepath.Join(destDir, cleanName)
	destDirWithSep := destDir + string(os.PathSeparator)
	if targetPath != destDir && !strings.HasPrefix(targetPath, destDirWithSep) {
		return fmt.Errorf("illegal path traversal in archive entry: %s", header.Name)
	}

	// We disallow symlinks to avoid unexpected filesystem writes.
	if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
		return fmt.Errorf("symlinks are not supported in archive: %s", header.Name)
	}

	if header.Typeflag == tar.TypeDir {
		if err := os.MkdirAll(targetPath, os.FileMode(header.Mode).Perm()); err != nil {
			return fmt.Errorf("create dir %s: %w", targetPath, err)
		}
		return nil
	}

	if header.Typeflag != tar.TypeReg {
		// Skip special file types we don't support
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create parent dirs for %s: %w", targetPath, err)
	}

	mode := os.FileMode(header.Mode)
	if mode == 0 {
		mode = 0o644
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
