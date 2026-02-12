package views

import (
	"fmt"
	"strings"
)

// renderAddBinaryURL renders the add binary URL input view
func (m Model) RenderAddBinaryURL() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("➕ Add Binary - Enter GitHub Release URL"))
	b.WriteString("\n\n")

	if m.ErrorMessage != "" {
		b.WriteString(errorStyle.Render("Error: " + m.ErrorMessage))
		b.WriteString("\n\n")
	}

	b.WriteString("Enter the GitHub release URL for the binary you want to add:\n")
	b.WriteString("\n")
	b.WriteString("Example: https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz\n")
	b.WriteString("\n")
	b.WriteString(formLabelStyle.Render("URL: "))
	b.WriteString("\n")
	b.WriteString(m.UrlTextInput.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(getHelpText(m.CurrentView)))

	return b.String()
}

// renderAddBinaryForm renders the add binary configuration form view
func (m Model) RenderAddBinaryForm() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("➕ Add Binary - Configuration"))
	b.WriteString("\n\n")

	if m.ErrorMessage != "" {
		b.WriteString(errorStyle.Render("Error: " + m.ErrorMessage))
		b.WriteString("\n\n")
	}

	if m.ParsedBinary == nil {
		b.WriteString(emptyStateStyle.Render("No binary data available"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render(getHelpText(m.CurrentView)))
		return b.String()
	}

	b.WriteString("Configure the binary details:\n")
	b.WriteString("\n")

	// Field labels
	fieldLabels := []string{
		"User ID",
		"Name",
		"Provider",
		"Path",
		"Format",
		"Install Path",
		"Asset Regex",
		"Release Regex",
		"Authenticated",
	}

	// Render each form field
	for i, label := range fieldLabels {
		labelStr := formLabelStyle.Render(fmt.Sprintf("%-15s: ", label))
		b.WriteString(labelStr)

		if i < len(m.FormInputs) {
			b.WriteString(m.FormInputs[i].View())
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(getHelpText(m.CurrentView)))

	return b.String()
}
