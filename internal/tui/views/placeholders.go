package views

import (
	"fmt"
	"strings"
)

// renderDownloads renders the downloads view
func (m Model) RenderDownloads() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.RenderTabsFn())

	// Show error if any
	if m.ErrorMessage != "" {
		b.WriteString(errorStyle.Render("Error: " + m.ErrorMessage))
		b.WriteString("\n\n")
	}

	// Show success message if any
	if m.SuccessMessage != "" {
		b.WriteString(successStyle.Render("✓ " + m.SuccessMessage))
		b.WriteString("\n\n")
	}

	// Show loading state
	if m.Loading {
		b.WriteString(loadingStyle.Render("Loading downloads..."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.CurrentView)))
		return b.String()
	}

	b.WriteString(headerStyle.Render("📥 Cached Downloads"))
	b.WriteString("\n\n")
	b.WriteString("This view will allow you to manage cached asset downloads.\n")
	b.WriteString("\n")

	b.WriteString(emptyStateStyle.Render("Downloads management features:"))
	b.WriteString("\n")
	b.WriteString("  • View all cached downloads with size and date\n")
	b.WriteString("  • Clear individual downloads to free up space\n")
	b.WriteString("  • Clear all downloads with confirmation\n")
	b.WriteString("  • View cache statistics and total size\n")
	b.WriteString("  • Verify checksums for cached files\n")
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("(Full implementation pending)"))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render(getHelpText(m.CurrentView)))

	return b.String()
}

// renderConfiguration renders the configuration view
func (m Model) RenderConfiguration() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.RenderTabsFn())

	// Show error if any
	if m.ErrorMessage != "" {
		b.WriteString(errorStyle.Render("Error: " + m.ErrorMessage))
		b.WriteString("\n\n")
	}

	// Show success message if any
	if m.SuccessMessage != "" {
		b.WriteString(successStyle.Render("✓ " + m.SuccessMessage))
		b.WriteString("\n\n")
	}

	// Show loading state
	if m.Loading {
		b.WriteString(loadingStyle.Render("Syncing configuration..."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.CurrentView)))
		return b.String()
	}

	// Configuration display
	b.WriteString(headerStyle.Render("Configuration Settings"))
	b.WriteString("\n\n")

	if m.Config != nil {
		b.WriteString(fmt.Sprintf("Version: %d\n", m.Config.Version))
		b.WriteString(fmt.Sprintf("Binaries in config: %d\n", len(m.Config.Binaries)))
		if m.Config.DateFormat != "" {
			b.WriteString(fmt.Sprintf("Date Format: %s\n", m.Config.DateFormat))
		}
		if m.Config.LogLevel != "" {
			b.WriteString(fmt.Sprintf("Log Level: %s\n", m.Config.LogLevel))
		}
		b.WriteString("\n")

		// Show first few binaries from config
		if len(m.Config.Binaries) > 0 {
			b.WriteString(headerStyle.Render("Configured Binaries:"))
			b.WriteString("\n")
			maxShow := 5
			if len(m.Config.Binaries) < maxShow {
				maxShow = len(m.Config.Binaries)
			}
			for i := 0; i < maxShow; i++ {
				binary := m.Config.Binaries[i]
				b.WriteString(fmt.Sprintf("  • %s (%s)\n", binary.Name, binary.Id))
			}
			if len(m.Config.Binaries) > maxShow {
				b.WriteString(fmt.Sprintf("  ... and %d more\n", len(m.Config.Binaries)-maxShow))
			}
		}
	} else {
		b.WriteString(emptyStateStyle.Render("No configuration loaded"))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("s: sync config to database • esc: back • q: quit"))

	return b.String()
}

// renderHelp renders the help view
func (m Model) RenderHelp() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.RenderTabsFn())

	b.WriteString(headerStyle.Render("Welcome to Binmate"))
	b.WriteString("\n\n")
	b.WriteString("Binmate helps you install and manage multiple versions of command-line binaries from GitHub releases.\n")
	b.WriteString("\n")

	// Binaries List View Help
	b.WriteString(headerStyle.Render("📦 Binaries List View"))
	b.WriteString("\n")
	b.WriteString("  ↑/↓      Navigate through binaries\n")
	b.WriteString("  Enter    View installed versions\n")
	b.WriteString("  a        Add new binary from GitHub URL\n")
	b.WriteString("  i        Install a specific version\n")
	b.WriteString("  u        Update selected binary to latest\n")
	b.WriteString("  U        Update all binaries to latest\n")
	b.WriteString("  r        Remove binary (with confirmation)\n")
	b.WriteString("  c        Check for updates without installing\n")
	b.WriteString("  m        Import existing binary from filesystem\n")
	b.WriteString("\n")

	// Versions View Help
	b.WriteString(headerStyle.Render("📋 Versions View"))
	b.WriteString("\n")
	b.WriteString("  ↑/↓      Navigate through installed versions\n")
	b.WriteString("  s/Enter  Switch to selected version\n")
	b.WriteString("  d        Delete selected version\n")
	b.WriteString("  Esc      Return to binaries list\n")
	b.WriteString("\n")

	// Configuration View Help
	b.WriteString(headerStyle.Render("⚙️  Configuration View"))
	b.WriteString("\n")
	b.WriteString("  s        Sync config file to database\n")
	b.WriteString("\n")

	// General Navigation
	b.WriteString(headerStyle.Render("🔄 General Navigation"))
	b.WriteString("\n")
	b.WriteString("  1-4      Switch between tabs directly\n")
	b.WriteString("  Tab      Cycle to next tab\n")
	b.WriteString("  Shift+Tab Cycle to previous tab\n")
	b.WriteString("  q        Quit application\n")
	b.WriteString("  Ctrl+C   Force quit\n")
	b.WriteString("\n")

	// Tips
	b.WriteString(headerStyle.Render("💡 Tips"))
	b.WriteString("\n")
	b.WriteString("  • Use 'latest' as version when installing to get the newest release\n")
	b.WriteString("  • Remove confirmations: 'y' for database only, 'Y' to also delete files\n")
	b.WriteString("  • Active versions are marked with ✓ in the versions view\n")
	b.WriteString("  • Success/error messages appear at the top of each view\n")
	b.WriteString("\n")

	b.WriteString(helpStyle.Render(getHelpText(m.CurrentView)))

	return b.String()
}
