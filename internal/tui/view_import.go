package tui

import (
	"strings"
)

// renderImportBinary renders the import binary view
func (m model) renderImportBinary() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📦 Import Existing Binary"))
	b.WriteString("\n\n")

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

	// Show form
	b.WriteString("Import an existing binary from your file system:\n")
	b.WriteString("\n")

	// Path input (focused if index 0)
	b.WriteString(formLabelStyle.Render("Binary Path: ") + requiredStyle.Render("*"))
	b.WriteString("\n")
	b.WriteString(m.importPathInput.View())
	b.WriteString("\n\n")

	// Name input (focused if index 1)
	b.WriteString(formLabelStyle.Render("Binary Name: ") + requiredStyle.Render("*"))
	b.WriteString("\n")
	b.WriteString(m.importNameInput.View())
	b.WriteString("\n\n")

	// URL input (focused if index 2) - optional
	b.WriteString(formLabelStyle.Render("GitHub Release URL: ") + helpStyle.Render("(optional - enables updates)"))
	b.WriteString("\n")
	b.WriteString(m.importURLInput.View())
	b.WriteString("\n\n")

	// Version input (focused if index 3) - optional
	b.WriteString(formLabelStyle.Render("Version: ") + helpStyle.Render("(optional - auto-extracted from URL)"))
	b.WriteString("\n")
	b.WriteString(m.importVersionInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("tab: next field • shift+tab: previous field • enter: import • esc: cancel"))

	return b.String()
}
