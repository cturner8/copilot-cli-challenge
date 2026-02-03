package tui

const (
	// Text truncation constants for table cells
	// These represent the overhead for padding and ellipsis
	cellPaddingOverhead = 2 // Minimum padding on each side
	ellipsisLength      = 3 // Length of "..." truncation indicator
	minTruncateWidth    = cellPaddingOverhead + ellipsisLength // Minimum width before truncation
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
