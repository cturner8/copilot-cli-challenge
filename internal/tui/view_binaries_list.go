package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderBinariesList renders the binaries list view
func (m model) renderBinariesList() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())

	// Show loading state
	if m.loading {
		b.WriteString(loadingStyle.Render("Loading binaries..."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Show empty state
	if len(m.binaries) == 0 {
		b.WriteString(emptyStateStyle.Render("No binaries configured"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press 'a' to add a binary"))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Show error if any
	if m.errorMessage != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMessage))
		b.WriteString("\n\n")
	}

	// Calculate proportional column widths based on available width
	// Default to 80 if width not set
	availableWidth := m.width
	if availableWidth == 0 {
		availableWidth = 80
	}
	
	// Allocate proportional widths: Name 35%, Provider 15%, Version 30%, Count 20%
	totalWidth := availableWidth - 8 // Account for padding
	nameWidth := int(float64(totalWidth) * 0.35)
	providerWidth := int(float64(totalWidth) * 0.15)
	versionWidth := int(float64(totalWidth) * 0.30)
	countWidth := int(float64(totalWidth) * 0.20)

	headers := []string{
		tableHeaderStyle.Width(nameWidth).Render("Name"),
		tableHeaderStyle.Width(providerWidth).Render("Provider"),
		tableHeaderStyle.Width(versionWidth).Render("Active Version"),
		tableHeaderStyle.Width(countWidth).Render("Installed"),
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headers...))
	b.WriteString("\n")

	// Separator line
	b.WriteString(strings.Repeat("─", nameWidth+providerWidth+versionWidth+countWidth+8))
	b.WriteString("\n")

	// Table rows
	for i, binary := range m.binaries {
		style := normalStyle
		if i == m.selectedIndex {
			style = selectedStyle
		}

		name := binary.Binary.Name
		if len(name) > nameWidth-2 {
			name = name[:nameWidth-5] + "..."
		}

		provider := binary.Binary.Provider
		if len(provider) > providerWidth-2 {
			provider = provider[:providerWidth-5] + "..."
		}

		version := binary.ActiveVersion
		if len(version) > versionWidth-2 {
			version = version[:versionWidth-5] + "..."
		}

		count := fmt.Sprintf("%d", binary.InstallCount)

		row := []string{
			style.Width(nameWidth).Render(name),
			style.Width(providerWidth).Render(provider),
			style.Width(versionWidth).Render(version),
			style.Width(countWidth).Render(count),
		}

		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row...))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(getHelpText(m.currentView)))

	return b.String()
}
