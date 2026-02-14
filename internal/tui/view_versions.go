package tui

import (
	"fmt"
	"strings"

	"cturner8/binmate/internal/core/format"

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

	// Show success message if any
	if m.successMessage != "" {
		b.WriteString(successStyle.Render("✓ " + m.successMessage))
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

	// Calculate proportional column widths based on available width
	availableWidth := m.width
	if availableWidth == 0 {
		availableWidth = defaultTerminalWidth
	}

	// Account for padding: 2 chars per column (5 columns = 10 total)
	totalWidth := availableWidth - columnPadding5

	// Allocate proportional widths: Active 5%, Version 20%, Installed 20%, Size 15%, Path 40%
	activeWidth := int(float64(totalWidth) * 0.05)
	versionWidth := int(float64(totalWidth) * 0.20)
	installedWidth := int(float64(totalWidth) * 0.20)
	sizeWidth := int(float64(totalWidth) * 0.15)
	pathWidth := int(float64(totalWidth) * 0.40)

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
	b.WriteString(strings.Repeat("─", activeWidth+versionWidth+installedWidth+sizeWidth+pathWidth+columnPadding5))
	b.WriteString("\n")

	// Get date format once before loop
	dateFormat := getDateFormat(m.config)

	// Table rows
	for i, installation := range m.installations {
		// Determine row style (selected or normal)
		rowStyle := tableCellStyle
		if i == m.selectedVersionIdx {
			rowStyle = selectedStyle
		}

		// Active indicator
		activeIndicator := ""
		if installation.ID == activeInstallationID {
			activeIndicator = activeIndicatorStyle.Render("✓")
		}

		// Version
		version := truncateText(installation.Version, versionWidth)

		// Installed date
		installedDate := format.FormatTimestamp(installation.InstalledAt, dateFormat)

		// File size (human-readable)
		size := formatBytes(installation.FileSize)

		// Install path (truncate from beginning, keep end)
		path := truncatePathEnd(installation.InstalledPath, pathWidth)

		row := []string{
			rowStyle.Width(activeWidth).Render(activeIndicator),
			rowStyle.Width(versionWidth).Render(version),
			rowStyle.Width(installedWidth).Render(installedDate),
			rowStyle.Width(sizeWidth).Render(size),
			rowStyle.Width(pathWidth).Render(path),
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
