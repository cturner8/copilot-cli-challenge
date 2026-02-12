package views

import (
	"fmt"
	"strings"
)

// renderInstallBinary renders the install binary view
func (m Model) RenderInstallBinary() string {
	var b strings.Builder

	// Get binary name if available
	binaryName := m.InstallBinaryID
	if len(m.Binaries) > 0 && m.SelectedIndex < len(m.Binaries) {
		binaryName = m.Binaries[m.SelectedIndex].Binary.Name
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("📥 Install Binary - %s", binaryName)))
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

	// Show installing progress
	if m.InstallingInProgress {
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
	b.WriteString(m.InstallVersionInput.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Tip: Use 'latest' to install the most recent version"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("enter: install • esc: cancel • q: quit"))

	return b.String()
}
