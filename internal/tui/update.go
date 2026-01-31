package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"cturner8/binmate/internal/database/repository"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case binariesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
		} else {
			m.binaries = msg.binaries
			m.errorMessage = ""
		}
		return m, nil

	case versionsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
		} else {
			m.installations = msg.installations
			m.errorMessage = ""
		}
		return m, nil

	case tea.KeyMsg:
		switch m.currentView {
		case viewBinariesList:
			return m.updateBinariesList(msg)
		case viewVersions:
			return m.updateVersions(msg)
		case viewAddBinaryURL:
			return m.updateAddBinaryURL(msg)
		case viewAddBinaryForm:
			return m.updateAddBinaryForm(msg)
		}

		// Global key handlers
		switch msg.String() {
		case keyCtrlC, keyQuit:
			return m, tea.Quit
		}
	}

	return m, nil
}

// updateBinariesList handles updates for the binaries list view
func (m model) updateBinariesList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp:
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case keyDown:
		if m.selectedIndex < len(m.binaries)-1 {
			m.selectedIndex++
		}

	case keyEnter:
		// Transition to versions view
		if len(m.binaries) > 0 && m.selectedIndex < len(m.binaries) {
			m.currentView = viewVersions
			m.selectedBinary = m.binaries[m.selectedIndex].Binary
			m.loading = true
			return m, loadVersions(m.dbService, m.selectedBinary.ID)
		}

	case keyAdd:
		// Transition to add binary view
		m.currentView = viewAddBinaryURL
		m.urlInput = ""
		m.errorMessage = ""

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	return m, nil
}

// updateVersions handles updates for the versions view
func (m model) updateVersions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		// Return to binaries list
		m.currentView = viewBinariesList
		m.selectedBinary = nil
		m.installations = nil
		m.errorMessage = ""

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	return m, nil
}

// updateAddBinaryURL handles updates for the add binary URL view
func (m model) updateAddBinaryURL(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		// Cancel and return to binaries list
		m.currentView = viewBinariesList
		m.urlInput = ""
		m.errorMessage = ""

	case keyEnter:
		// TODO: Parse URL and transition to form view
		m.errorMessage = "URL parsing not yet implemented"

	case keyQuit, keyCtrlC:
		return m, tea.Quit

	default:
		// Handle text input
		// TODO: Implement text input handling
	}

	return m, nil
}

// updateAddBinaryForm handles updates for the add binary form view
func (m model) updateAddBinaryForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		// Cancel and return to binaries list
		m.currentView = viewBinariesList
		m.parsedBinary = nil
		m.formFields = make(map[string]string)
		m.errorMessage = ""

	case keySave:
		// TODO: Save binary to database
		m.errorMessage = "Save not yet implemented"

	case keyQuit, keyCtrlC:
		return m, tea.Quit

	default:
		// TODO: Handle form input
	}

	return m, nil
}

// loadVersions returns a command to load versions for a binary
func loadVersions(dbService *repository.Service, binaryID int64) tea.Cmd {
	return func() tea.Msg {
		installations, err := getVersionsForBinary(dbService, binaryID)
		return versionsLoadedMsg{installations: installations, err: err}
	}
}
