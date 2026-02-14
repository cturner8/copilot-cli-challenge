package tui

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
		{
			name:     "bytes less than 1KB",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "exactly 1KB",
			bytes:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "kilobytes",
			bytes:    2048,
			expected: "2.0 KB",
		},
		{
			name:     "fractional kilobytes",
			bytes:    1536,
			expected: "1.5 KB",
		},
		{
			name:     "megabytes",
			bytes:    1048576,
			expected: "1.0 MB",
		},
		{
			name:     "fractional megabytes",
			bytes:    2621440,
			expected: "2.5 MB",
		},
		{
			name:     "gigabytes",
			bytes:    1073741824,
			expected: "1.0 GB",
		},
		{
			name:     "fractional gigabytes",
			bytes:    5368709120,
			expected: "5.0 GB",
		},
		{
			name:     "terabytes",
			bytes:    1099511627776,
			expected: "1.0 TB",
		},
		{
			name:     "large value",
			bytes:    9999999999999,
			expected: "9.1 TB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "text shorter than width",
			text:     "hello",
			width:    20,
			expected: "hello",
		},
		{
			name:     "text exactly at width threshold",
			text:     "hello world",
			width:    13, // 11 chars + 2 padding
			expected: "hello world",
		},
		{
			name:     "text longer than width",
			text:     "this is a very long text that needs truncation",
			width:    20,
			expected: "this is a very ...",
		},
		{
			name:     "width too small for ellipsis",
			width:    4,
			text:     "hello",
			expected: "hell",
		},
		{
			name:     "width of zero",
			width:    0,
			text:     "hello",
			expected: "",
		},
		{
			name:     "empty text",
			width:    20,
			text:     "",
			expected: "",
		},
		{
			name:     "minimum truncate width",
			width:    minTruncateWidth,
			text:     "hello world",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateText(tt.text, tt.width)
			if result != tt.expected {
				t.Errorf("truncateText(%q, %d) = %q, want %q", tt.text, tt.width, result, tt.expected)
			}
		})
	}
}

func TestTruncatePathEnd(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		width    int
		expected string
	}{
		{
			name:     "path shorter than width",
			path:     "/usr/bin",
			width:    20,
			expected: "/usr/bin",
		},
		{
			name:     "path exactly at width threshold",
			path:     "/usr/local/bin",
			width:    16, // 14 chars + 2 padding
			expected: "/usr/local/bin",
		},
		{
			name:     "path longer than width - keeps end",
			path:     "/usr/local/bin/very/long/path/to/binary",
			width:    20,
			expected: ".../path/to/binary",
		},
		{
			name:     "width too small for ellipsis",
			width:    4,
			path:     "/usr/bin",
			expected: "/bin",
		},
		{
			name:     "width of zero",
			width:    0,
			path:     "/usr/bin",
			expected: "",
		},
		{
			name:     "empty path",
			width:    20,
			path:     "",
			expected: "",
		},
		{
			name:     "minimum truncate width",
			width:    minTruncateWidth,
			path:     "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "home directory path",
			path:     "/home/user/.local/share/binmate/installations/binary",
			width:    30,
			expected: "...mate/installations/binary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncatePathEnd(tt.path, tt.width)
			if result != tt.expected {
				t.Errorf("truncatePathEnd(%q, %d) = %q, want %q", tt.path, tt.width, result, tt.expected)
			}
		})
	}
}

func TestPadLeft(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		minLen   int // minimum expected length
	}{
		{
			name:   "pad text to width",
			text:   "hello",
			width:  10,
			minLen: 10,
		},
		{
			name:   "text longer than width",
			text:   "hello world",
			width:  5,
			minLen: 5,
		},
		{
			name:   "empty text",
			text:   "",
			width:  10,
			minLen: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padLeft(tt.text, tt.width)
			// Note: lipgloss may add ANSI codes, so we check minimum length
			if len(result) < tt.minLen {
				t.Errorf("padLeft(%q, %d) length = %d, want at least %d", tt.text, tt.width, len(result), tt.minLen)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		minLen   int // minimum expected length
	}{
		{
			name:   "pad text to width",
			text:   "hello",
			width:  10,
			minLen: 10,
		},
		{
			name:   "text longer than width",
			text:   "hello world",
			width:  5,
			minLen: 5,
		},
		{
			name:   "empty text",
			text:   "",
			width:  10,
			minLen: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padRight(tt.text, tt.width)
			// Note: lipgloss may add ANSI codes, so we check minimum length
			if len(result) < tt.minLen {
				t.Errorf("padRight(%q, %d) length = %d, want at least %d", tt.text, tt.width, len(result), tt.minLen)
			}
		})
	}
}

func TestCenter(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		minLen   int // minimum expected length
	}{
		{
			name:   "center text in width",
			text:   "hello",
			width:  10,
			minLen: 10,
		},
		{
			name:   "text longer than width",
			text:   "hello world",
			width:  5,
			minLen: 5,
		},
		{
			name:   "empty text",
			text:   "",
			width:  10,
			minLen: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := center(tt.text, tt.width)
			// Note: lipgloss may add ANSI codes, so we check minimum length
			if len(result) < tt.minLen {
				t.Errorf("center(%q, %d) length = %d, want at least %d", tt.text, tt.width, len(result), tt.minLen)
			}
		})
	}
}
