package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color scheme
	primaryColor   = lipgloss.Color("#7C3AED")  // Purple
	secondaryColor = lipgloss.Color("#A78BFA")  // Light purple
	accentColor    = lipgloss.Color("#10B981")  // Green
	errorColor     = lipgloss.Color("#EF4444")  // Red
	successColor   = lipgloss.Color("#10B981")  // Green
	mutedColor     = lipgloss.Color("#6B7280")  // Grey

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Underline(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Padding(0, 1)

	tableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Border styles
	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Help/status bar style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(1, 0)

	// Empty state style
	emptyStateStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Padding(2, 4)

	// Loading style
	loadingStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	// Active indicator style
	activeIndicatorStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	// Form styles
	formLabelStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	formInputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	formInputFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)
)

// Helper functions for consistent padding and spacing
func padLeft(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Left).Render(s)
}

func padRight(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(s)
}

func center(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(s)
}
