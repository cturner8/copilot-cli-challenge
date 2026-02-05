package binary

import (
	"fmt"
	"log"

	"cturner8/binmate/internal/core/crypto"
	"cturner8/binmate/internal/core/url"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

// AddBinaryFromURL adds a binary by parsing a GitHub release URL
func AddBinaryFromURL(rawURL string, dbService *repository.Service) (*database.Binary, error) {
	// Parse the GitHub release URL
	parsed, err := url.ParseGitHubReleaseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Generate binary ID and name
	binaryID := url.GenerateBinaryID(parsed.AssetName)
	binaryName := url.GenerateBinaryName(parsed.AssetName)

	// Create binary definition
	binary := &database.Binary{
		UserID:        binaryID,
		Name:          binaryName,
		Provider:      "github",
		ProviderPath:  fmt.Sprintf("%s/%s", parsed.Owner, parsed.Repo),
		Format:        parsed.Format,
		ConfigVersion: 0, // TUI-added binaries have ConfigVersion=0
	}

	// Compute config digest
	binary.ConfigDigest = crypto.ComputeDigest(
		binary.UserID, binary.Name, "", binary.Provider,
		binary.ProviderPath, "", binary.Format, "", "",
	)

	// Check if binary already exists
	existing, err := dbService.Binaries.GetByUserID(binaryID)
	if err == nil {
		// Binary exists, return it
		log.Printf("Binary %s already exists", binaryID)
		return existing, nil
	} else if err != database.ErrNotFound {
		return nil, fmt.Errorf("failed to check existing binary: %w", err)
	}

	// Create the binary
	if err := dbService.Binaries.Create(binary); err != nil {
		return nil, fmt.Errorf("failed to create binary: %w", err)
	}

	log.Printf("Binary %s added successfully", binaryID)
	return binary, nil
}

// RemoveBinary removes a binary and all its installations
func RemoveBinary(binaryID string, dbService *repository.Service, removeFiles bool) error {
	// Get the binary
	binary, err := dbService.Binaries.GetByUserID(binaryID)
	if err != nil {
		return fmt.Errorf("binary not found: %w", err)
	}

	// Get all installations
	installations, err := dbService.Installations.ListByBinary(binary.ID)
	if err != nil {
		return fmt.Errorf("failed to list installations: %w", err)
	}

	// TODO: If removeFiles is true, delete the physical files and symlinks
	// This would require implementing file system operations
	if removeFiles {
		log.Printf("Warning: physical file removal not yet implemented")
	}

	// Delete from database (cascade will handle installations, versions)
	if err := dbService.Binaries.Delete(binary.ID); err != nil {
		return fmt.Errorf("failed to delete binary: %w", err)
	}

	log.Printf("Binary %s removed successfully (%d installations)", binaryID, len(installations))
	return nil
}

// ListBinariesWithDetails retrieves all binaries with version information
func ListBinariesWithDetails(dbService *repository.Service) ([]*repository.BinaryWithVersionDetails, error) {
	return dbService.Binaries.ListWithVersionDetails("No active version")
}

// GetBinaryByID retrieves a binary by its user ID
func GetBinaryByID(binaryID string, dbService *repository.Service) (*database.Binary, error) {
	binary, err := dbService.Binaries.GetByUserID(binaryID)
	if err != nil {
		return nil, fmt.Errorf("binary not found: %w", err)
	}
	return binary, nil
}

// ImportBinary imports an existing binary from the file system
// This is a placeholder for future implementation
func ImportBinary(path string, name string, dbService *repository.Service) (*database.Binary, error) {
	// TODO: Implement binary import functionality
	// This would involve:
	// 1. Verify the file exists and is executable
	// 2. Compute checksum
	// 3. Determine version (if possible)
	// 4. Create binary and installation records
	// 5. Create symlink

	return nil, fmt.Errorf("import functionality not yet implemented")
}
