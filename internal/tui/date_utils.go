package tui

import (
	"cturner8/binmate/internal/core/format"
)

// getDefaultDateFormat returns the default date format
// This is a wrapper around the shared format package function
func getDefaultDateFormat() string {
	return format.GetDefaultDateFormat()
}
