package format

import (
	"os"
	"strings"
	"time"
)

// FormatTimestamp formats a Unix timestamp to a human-readable date string
// using the provided format or a default format based on locale/timezone
func FormatTimestamp(timestamp int64, dateFormat string) string {
	t := time.Unix(timestamp, 0)

	if dateFormat == "" {
		dateFormat = GetDefaultDateFormat()
	}

	return t.Format(dateFormat)
}

// GetDefaultDateFormat attempts to determine the default date format
// based on the system's locale and timezone settings
func GetDefaultDateFormat() string {
	// Try to get locale information from environment
	locale := getLocaleFromEnv()

	// Determine format based on locale
	if locale != "" {
		// European/UK formats typically use DD/MM/YYYY
		if isEuropeanLocale(locale) {
			return "02/01/2006 15:04"
		}
		// US format uses MM/DD/YYYY
		if isUSLocale(locale) {
			return "01/02/2006 15:04"
		}
	}

	// Check timezone as a fallback hint
	tz := getTimezone()
	if tz != "" {
		// European timezones suggest European date format
		if isEuropeanTimezone(tz) {
			return "02/01/2006 15:04"
		}
		// US timezones suggest US date format
		if isUSTimezone(tz) {
			return "01/02/2006 15:04"
		}
	}

	// Default to ISO format (most unambiguous)
	return "2006-01-02 15:04"
}

// getLocaleFromEnv extracts locale information from environment variables
func getLocaleFromEnv() string {
	// Check common locale environment variables in order of preference
	for _, envVar := range []string{"LC_TIME", "LC_ALL", "LANG"} {
		if val := os.Getenv(envVar); val != "" {
			return val
		}
	}
	return ""
}

// getTimezone gets the system timezone
func getTimezone() string {
	// Get local timezone name
	t := time.Now()
	zone, _ := t.Zone()
	return zone
}

// isEuropeanLocale checks if the locale string indicates a European locale
func isEuropeanLocale(locale string) bool {
	locale = strings.ToLower(locale)
	// Common European locale prefixes
	europeanCountries := []string{
		"en_gb", "en_ie", "en_au", "en_nz", // English-speaking Commonwealth countries
		"de_", "fr_", "es_", "it_", "pt_", "nl_", "pl_", "ru_", // Major European countries
		"sv_", "no_", "da_", "fi_", "cs_", "hu_", "ro_", "el_", // Nordic and Eastern European
	}

	for _, country := range europeanCountries {
		if strings.HasPrefix(locale, country) {
			return true
		}
	}
	return false
}

// isUSLocale checks if the locale string indicates a US locale
func isUSLocale(locale string) bool {
	locale = strings.ToLower(locale)
	return strings.HasPrefix(locale, "en_us")
}

// isEuropeanTimezone checks if the timezone suggests European location
func isEuropeanTimezone(tz string) bool {
	europeanTZ := []string{
		"GMT", "BST", "CET", "CEST", "EET", "EEST", "WET", "WEST",
		"MSK", // Common European timezones
	}

	for _, europeanZone := range europeanTZ {
		if strings.Contains(tz, europeanZone) {
			return true
		}
	}
	return false
}

// isUSTimezone checks if the timezone suggests US location
// Note: These abbreviations can be ambiguous (e.g., CST could be China/Cuba/Central)
// This is a best-effort heuristic for common cases
func isUSTimezone(tz string) bool {
	usTZ := []string{
		"EST", "EDT", "CST", "CDT", "MST", "MDT", "PST", "PDT",
		"AKST", "AKDT", "HST", // US timezones
	}

	for _, usZone := range usTZ {
		if strings.Contains(tz, usZone) {
			return true
		}
	}
	return false
}
