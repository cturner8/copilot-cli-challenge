package tui

import (
	"fmt"
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

// renderConfiguration renders the configuration view
func (m model) renderConfiguration() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())

	// Show error if any
	if m.errorMessage != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMessage))
		b.WriteString("\n\n")
	}

	// Show success message if any
	if m.successMessage != "" {
		b.WriteString(successStyle.Render("✓ " + m.successMessage))
		b.WriteString("\n\n")
	}

	// Show loading state
	if m.loading {
		b.WriteString(loadingStyle.Render("Syncing configuration..."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Configuration display
	b.WriteString(headerStyle.Render("Configuration Settings"))
	b.WriteString("\n\n")

	if m.config != nil {
		b.WriteString(fmt.Sprintf("Version: %d\n", m.config.Version))
		b.WriteString(fmt.Sprintf("Binaries in config: %d\n", len(m.config.Binaries)))
		if m.config.DateFormat != "" {
			b.WriteString(fmt.Sprintf("Date Format: %s\n", m.config.DateFormat))
		}
		if m.config.LogLevel != "" {
			b.WriteString(fmt.Sprintf("Log Level: %s\n", m.config.LogLevel))
		}
		b.WriteString("\n")

		// Show first few binaries from config
		if len(m.config.Binaries) > 0 {
			b.WriteString(headerStyle.Render("Configured Binaries:"))
			b.WriteString("\n")
			maxShow := 5
			if len(m.config.Binaries) < maxShow {
				maxShow = len(m.config.Binaries)
			}
			for i := 0; i < maxShow; i++ {
				binary := m.config.Binaries[i]
				b.WriteString(fmt.Sprintf("  • %s (%s)\n", binary.Name, binary.Id))
			}
			if len(m.config.Binaries) > maxShow {
				b.WriteString(fmt.Sprintf("  ... and %d more\n", len(m.config.Binaries)-maxShow))
			}
		}
	} else {
		b.WriteString(emptyStateStyle.Render("No configuration loaded"))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("s: sync config to database • esc: back • q: quit"))

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
