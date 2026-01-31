package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// ComputeSHA256 computes the SHA256 checksum of a file
func ComputeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to compute checksum: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifySHA256 verifies a file's SHA256 checksum matches the expected value
func VerifySHA256(filePath string, expectedChecksum string) error {
	actualChecksum, err := ComputeSHA256(filePath)
	if err != nil {
		return err
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// VerifyDigest parses a digest string and verifies a file's checksum
// Expected digest format: "algorithm:checksum"
// Example: "sha256:8bb862f8b61be63bb8b3f6b1dfb85bd556b7a8c174eb595e8db6d43e21c51afe"
func VerifyDigest(filePath string, digest string) error {
	parts := strings.SplitN(digest, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid digest format: expected 'algorithm:checksum', got '%s'", digest)
	}

	algorithm := strings.ToLower(parts[0])
	expectedChecksum := parts[1]

	switch algorithm {
	case "sha256":
		return VerifySHA256(filePath, expectedChecksum)
	default:
		return fmt.Errorf("unsupported algorithm: %s (only sha256 is supported)", algorithm)
	}
}
