package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Tab styles
var (
	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 2).
			Bold(true)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Background(lipgloss.Color("#1F2937")).
				Padding(0, 2)

	tabGapStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)

// tabDefinition represents a tab in the UI
type tabDefinition struct {
	view  viewState
	label string
	key   string
}

// availableTabs returns the list of tabs available in the main views
var availableTabs = []tabDefinition{
	{viewBinariesList, "📦 Binaries", ""},
	{viewDownloads, "📥 Downloads", ""},
	{viewConfiguration, "⚙️  Config", ""},
	{viewHelp, "❓ Help", ""},
}

// renderTabs renders the tab bar
func (m model) renderTabs() string {
	// Don't show tabs in add binary or versions views
	if m.currentView == viewAddBinaryURL || m.currentView == viewAddBinaryForm || m.currentView == viewVersions {
		return ""
	}

	var tabs []string
	for _, tab := range availableTabs {
		var style lipgloss.Style
		if m.currentView == tab.view {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		tabs = append(tabs, style.Render(tab.label))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return tabBar + "\n\n"
}

// getTabForKey returns the view state for a given key press
func getTabForKey(key string) (viewState, bool) {
	switch key {
	case "1":
		return viewBinariesList, true
	case "2":
		return viewDownloads, true
	case "3":
		return viewConfiguration, true
	case "4":
		return viewHelp, true
	default:
		return viewBinariesList, false
	}
}
