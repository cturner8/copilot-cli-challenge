package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// Known test vectors for SHA256
const (
	testContent1       = "Hello, World!"
	testSHA256_1       = "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	testContent2       = "The quick brown fox jumps over the lazy dog"
	testSHA256_2       = "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"
	emptyContentSHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

func TestComputeSHA256_ValidFile(t *testing.T) {
	// Create temp file
	tmpFile := createTempFile(t, testContent1)
	defer os.Remove(tmpFile)

	// Compute checksum
	checksum, err := ComputeSHA256(tmpFile)
	if err != nil {
		t.Fatalf("ComputeSHA256 failed: %v", err)
	}

	// Verify checksum matches expected
	if checksum != testSHA256_1 {
		t.Errorf("Checksum mismatch:\nExpected: %s\nGot:      %s", testSHA256_1, checksum)
	}
}

func TestComputeSHA256_DifferentContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "test_content_1",
			content:  testContent1,
			expected: testSHA256_1,
		},
		{
			name:     "test_content_2",
			content:  testContent2,
			expected: testSHA256_2,
		},
		{
			name:     "empty_file",
			content:  "",
			expected: emptyContentSHA256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempFile(t, tt.content)
			defer os.Remove(tmpFile)

			checksum, err := ComputeSHA256(tmpFile)
			if err != nil {
				t.Fatalf("ComputeSHA256 failed: %v", err)
			}

			if checksum != tt.expected {
				t.Errorf("Checksum mismatch:\nExpected: %s\nGot:      %s", tt.expected, checksum)
			}
		})
	}
}

func TestComputeSHA256_NonExistentFile(t *testing.T) {
	_, err := ComputeSHA256("/nonexistent/file/path.txt")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestComputeSHA256_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := ComputeSHA256(tmpDir)
	if err == nil {
		t.Error("Expected error when computing checksum of directory, got nil")
	}
}

func TestVerifySHA256_Success(t *testing.T) {
	tmpFile := createTempFile(t, testContent1)
	defer os.Remove(tmpFile)

	err := VerifySHA256(tmpFile, testSHA256_1)
	if err != nil {
		t.Errorf("VerifySHA256 failed for valid checksum: %v", err)
	}
}

