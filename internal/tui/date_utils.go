package tui

import (
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/core/format"
)

// getDefaultDateFormat returns the default date format
// This is a wrapper around the shared format package function
func getDefaultDateFormat() string {
	return format.GetDefaultDateFormat()
}

// getDateFormat returns the configured date format or a sensible default.
func getDateFormat(cfg *config.Config) string {
	dateFormat := format.GetDefaultDateFormat()
	if cfg != nil && cfg.DateFormat != "" {
		dateFormat = cfg.DateFormat
	}
	return dateFormat
}
