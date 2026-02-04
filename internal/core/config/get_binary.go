package config

import "fmt"

// GetBinary searches for a binary with the given id in the provided slice.
// Returns the binary if found, or an error if not found.
func GetBinary(id string, binaries []Binary) (Binary, error) {
	for _, binary := range binaries {
		if binary.Id == id {
			return binary, nil
		}
	}
	return Binary{}, fmt.Errorf("binary with id '%s' not found", id)
}
