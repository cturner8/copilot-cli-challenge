package tui

import (
	"cturner8/binmate/internal/tui/views"

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
	view  views.ViewState
	label string
}

// availableTabs returns the list of tabs available in the main views
var availableTabs = []tabDefinition{
	{views.BinariesList, "📦 Binaries"},
	{views.Downloads, "📥 Downloads"},
	{views.Configuration, "⚙️  Config"},
	{views.Help, "❓ Help"},
}

// renderTabs renders the tab bar
func (m Model) renderTabs() string {
	// Don't show tabs in add binary or versions views
	if m.CurrentView == views.AddBinaryURL || m.CurrentView == views.AddBinaryForm || m.CurrentView == views.Versions {
		return ""
	}

	var tabs []string
	for _, tab := range availableTabs {
		var style lipgloss.Style
		if m.CurrentView == tab.view {
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
func getTabForKey(key string) (views.ViewState, bool) {
	switch key {
	case "1":
		return views.BinariesList, true
	case "2":
		return views.Downloads, true
	case "3":
		return views.Configuration, true
	case "4":
		return views.Help, true
	default:
		return views.BinariesList, false
	}
}

// getNextTab returns the next tab in the sequence
func getNextTab(current views.ViewState) views.ViewState {
	for i, tab := range availableTabs {
		if tab.view == current {
			// Return next tab, wrapping around to the beginning
			nextIndex := (i + 1) % len(availableTabs)
			return availableTabs[nextIndex].view
		}
	}
	// Default to first tab if current not found
	return availableTabs[0].view
}

// getPreviousTab returns the previous tab in the sequence
func getPreviousTab(current views.ViewState) views.ViewState {
	for i, tab := range availableTabs {
		if tab.view == current {
			// Return previous tab, wrapping around to the end
			prevIndex := (i - 1 + len(availableTabs)) % len(availableTabs)
			return availableTabs[prevIndex].view
		}
	}
	// Default to last tab if current not found
	return availableTabs[len(availableTabs)-1].view
}

// handleTabCycling handles shift+tab and ctrl+shift+tab for cycling between tabs
// Returns true if a tab cycle key was pressed, along with the updated model
func handleTabCycling(m Model, key string) (Model, bool) {
	switch key {
	case keyTab:
		m.CurrentView = getNextTab(m.CurrentView)
		return m, true
	case keyShiftTab:
		m.CurrentView = getPreviousTab(m.CurrentView)
		return m, true
	}
	return m, false
}
