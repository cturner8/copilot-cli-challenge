package views

import (
	"strings"
)

// renderImportBinary renders the import binary view
func (m Model) RenderImportBinary() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📦 Import Existing Binary"))
	b.WriteString("\n\n")

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

	// Show form
	b.WriteString("Import an existing binary from your file system:\n")
	b.WriteString("\n")

	// Path input (focused if index 0)
	b.WriteString(formLabelStyle.Render("Binary Path: "))
	b.WriteString("\n")
	b.WriteString(m.ImportPathInput.View())
	b.WriteString("\n\n")

	// Name input (focused if index 1)
	b.WriteString(formLabelStyle.Render("Binary Name: "))
	b.WriteString("\n")
	b.WriteString(m.ImportNameInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Note: Import functionality is not yet fully implemented in the service layer"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("tab: next field • enter: import • esc: cancel • q: quit"))

	return b.String()
}
