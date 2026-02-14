package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderReleaseNotes renders the release notes view
func (m model) renderReleaseNotes() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("📰 Release Notes"))
	b.WriteString("\n\n")

	// Show loading state
	if m.githubLoading {
		b.WriteString(loadingStyle.Render("Loading release notes..."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Show error if any
	if m.githubError != "" {
		b.WriteString(errorStyle.Render("Error: " + m.githubError))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Show release notes if available
	if m.githubReleaseInfo != nil {
		release := m.githubReleaseInfo

		// Header with version and date
		header := fmt.Sprintf("%s (%s)", release.TagName, release.Name)
		if release.Prerelease {
			header += " [Pre-release]"
		}
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Published: %s", release.PublishedAt))
		b.WriteString("\n")
		contentWidth := m.width - 4
		if contentWidth <= 0 {
			contentWidth = defaultTerminalWidth - 4
		}
		if contentWidth < 20 {
			contentWidth = 20
		}

		b.WriteString(lipgloss.NewStyle().Width(contentWidth).Render(fmt.Sprintf("URL: %s", release.HTMLURL)))
		b.WriteString("\n\n")

		// Release body
		b.WriteString(headerStyle.Render("Release Notes:"))
		b.WriteString("\n\n")

		if release.Body != "" {
			// Display release body with word wrapping
			lines := strings.Split(release.Body, "\n")
			for _, line := range lines {
				if line == "" {
					b.WriteString("\n")
					continue
				}
				b.WriteString(lipgloss.NewStyle().Width(contentWidth).Render(line))
				b.WriteString("\n")
			}
		} else {
			b.WriteString(mutedStyle.Render("No release notes available"))
			b.WriteString("\n")
		}
	} else {
		b.WriteString(emptyStateStyle.Render("No release notes available"))
		b.WriteString("\n\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(getHelpText(m.currentView)))

	return b.String()
}

// renderAvailableVersions renders the available versions view
func (m model) renderAvailableVersions() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("📦 Available Versions"))
	b.WriteString("\n\n")

	// Show binary name if available
	if m.selectedBinary != nil {
		b.WriteString(fmt.Sprintf("Binary: %s\n\n", m.selectedBinary.Name))
	}

	// Show loading state
	if m.githubLoading {
		b.WriteString(loadingStyle.Render("Loading available versions..."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Show error if any
	if m.githubError != "" {
		b.WriteString(errorStyle.Render("Error: " + m.githubError))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Show versions if available
	if len(m.githubAvailableVers) > 0 {
		// Calculate column widths
		availableWidth := m.width
		if availableWidth == 0 {
			availableWidth = defaultTerminalWidth
		}

		versionWidth := 20
		dateWidth := 20
		typeWidth := 15

		// Headers
		headers := []string{
			tableHeaderStyle.Width(versionWidth).Render("Version"),
			tableHeaderStyle.Width(dateWidth).Render("Published"),
			tableHeaderStyle.Width(typeWidth).Render("Type"),
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headers...))
		b.WriteString("\n")

		// Separator
		b.WriteString(strings.Repeat("─", versionWidth+dateWidth+typeWidth+4))
		b.WriteString("\n")

		// Rows
		for i, release := range m.githubAvailableVers {
			version := truncateText(release.TagName, versionWidth)
			date := truncateText(release.PublishedAt, dateWidth)
			releaseType := "Release"
			if release.Prerelease {
				releaseType = "Pre-release"
			}
			releaseType = truncateText(releaseType, typeWidth)

			rowStyle := normalStyle
			if i == m.selectedAvailableVersionIdx {
				rowStyle = selectedStyle
			}

			row := []string{
				rowStyle.Width(versionWidth).Render(version),
				rowStyle.Width(dateWidth).Render(date),
				rowStyle.Width(typeWidth).Render(releaseType),
			}

			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row...))
			b.WriteString("\n")
		}
	} else {
		b.WriteString(emptyStateStyle.Render("No versions available"))
		b.WriteString("\n\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(getHelpText(m.currentView)))

	return b.String()
}

// renderRepositoryInfo renders the repository information view
func (m model) renderRepositoryInfo() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("ℹ️  Repository Information"))
	b.WriteString("\n\n")

	// Show success message if any
	if m.successMessage != "" {
		b.WriteString(successStyle.Render("✓ " + m.successMessage))
		b.WriteString("\n\n")
	}

	// Show loading state
	if m.githubLoading {
		b.WriteString(loadingStyle.Render("Loading repository information..."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Show error if any
	if m.githubError != "" {
		b.WriteString(errorStyle.Render("Error: " + m.githubError))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
		return b.String()
	}

	// Show repository info if available
	if m.githubRepoInfo != nil {
		repo := m.githubRepoInfo

		// Repository name
		b.WriteString(headerStyle.Render(repo.FullName))
		b.WriteString("\n\n")

		// Description
		if repo.Description != "" {
			b.WriteString(headerStyle.Render("Description:"))
			b.WriteString("\n")
			b.WriteString(repo.Description)
			b.WriteString("\n\n")
		}

		// Stats
		b.WriteString(headerStyle.Render("Statistics:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("⭐ Stars: %s\n", formatIntWithCommas(repo.Stars)))
		b.WriteString(fmt.Sprintf("🔀 Forks: %s\n", formatIntWithCommas(repo.Forks)))
		b.WriteString("\n")

		// URL
		b.WriteString(headerStyle.Render("URL:"))
		b.WriteString("\n")
		b.WriteString(repo.HTMLURL)
		b.WriteString("\n")
	} else {
		b.WriteString(emptyStateStyle.Render("No repository information available"))
		b.WriteString("\n\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(getHelpText(m.currentView)))

	return b.String()
}
