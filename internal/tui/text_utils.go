package tui

import (
	"strconv"
	"strings"
)

const (
	// Text truncation constants for table cells
	// These represent the overhead for padding and ellipsis
	cellPaddingOverhead = 2                                    // Minimum padding on each side
	ellipsisLength      = 3                                    // Length of "..." truncation indicator
	minTruncateWidth    = cellPaddingOverhead + ellipsisLength // Minimum width before truncation

	// Layout constants
	defaultTerminalWidth = 80 // Default terminal width when not detected
	columnPadding4       = 8  // Padding for 4 columns (2 chars each)
	columnPadding5       = 10 // Padding for 5 columns (2 chars each)
)

// truncateText truncates text to fit within the specified width
// If the text is longer than width, it truncates and adds "..."
func truncateText(text string, width int) string {
	if len(text) <= width-cellPaddingOverhead {
		return text
	}

	if width < minTruncateWidth {
		// Not enough space for ellipsis, just truncate
		if width > 0 {
			return text[:width]
		}
		return ""
	}

	// Truncate with ellipsis
	maxLen := width - minTruncateWidth
	if maxLen > 0 && maxLen < len(text) {
		return text[:maxLen] + "..."
	}

	return text
}

// truncatePathEnd truncates a path from the beginning, keeping the end
// This is useful for file paths where the end is usually more relevant
func truncatePathEnd(path string, width int) string {
	if len(path) <= width-cellPaddingOverhead {
		return path
	}

	if width < minTruncateWidth {
		if width > 0 {
			return path[len(path)-width:]
		}
		return ""
	}

	// Keep the end of the path
	keepLen := width - minTruncateWidth
	if keepLen > 0 && keepLen < len(path) {
		return "..." + path[len(path)-keepLen:]
	}

	return path
}

// formatIntWithCommas formats an integer with thousands separators.
func formatIntWithCommas(value int) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	digits := strconv.Itoa(value)
	if len(digits) <= 3 {
		return sign + digits
	}

	var b strings.Builder
	b.WriteString(sign)

	prefixLen := len(digits) % 3
	if prefixLen > 0 {
		b.WriteString(digits[:prefixLen])
		if len(digits) > prefixLen {
			b.WriteByte(',')
		}
	}

	for i := prefixLen; i < len(digits); i += 3 {
		b.WriteString(digits[i : i+3])
		if i+3 < len(digits) {
			b.WriteByte(',')
		}
	}

	return b.String()
}
