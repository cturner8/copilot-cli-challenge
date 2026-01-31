package tui

import (
	"strings"
)

// renderAddBinaryURL renders the add binary URL input view
func (m model) renderAddBinaryURL() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("➕ Add Binary - Enter GitHub Release URL"))
	b.WriteString("\n\n")
	
	if m.errorMessage != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMessage))
		b.WriteString("\n\n")
	}

	b.WriteString("Enter the GitHub release URL for the binary you want to add:\n")
	b.WriteString("\n")
	b.WriteString("Example: https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz\n")
	b.WriteString("\n")
	b.WriteString(formLabelStyle.Render("URL: "))
	b.WriteString(formInputStyle.Render(m.urlInput))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(getHelpText(m.currentView)))

	return b.String()
}

// renderAddBinaryForm renders the add binary configuration form view
func (m model) renderAddBinaryForm() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("➕ Add Binary - Configuration"))
	b.WriteString("\n\n")

	if m.errorMessage != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMessage))
		b.WriteString("\n\n")
	}

	if m.parsedBinary != nil {
		b.WriteString("Configure the binary details:\n")
		b.WriteString("\n")
		b.WriteString(formLabelStyle.Render("User ID: "))
		b.WriteString(m.parsedBinary.userID)
		b.WriteString("\n")
		b.WriteString(formLabelStyle.Render("Name: "))
		b.WriteString(m.parsedBinary.name)
		b.WriteString("\n")
		b.WriteString(formLabelStyle.Render("Provider: "))
		b.WriteString(m.parsedBinary.provider)
		b.WriteString("\n")
		b.WriteString(formLabelStyle.Render("Path: "))
		b.WriteString(m.parsedBinary.path)
		b.WriteString("\n")
		b.WriteString(formLabelStyle.Render("Format: "))
		b.WriteString(m.parsedBinary.format)
		b.WriteString("\n\n")
		b.WriteString(emptyStateStyle.Render("(Form editing not yet implemented)"))
	} else {
		b.WriteString(emptyStateStyle.Render("No binary data available"))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(getHelpText(m.currentView)))

	return b.String()
}
