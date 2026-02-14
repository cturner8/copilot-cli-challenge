package install

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"cturner8/binmate/internal/database"
)

// Test helpers for creating test archives

// createTestTarGz creates a tar.gz archive with test files
func createTestTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar.gz")
	
	f, err := os.Create(tarPath)
	if err != nil {
		t.Fatalf("Failed to create tar file: %v", err)
	}
	defer f.Close()
	
	gw := gzip.NewWriter(f)
	defer gw.Close()
	
	tw := tar.NewWriter(gw)
	defer tw.Close()
	
	for name, content := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0755,
			Size: int64(len(content)),
			Typeflag: tar.TypeReg,
		}
		
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("Failed to write tar header: %v", err)
		}
		
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write tar content: %v", err)
		}
	}
	
	return tarPath
}

// createTestZip creates a zip archive with test files
func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()
	
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer f.Close()
	
	zw := zip.NewWriter(f)
	defer zw.Close()
	
	for name, content := range files {
		fw, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		})
		if err != nil {
			t.Fatalf("Failed to create zip header: %v", err)
		}
		
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write zip content: %v", err)
		}
	}
	
	return zipPath
}

// TestExtractTar tests tar.gz extraction
func TestExtractTar_Success(t *testing.T) {
	// Create test tar.gz with binary
	files := map[string]string{
		"testbin": "#!/bin/bash\necho 'test binary'",
	}
	tarPath := createTestTarGz(t, files)
	destDir := t.TempDir()
	
	// Extract
	extractedPath, err := extractTar(tarPath, destDir, "testbin")
	if err != nil {
		t.Fatalf("extractTar failed: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
		t.Errorf("Extracted file does not exist: %s", extractedPath)
	}
	
	// Verify content
	content, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	
	expected := files["testbin"]
	if string(content) != expected {
		t.Errorf("Content mismatch:\nExpected: %s\nGot: %s", expected, string(content))
	}
	
	// Verify executable permissions
	info, err := os.Stat(extractedPath)
	if err != nil {
		t.Fatalf("Failed to stat extracted file: %v", err)
	}
	if info.Mode().Perm()&0100 == 0 {
		t.Error("Extracted binary should be executable")
	}
}

func TestExtractTar_BinaryInSubdirectory(t *testing.T) {
	// Create tar.gz with binary in subdirectory
	files := map[string]string{
		"bin/mybin": "binary content",
	}
	tarPath := createTestTarGz(t, files)
	destDir := t.TempDir()
	
	// Extract (should find mybin even though it's in bin/)
	extractedPath, err := extractTar(tarPath, destDir, "mybin")
	if err != nil {
		t.Fatalf("extractTar failed: %v", err)
	}
	
	// Verify extracted to root of destDir, not bin/
	expectedPath := filepath.Join(destDir, "mybin")
	if extractedPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, extractedPath)
	}
	
	// Verify content
	content, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	if string(content) != "binary content" {
		t.Errorf("Content mismatch: got %s", string(content))
	}
}

