package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderVersions renders the versions detail view
func (m model) renderVersions() string {
	var b strings.Builder

	// Title
	if m.selectedBinary != nil {
		title := fmt.Sprintf("📦 %s - Installed Versions", m.selectedBinary.Name)
		b.WriteString(titleStyle.Render(title))
		b.WriteString("\n\n")

		// Binary details
		b.WriteString(headerStyle.Render("Binary Details:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Provider: %s\n", m.selectedBinary.Provider))
		b.WriteString(fmt.Sprintf("Path: %s\n", m.selectedBinary.ProviderPath))
		b.WriteString(fmt.Sprintf("Format: %s\n", m.selectedBinary.Format))
		b.WriteString("\n")
	}

	// Show loading state
	if m.loading {
		b.WriteString(loadingStyle.Render("Loading versions..."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Show empty state
	if len(m.installations) == 0 {
		b.WriteString(emptyStateStyle.Render("No versions installed"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Show error if any
	if m.errorMessage != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMessage))
		b.WriteString("\n\n")
	}

	// Get active version
	var activeInstallationID int64
	if m.selectedBinary != nil {
		activeVersion, _ := getActiveVersion(m.dbService, m.selectedBinary.ID)
		if activeVersion != nil {
			activeInstallationID = activeVersion.ID
		}
	}

	// Table header
	activeWidth := 4
	versionWidth := 20
	installedWidth := 20
	sizeWidth := 12
	pathWidth := 30

	headers := []string{
		tableHeaderStyle.Width(activeWidth).Render(""),
		tableHeaderStyle.Width(versionWidth).Render("Version"),
		tableHeaderStyle.Width(installedWidth).Render("Installed"),
		tableHeaderStyle.Width(sizeWidth).Render("Size"),
		tableHeaderStyle.Width(pathWidth).Render("Path"),
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headers...))
	b.WriteString("\n")

	// Separator line
	b.WriteString(strings.Repeat("─", activeWidth+versionWidth+installedWidth+sizeWidth+pathWidth+8))
	b.WriteString("\n")

	// Table rows
	for _, installation := range m.installations {
		// Active indicator
		activeIndicator := ""
		if installation.ID == activeInstallationID {
			activeIndicator = activeIndicatorStyle.Render("✓")
		}

		// Version
		version := installation.Version
		if len(version) > versionWidth-2 {
			version = version[:versionWidth-5] + "..."
		}

		// Installed date
		dateFormat := "2006-01-02 15:04" // Default ISO format
		if m.config != nil && m.config.DateFormat != "" {
			dateFormat = m.config.DateFormat
		}
		installedDate := time.Unix(installation.InstalledAt, 0).Format(dateFormat)

		// File size (human-readable)
		size := formatBytes(installation.FileSize)

		// Install path
		path := installation.InstalledPath
		if len(path) > pathWidth-2 {
			path = "..." + path[len(path)-(pathWidth-5):]
		}

		row := []string{
			tableCellStyle.Width(activeWidth).Render(activeIndicator),
			tableCellStyle.Width(versionWidth).Render(version),
			tableCellStyle.Width(installedWidth).Render(installedDate),
			tableCellStyle.Width(sizeWidth).Render(size),
			tableCellStyle.Width(pathWidth).Render(path),
		}

		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row...))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(getHelpText(m.currentView)))

	return b.String()
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	// Unit prefixes: Kilo, Mega, Giga, Tera, Peta, Exa
	const unitPrefixes = "KMGTPE"
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), unitPrefixes[exp])
}
