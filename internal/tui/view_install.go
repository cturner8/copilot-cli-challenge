package tui

import (
	"fmt"
	"strings"
)

// renderInstallBinary renders the install binary view
func (m model) renderInstallBinary() string {
	var b strings.Builder

	// Get binary name if available
	binaryName := m.installBinaryID
	if len(m.binaries) > 0 && m.selectedIndex < len(m.binaries) {
		binaryName = m.binaries[m.selectedIndex].Binary.Name
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("📥 Install Binary - %s", binaryName)))
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

	// Show installing progress
	if m.installingInProgress {
		b.WriteString(loadingStyle.Render(fmt.Sprintf("Installing %s...", binaryName)))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("This may take a few moments depending on file size and network speed"))
		return b.String()
	}

	// Show form
	b.WriteString("Enter the version to install:\n")
	b.WriteString("\n")
	b.WriteString(formLabelStyle.Render("Version: "))
	b.WriteString("\n")
	b.WriteString(m.installVersionInput.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Tip: Use 'latest' to install the most recent version"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("enter: install • esc: cancel • q: quit"))

	return b.String()
}