func TestExtractTar_BinaryNotFound(t *testing.T) {
	files := map[string]string{
		"other": "content",
	}
	tarPath := createTestTarGz(t, files)
	destDir := t.TempDir()
	
	_, err := extractTar(tarPath, destDir, "missing")
	if err == nil {
		t.Error("Expected error for missing binary, got none")
	}
	if err != nil && err.Error() != "binary missing not found in archive" {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestExtractTar_InvalidTarFile(t *testing.T) {
	// Create invalid tar file
	tmpDir := t.TempDir()
	invalidTar := filepath.Join(tmpDir, "invalid.tar.gz")
	if err := os.WriteFile(invalidTar, []byte("not a tar file"), 0644); err != nil {
		t.Fatalf("Failed to create invalid tar: %v", err)
	}
	
	destDir := t.TempDir()
	_, err := extractTar(invalidTar, destDir, "anybin")
	if err == nil {
		t.Error("Expected error for invalid tar file, got none")
	}
}

func TestExtractTar_EmptyDestDir(t *testing.T) {
	files := map[string]string{
		"testbin": "content",
	}
	tarPath := createTestTarGz(t, files)
	
	_, err := extractTar(tarPath, "", "testbin")
	if err == nil {
		t.Error("Expected error for empty destDir, got none")
	}
}

func TestExtractTar_EmptyBinaryName(t *testing.T) {
	files := map[string]string{
		"testbin": "content",
	}
	tarPath := createTestTarGz(t, files)
	destDir := t.TempDir()
	
	_, err := extractTar(tarPath, destDir, "")
	if err == nil {
		t.Error("Expected error for empty binary name, got none")
	}
}

func TestExtractTar_NonExistentFile(t *testing.T) {
	destDir := t.TempDir()
	_, err := extractTar("/nonexistent/file.tar.gz", destDir, "testbin")
	if err == nil {
		t.Error("Expected error for nonexistent tar file, got none")
	}
}

// TestExtractZip tests zip extraction
func TestExtractZip_Success(t *testing.T) {
	files := map[string]string{
		"testbin": "#!/bin/bash\necho 'test binary'",
	}
	zipPath := createTestZip(t, files)
	destDir := t.TempDir()
	
	extractedPath, err := extractZip(zipPath, destDir, "testbin")
	if err != nil {
		t.Fatalf("extractZip failed: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
		t.Errorf("Extracted file does not exist: %s", extractedPath)
	}
	
	// Verify content
	content, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	
	expected := files["testbin"]
	if string(content) != expected {
		t.Errorf("Content mismatch:\nExpected: %s\nGot: %s", expected, string(content))
	}
}

func TestExtractZip_BinaryInSubdirectory(t *testing.T) {
	files := map[string]string{
		"bin/mybin": "binary content",
	}
	zipPath := createTestZip(t, files)
	destDir := t.TempDir()
	
	extractedPath, err := extractZip(zipPath, destDir, "mybin")
	if err != nil {
		t.Fatalf("extractZip failed: %v", err)
	}
	
	// Verify extracted to root of destDir
	expectedPath := filepath.Join(destDir, "mybin")
	if extractedPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, extractedPath)
	}
	
	// Verify content
	content, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	if string(content) != "binary content" {
		t.Errorf("Content mismatch: got %s", string(content))
	}
}

func TestExtractZip_BinaryNotFound(t *testing.T) {
	files := map[string]string{
		"other": "content",
	}
	zipPath := createTestZip(t, files)
	destDir := t.TempDir()
	
	_, err := extractZip(zipPath, destDir, "missing")
	if err == nil {
		t.Error("Expected error for missing binary, got none")
	}
	if err != nil && err.Error() != "binary missing not found in archive" {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestExtractZip_InvalidZipFile(t *testing.T) {
	// Create invalid zip file
	tmpDir := t.TempDir()
	invalidZip := filepath.Join(tmpDir, "invalid.zip")
	if err := os.WriteFile(invalidZip, []byte("not a zip file"), 0644); err != nil {
		t.Fatalf("Failed to create invalid zip: %v", err)
	}
	
	destDir := t.TempDir()
	_, err := extractZip(invalidZip, destDir, "anybin")
	if err == nil {
		t.Error("Expected error for invalid zip file, got none")
	}
}

func TestExtractZip_EmptyDestDir(t *testing.T) {
	files := map[string]string{
		"testbin": "content",
	}
	zipPath := createTestZip(t, files)
	
	_, err := extractZip(zipPath, "", "testbin")
	if err == nil {
		t.Error("Expected error for empty destDir, got none")
	}
}

func TestExtractZip_EmptyBinaryName(t *testing.T) {
	files := map[string]string{
		"testbin": "content",
	}
	zipPath := createTestZip(t, files)
	destDir := t.TempDir()
	
	_, err := extractZip(zipPath, destDir, "")
	if err == nil {
		t.Error("Expected error for empty binary name, got none")
	}
}

func TestExtractZip_NonExistentFile(t *testing.T) {
	destDir := t.TempDir()
	_, err := extractZip("/nonexistent/file.zip", destDir, "testbin")
	if err == nil {
		t.Error("Expected error for nonexistent zip file, got none")
	}
}

// TestExtractAsset tests the main extraction function
func TestExtractAsset_TarGz(t *testing.T) {
	files := map[string]string{
		"mybin": "binary content",
	}
	tarPath := createTestTarGz(t, files)
	
	binary := &database.Binary{
		UserID: "testbin",
		Name:   "mybin",
		Format: ".tar.gz",
	}
	
	// Note: This will fail because getExtractPath tries to use real paths
	// We're testing the dispatch logic here
	_, err := ExtractAsset(tarPath, binary, "v1.0.0")
	// We expect an error about extract path, but not about format
	if err != nil && err.Error() == "unsupported asset format: .tar.gz" {
		t.Error("Should not report unsupported format for .tar.gz")
	}
}

func TestExtractAsset_Zip(t *testing.T) {
	files := map[string]string{
		"mybin": "binary content",
	}
	zipPath := createTestZip(t, files)
	
	binary := &database.Binary{
		UserID: "testbin",
		Name:   "mybin",
		Format: ".zip",
	}
	
	// Note: This will fail because getExtractPath tries to use real paths
	// We're testing the dispatch logic here
	_, err := ExtractAsset(zipPath, binary, "v1.0.0")
	// We expect an error about extract path, but not about format
	if err != nil && err.Error() == "unsupported asset format: .zip" {
		t.Error("Should not report unsupported format for .zip")
	}
}

func TestExtractAsset_UnsupportedFormat(t *testing.T) {
	binary := &database.Binary{
		UserID: "testbin",
		Name:   "mybin",
		Format: ".rar",
	}
	
	_, err := ExtractAsset("/some/path", binary, "v1.0.0")
	if err == nil {
		t.Error("Expected error for unsupported format, got none")
	}
	if err != nil && err.Error() != "unsupported asset format: .rar" {
		t.Errorf("Expected unsupported format error, got: %v", err)
	}
}

// Test extraction with multiple files to ensure we find the right one
func TestExtractTar_MultipleFiles(t *testing.T) {
	files := map[string]string{
		"README.md":     "readme content",
		"LICENSE":       "license content",
		"bin/testbin":   "this is the binary",
		"bin/other":     "another file",
		"doc/guide.txt": "documentation",
	}
	tarPath := createTestTarGz(t, files)
	destDir := t.TempDir()
	
	extractedPath, err := extractTar(tarPath, destDir, "testbin")
	if err != nil {
		t.Fatalf("extractTar failed: %v", err)
	}
	
	content, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	if string(content) != "this is the binary" {
		t.Errorf("Extracted wrong file, got: %s", string(content))
	}
}

func TestExtractZip_MultipleFiles(t *testing.T) {
	files := map[string]string{
		"README.md":     "readme content",
		"LICENSE":       "license content",
		"bin/testbin":   "this is the binary",
		"bin/other":     "another file",
		"doc/guide.txt": "documentation",
	}
	zipPath := createTestZip(t, files)
	destDir := t.TempDir()
	
	extractedPath, err := extractZip(zipPath, destDir, "testbin")
	if err != nil {
		t.Fatalf("extractZip failed: %v", err)
	}
	
	content, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	if string(content) != "this is the binary" {
		t.Errorf("Extracted wrong file, got: %s", string(content))
	}
}

// Test large binary extraction
func TestExtractTar_LargeBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large binary test in short mode")
	}
	
	// Create a "large" binary (1MB)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	
	files := map[string]string{
		"largebin": string(largeContent),
	}
	tarPath := createTestTarGz(t, files)
	destDir := t.TempDir()
	
	extractedPath, err := extractTar(tarPath, destDir, "largebin")
	if err != nil {
		t.Fatalf("extractTar failed for large binary: %v", err)
	}
	
	// Verify size
	info, err := os.Stat(extractedPath)
	if err != nil {
		t.Fatalf("Failed to stat extracted file: %v", err)
	}
	if info.Size() != int64(len(largeContent)) {
		t.Errorf("Size mismatch: expected %d, got %d", len(largeContent), info.Size())
	}
}

// Benchmark extraction performance
func BenchmarkExtractTar(b *testing.B) {
	files := map[string]string{
		"bin/testbin": "binary content here",
	}
	
	// Create tar once
	tmpFile, err := os.CreateTemp("", "bench-*.tar.gz")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	tarPath := tmpFile.Name()
	defer os.Remove(tarPath)
	tmpFile.Close()
	
	// Write tar
	f, _ := os.Create(tarPath)
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	
	for name, content := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0755,
			Size: int64(len(content)),
			Typeflag: tar.TypeReg,
		}
		tw.WriteHeader(header)
		tw.Write([]byte(content))
	}
	tw.Close()
	gw.Close()
	f.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destDir := b.TempDir()
		if _, err := extractTar(tarPath, destDir, "testbin"); err != nil {
			b.Fatalf("extractTar failed: %v", err)
		}
	}
}

func BenchmarkExtractZip(b *testing.B) {
	files := map[string]string{
		"bin/testbin": "binary content here",
	}
	
	// Create zip once
	tmpFile, err := os.CreateTemp("", "bench-*.zip")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	zipPath := tmpFile.Name()
	defer os.Remove(zipPath)
	tmpFile.Close()
	
	// Write zip
	f, _ := os.Create(zipPath)
	zw := zip.NewWriter(f)
	
	for name, content := range files {
		fw, _ := zw.Create(name)
		fw.Write([]byte(content))
	}
	zw.Close()
	f.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destDir := b.TempDir()
		if _, err := extractZip(zipPath, destDir, "testbin"); err != nil {
			b.Fatalf("extractZip failed: %v", err)
		}
	}
}
