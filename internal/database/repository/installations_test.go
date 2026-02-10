package repository

import (
	"path/filepath"
	"testing"

	"cturner8/binmate/internal/database"
)

func TestVerifyChecksum(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	svc := NewService(db)

	// Create a test binary
	binary := &database.Binary{
		UserID:       "test-id",
		Name:         "TestBin",
		Provider:     "github",
		ProviderPath: "owner/repo",
		Format:       ".tar.gz",
		Source:       "manual",
	}
	err = svc.Binaries.Create(binary)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	// Create a test installation with a checksum
	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "v1.0.0",
		InstalledPath:     "/test/path",
		SourceURL:         "http://example.com/test.tar.gz",
		FileSize:          1024,
		Checksum:          "abc123",
		ChecksumAlgorithm: "sha256",
	}
	err = svc.Installations.Create(installation)
	if err != nil {
		t.Fatalf("Failed to create installation: %v", err)
	}

	tests := []struct {
		name             string
		installationID   int64
		expectedChecksum string
		wantMatch        bool
		wantError        error
	}{
		{
			name:             "checksum matches",
			installationID:   installation.ID,
			expectedChecksum: "abc123",
			wantMatch:        true,
			wantError:        nil,
		},
		{
			name:             "checksum does not match",
			installationID:   installation.ID,
			expectedChecksum: "wrong123",
			wantMatch:        false,
			wantError:        nil,
		},
		{
			name:             "installation not found",
			installationID:   99999,
			expectedChecksum: "abc123",
			wantMatch:        false,
			wantError:        database.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := svc.Installations.VerifyChecksum(tt.installationID, tt.expectedChecksum)

			if err != tt.wantError {
				t.Errorf("VerifyChecksum() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if match != tt.wantMatch {
				t.Errorf("VerifyChecksum() match = %v, wantMatch %v", match, tt.wantMatch)
			}
		})
	}
}
