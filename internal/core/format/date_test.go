package format

import (
	"os"
	"testing"
	"time"
)

func TestFormatTimestamp(t *testing.T) {
	// Use a fixed timestamp for testing
	timestamp := int64(1609459200) // 2021-01-01 00:00:00 UTC
	
	tests := []struct {
		name       string
		timestamp  int64
		dateFormat string
		want       string
	}{
		{
			name:       "ISO format",
			timestamp:  timestamp,
			dateFormat: "2006-01-02 15:04",
			want:       "2021-01-01 00:00",
		},
		{
			name:       "US format",
			timestamp:  timestamp,
			dateFormat: "01/02/2006 15:04",
			want:       "01/01/2021 00:00",
		},
		{
			name:       "European format",
			timestamp:  timestamp,
			dateFormat: "02/01/2006 15:04",
			want:       "01/01/2021 00:00",
		},
		{
			name:       "empty format uses default",
			timestamp:  timestamp,
			dateFormat: "",
			// We can't predict the exact output as it depends on locale
			// Just verify it doesn't panic
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimestamp(tt.timestamp, tt.dateFormat)
			if tt.want != "" && got != tt.want {
				t.Errorf("FormatTimestamp() = %v, want %v", got, tt.want)
			}
			// At minimum, verify we got some output
			if got == "" {
				t.Error("FormatTimestamp() returned empty string")
			}
		})
	}
}

func TestGetDefaultDateFormat(t *testing.T) {
	// Test that it returns a valid format
	format := GetDefaultDateFormat()
	if format == "" {
		t.Error("GetDefaultDateFormat() returned empty string")
	}
	
	// Verify the format works with time.Format
	now := time.Now()
	result := now.Format(format)
	if result == "" {
		t.Error("GetDefaultDateFormat() returned invalid format")
	}
}

func TestIsEuropeanLocale(t *testing.T) {
	tests := []struct {
		name   string
		locale string
		want   bool
	}{
		{"UK locale", "en_GB.UTF-8", true},
		{"German locale", "de_DE.UTF-8", true},
		{"French locale", "fr_FR.UTF-8", true},
		{"US locale", "en_US.UTF-8", false},
		{"empty locale", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEuropeanLocale(tt.locale); got != tt.want {
				t.Errorf("isEuropeanLocale(%q) = %v, want %v", tt.locale, got, tt.want)
			}
		})
	}
}

func TestIsUSLocale(t *testing.T) {
	tests := []struct {
		name   string
		locale string
		want   bool
	}{
		{"US locale", "en_US.UTF-8", true},
		{"UK locale", "en_GB.UTF-8", false},
		{"empty locale", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUSLocale(tt.locale); got != tt.want {
				t.Errorf("isUSLocale(%q) = %v, want %v", tt.locale, got, tt.want)
			}
		})
	}
}

func TestIsEuropeanTimezone(t *testing.T) {
	tests := []struct {
		name string
		tz   string
		want bool
	}{
		{"GMT", "GMT", true},
		{"BST", "BST", true},
		{"CET", "CET", true},
		{"EST", "EST", false},
		{"empty", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEuropeanTimezone(tt.tz); got != tt.want {
				t.Errorf("isEuropeanTimezone(%q) = %v, want %v", tt.tz, got, tt.want)
			}
		})
	}
}

func TestIsUSTimezone(t *testing.T) {
	tests := []struct {
		name string
		tz   string
		want bool
	}{
		{"EST", "EST", true},
		{"PST", "PST", true},
		{"GMT", "GMT", false},
		{"empty", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUSTimezone(tt.tz); got != tt.want {
				t.Errorf("isUSTimezone(%q) = %v, want %v", tt.tz, got, tt.want)
			}
		})
	}
}

func TestGetLocaleFromEnv(t *testing.T) {
	// Save original env vars
	origLCTime := os.Getenv("LC_TIME")
	origLCAll := os.Getenv("LC_ALL")
	origLang := os.Getenv("LANG")
	
	defer func() {
		// Restore original env vars
		os.Setenv("LC_TIME", origLCTime)
		os.Setenv("LC_ALL", origLCAll)
		os.Setenv("LANG", origLang)
	}()
	
	// Clear all locale env vars
	os.Unsetenv("LC_TIME")
	os.Unsetenv("LC_ALL")
	os.Unsetenv("LANG")
	
	// Test empty case
	if got := getLocaleFromEnv(); got != "" {
		t.Errorf("getLocaleFromEnv() with no env vars = %q, want empty string", got)
	}
	
	// Test LANG
	os.Setenv("LANG", "en_US.UTF-8")
	if got := getLocaleFromEnv(); got != "en_US.UTF-8" {
		t.Errorf("getLocaleFromEnv() with LANG = %q, want %q", got, "en_US.UTF-8")
	}
	
	// Test LC_ALL takes precedence
	os.Setenv("LC_ALL", "en_GB.UTF-8")
	if got := getLocaleFromEnv(); got != "en_GB.UTF-8" {
		t.Errorf("getLocaleFromEnv() with LC_ALL = %q, want %q", got, "en_GB.UTF-8")
	}
	
	// Test LC_TIME takes precedence
	os.Setenv("LC_TIME", "de_DE.UTF-8")
	if got := getLocaleFromEnv(); got != "de_DE.UTF-8" {
		t.Errorf("getLocaleFromEnv() with LC_TIME = %q, want %q", got, "de_DE.UTF-8")
	}
}

func TestGetTimezone(t *testing.T) {
	tz := getTimezone()
	// We can't predict the exact timezone, just verify we get something
	if tz == "" {
		t.Error("getTimezone() returned empty string")
	}
}