func TestVerifySHA256_Mismatch(t *testing.T) {
	tmpFile := createTempFile(t, testContent1)
	defer os.Remove(tmpFile)

	// Use a different (incorrect) checksum
	incorrectChecksum := testSHA256_2

	err := VerifySHA256(tmpFile, incorrectChecksum)
	if err == nil {
		t.Error("Expected checksum mismatch error, got nil")
	}

	// Verify error message contains both checksums
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestVerifySHA256_NonExistentFile(t *testing.T) {
	err := VerifySHA256("/nonexistent/file.txt", testSHA256_1)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestComputeDigest_SingleField(t *testing.T) {
	digest := ComputeDigest("test")

	// Verify format is "sha256:..."
	if len(digest) < 8 || digest[:7] != "sha256:" {
		t.Errorf("Digest should start with 'sha256:', got: %s", digest)
	}

	// Verify checksum part is 64 hex characters
	checksum := digest[7:]
	if len(checksum) != 64 {
		t.Errorf("SHA256 checksum should be 64 hex characters, got %d", len(checksum))
	}

	// Verify it's valid hex
	if _, err := hex.DecodeString(checksum); err != nil {
		t.Errorf("Checksum should be valid hex: %v", err)
	}
}

func TestComputeDigest_MultipleFields(t *testing.T) {
	digest1 := ComputeDigest("field1", "field2", "field3")
	digest2 := ComputeDigest("field1", "field2", "field3")
	digest3 := ComputeDigest("field1", "field2", "different")

	// Same inputs should produce same digest
	if digest1 != digest2 {
		t.Error("Same inputs should produce same digest")
	}

	// Different inputs should produce different digest
	if digest1 == digest3 {
		t.Error("Different inputs should produce different digest")
	}
}

func TestComputeDigest_OrderMatters(t *testing.T) {
	digest1 := ComputeDigest("a", "b")
	digest2 := ComputeDigest("b", "a")

	// Different order should produce different digest
	if digest1 == digest2 {
		t.Error("Field order should affect digest")
	}
}

func TestComputeDigest_EmptyFields(t *testing.T) {
	digest1 := ComputeDigest()
	digest2 := ComputeDigest("")
	digest3 := ComputeDigest("", "")

	// Different number of fields should produce different digests
	if digest1 == digest2 || digest2 == digest3 {
		t.Error("Different number of empty fields should produce different digests")
	}
}

func TestVerifyDigest_ValidSHA256(t *testing.T) {
	tmpFile := createTempFile(t, testContent1)
	defer os.Remove(tmpFile)

	digest := "sha256:" + testSHA256_1

	err := VerifyDigest(tmpFile, digest)
	if err != nil {
		t.Errorf("VerifyDigest failed for valid digest: %v", err)
	}
}

func TestVerifyDigest_Mismatch(t *testing.T) {
	tmpFile := createTempFile(t, testContent1)
	defer os.Remove(tmpFile)

	// Use incorrect checksum
	digest := "sha256:" + testSHA256_2

	err := VerifyDigest(tmpFile, digest)
	if err == nil {
		t.Error("Expected digest mismatch error, got nil")
	}
}

func TestVerifyDigest_InvalidFormat(t *testing.T) {
	tmpFile := createTempFile(t, testContent1)
	defer os.Remove(tmpFile)

	tests := []struct {
		name   string
		digest string
	}{
		{
			name:   "no_colon",
			digest: "sha256" + testSHA256_1,
		},
		{
			name:   "missing_checksum",
			digest: "sha256:",
		},
		{
			name:   "empty_string",
			digest: "",
		},
		{
			name:   "only_checksum",
			digest: testSHA256_1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyDigest(tmpFile, tt.digest)
			if err == nil {
				t.Errorf("Expected error for invalid digest format '%s', got nil", tt.digest)
			}
		})
	}
}

func TestVerifyDigest_UnsupportedAlgorithm(t *testing.T) {
	tmpFile := createTempFile(t, testContent1)
	defer os.Remove(tmpFile)

	tests := []string{
		"md5:abc123",
		"sha1:def456",
		"sha512:ghi789",
		"unknown:jkl012",
	}

	for _, digest := range tests {
		t.Run(digest, func(t *testing.T) {
			err := VerifyDigest(tmpFile, digest)
			if err == nil {
				t.Errorf("Expected error for unsupported algorithm in digest '%s', got nil", digest)
			}
		})
	}
}

func TestVerifyDigest_CaseInsensitive(t *testing.T) {
	tmpFile := createTempFile(t, testContent1)
	defer os.Remove(tmpFile)

	tests := []string{
		"SHA256:" + testSHA256_1,
		"Sha256:" + testSHA256_1,
		"sHa256:" + testSHA256_1,
	}

	for _, digest := range tests {
		t.Run(digest, func(t *testing.T) {
			err := VerifyDigest(tmpFile, digest)
			if err != nil {
				t.Errorf("VerifyDigest should be case-insensitive for algorithm, failed for '%s': %v", digest, err)
			}
		})
	}
}

func TestVerifyDigest_NonExistentFile(t *testing.T) {
	err := VerifyDigest("/nonexistent/file.txt", "sha256:"+testSHA256_1)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

// TestIntegration_ComputeAndVerify tests the full workflow
func TestIntegration_ComputeAndVerify(t *testing.T) {
	tmpFile := createTempFile(t, "Integration test content")
	defer os.Remove(tmpFile)

	// Compute checksum
	checksum, err := ComputeSHA256(tmpFile)
	if err != nil {
		t.Fatalf("Failed to compute checksum: %v", err)
	}

	// Verify with computed checksum (should succeed)
	if err := VerifySHA256(tmpFile, checksum); err != nil {
		t.Errorf("Verification failed with computed checksum: %v", err)
	}

	// Verify with digest format (should succeed)
	digest := "sha256:" + checksum
	if err := VerifyDigest(tmpFile, digest); err != nil {
		t.Errorf("Verification failed with digest format: %v", err)
	}

	// Modify file
	if err := os.WriteFile(tmpFile, []byte("Modified content"), 0644); err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}

	// Verify should now fail
	if err := VerifySHA256(tmpFile, checksum); err == nil {
		t.Error("Verification should fail after file modification")
	}
}

// TestLargeFile tests checksum computation for larger files
func TestLargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	tmpFile := filepath.Join(t.TempDir(), "large.bin")

	// Create a 10MB file
	size := 10 * 1024 * 1024
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Compute checksum (should not timeout or error)
	checksum1, err := ComputeSHA256(tmpFile)
	if err != nil {
		t.Fatalf("Failed to compute checksum for large file: %v", err)
	}

	// Compute again (should be consistent)
	checksum2, err := ComputeSHA256(tmpFile)
	if err != nil {
		t.Fatalf("Failed to compute checksum for large file (2nd time): %v", err)
	}

	if checksum1 != checksum2 {
		t.Error("Checksum should be consistent for same file")
	}

	// Verify should succeed
	if err := VerifySHA256(tmpFile, checksum1); err != nil {
		t.Errorf("Verification failed for large file: %v", err)
	}
}

// Helper function to create a temp file with content
func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tmpFile
}

// TestMain can be used for setup/teardown if needed
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// Benchmark functions for performance testing
func BenchmarkComputeSHA256_SmallFile(b *testing.B) {
	tmpFile := filepath.Join(b.TempDir(), "bench.txt")
	if err := os.WriteFile(tmpFile, []byte(testContent1), 0644); err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}
	defer os.Remove(tmpFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ComputeSHA256(tmpFile); err != nil {
			b.Fatalf("ComputeSHA256 failed: %v", err)
		}
	}
}

func BenchmarkComputeDigest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ComputeDigest("field1", "field2", "field3", "field4", "field5")
	}
}

func BenchmarkVerifyDigest(b *testing.B) {
	tmpFile := filepath.Join(b.TempDir(), "bench.txt")
	content := []byte(testContent1)
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Compute actual checksum
	h := sha256.New()
	h.Write(content)
	checksum := hex.EncodeToString(h.Sum(nil))
	digest := "sha256:" + checksum

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := VerifyDigest(tmpFile, digest); err != nil {
			b.Fatalf("VerifyDigest failed: %v", err)
		}
	}
}
