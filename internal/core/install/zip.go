package install

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractZip extracts the given ZIP archive into destDir using only the Go standard library.
func extractZip(srcZip string, destDir string) error {
	if destDir == "" {
		return fmt.Errorf("destination directory is required")
	}

	destDir = filepath.Clean(destDir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create destination: %w", err)
	}

	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if err := extractZipEntry(f, destDir); err != nil {
			return err
		}
	}

	return nil
}

func extractZipEntry(f *zip.File, destDir string) error {
	cleanName := filepath.Clean(f.Name)
	if filepath.IsAbs(cleanName) {
		return fmt.Errorf("absolute paths are not allowed in archive: %s", f.Name)
	}

	targetPath := filepath.Join(destDir, cleanName)
	destDirWithSep := destDir + string(os.PathSeparator)
	if targetPath != destDir && !strings.HasPrefix(targetPath, destDirWithSep) {
		return fmt.Errorf("illegal path traversal in archive entry: %s", f.Name)
	}

	// We disallow symlinks to avoid unexpected filesystem writes.
	if f.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symlinks are not supported in archive: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(targetPath, f.Mode().Perm()); err != nil {
			return fmt.Errorf("create dir %s: %w", targetPath, err)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create parent dirs for %s: %w", targetPath, err)
	}

	reader, err := f.Open()
	if err != nil {
		return fmt.Errorf("open file %s: %w", f.Name, err)
	}
	defer reader.Close()

	mode := f.Mode()
	if mode == 0 {
		mode = 0o644
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
