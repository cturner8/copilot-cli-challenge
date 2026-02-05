package tui

import (
	"strings"
)

// renderDownloads renders the downloads placeholder view
func (m model) renderDownloads() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())

	b.WriteString(emptyStateStyle.Render("This view will allow you to manage cached asset downloads."))
	b.WriteString("\n")
	b.WriteString(emptyStateStyle.Render("(Not yet implemented)"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(getHelpText(m.currentView)))

	return b.String()
}

// renderConfiguration renders the configuration placeholder view
func (m model) renderConfiguration() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())

	b.WriteString(emptyStateStyle.Render("This view will allow you to manage global configuration/settings."))
	b.WriteString("\n")
	b.WriteString(emptyStateStyle.Render("(Not yet implemented)"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(getHelpText(m.currentView)))

	return b.String()
}

// renderHelp renders the help placeholder view
func (m model) renderHelp() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString(headerStyle.Render("Binmate - Binary Version Manager"))
	b.WriteString("\n\n")
	b.WriteString("Binmate helps you install and manage multiple versions of command-line binaries.\n")
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString("  ↑/↓ - Navigate through lists\n")
	b.WriteString("  Enter - View details or confirm\n")
	b.WriteString("  Esc - Go back\n")
	b.WriteString("  q - Quit application\n")
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("Actions"))
	b.WriteString("\n")
	b.WriteString("  a - Add a new binary\n")
	b.WriteString("  Ctrl+S - Save (in forms)\n")
	b.WriteString("\n")
	b.WriteString(emptyStateStyle.Render("(More help content to be added)"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(getHelpText(m.currentView)))

	return b.String()
}
