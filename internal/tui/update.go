package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"

	"cturner8/binmate/internal/core/crypto"
	urlparser "cturner8/binmate/internal/core/url"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

const (
	// ConfigVersionManual indicates a binary was added manually via TUI, not from config file
	ConfigVersionManual = 0
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

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

	case binarySavedMsg:
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
		} else {
			// Return to binaries list and reload
			m.currentView = viewBinariesList
			m.parsedBinary = nil
			m.formInputs = []textinput.Model{}
			m.errorMessage = ""
			m.loading = true
			return m, loadBinaries(m.dbService)
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
		case viewDownloads:
			return m.updatePlaceholderView(msg)
		case viewConfiguration:
			return m.updatePlaceholderView(msg)
		case viewHelp:
			return m.updatePlaceholderView(msg)
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
	// Check for tab switching
	if view, ok := getTabForKey(msg.String()); ok {
		m.currentView = view
		return m, nil
	}

	// Handle tab cycling
	switch msg.String() {
	case keyShiftTab:
		// Cycle to next tab
		m.currentView = getNextTab(m.currentView)
		return m, nil
	case keyCtrlShiftTab:
		// Cycle to previous tab
		m.currentView = getPreviousTab(m.currentView)
		return m, nil
	}

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
		m.urlTextInput.Reset()
		m.urlTextInput.Focus()
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
		m.urlTextInput.Reset()
		m.errorMessage = ""
		return m, nil

	case keyEnter:
		// Parse URL and transition to form view
		url := m.urlTextInput.Value()
		if url == "" {
			m.errorMessage = "Please enter a URL"
			return m, nil
		}

		parsed, err := urlparser.ParseGitHubReleaseURL(url)
		if err != nil {
			m.errorMessage = fmt.Sprintf("Invalid URL: %v", err)
			return m, nil
		}

		// Create parsed binary config
		m.parsedBinary = &parsedBinaryConfig{
			userID:   urlparser.GenerateBinaryID(parsed.AssetName),
			name:     urlparser.GenerateBinaryName(parsed.AssetName),
			provider: "github",
			path:     fmt.Sprintf("%s/%s", parsed.Owner, parsed.Repo),
			format:   parsed.Format,
			version:  parsed.Version,
			assetName: parsed.AssetName,
		}

		// Create form inputs
		m.formInputs = createFormInputs(m.parsedBinary)
		m.focusedField = 0
		m.formInputs[0].Focus()

		// Transition to form view
		m.currentView = viewAddBinaryForm
		m.errorMessage = ""
		return m, nil

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	// Handle text input
	var cmd tea.Cmd
	m.urlTextInput, cmd = m.urlTextInput.Update(msg)
	return m, cmd
}

// updateAddBinaryForm handles updates for the add binary form view
func (m model) updateAddBinaryForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		// Cancel and return to binaries list
		m.currentView = viewBinariesList
		m.parsedBinary = nil
		m.formInputs = []textinput.Model{}
		m.errorMessage = ""
		return m, nil

	case keyTab:
		// Move to next field
		m.formInputs[m.focusedField].Blur()
		m.focusedField = (m.focusedField + 1) % len(m.formInputs)
		m.formInputs[m.focusedField].Focus()
		return m, nil

	case "shift+tab":
		// Move to previous field
		m.formInputs[m.focusedField].Blur()
		m.focusedField = (m.focusedField - 1 + len(m.formInputs)) % len(m.formInputs)
		m.formInputs[m.focusedField].Focus()
		return m, nil

	case keySave:
		// Save binary to database
		return m, saveBinary(m)

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	// Handle text input for focused field
	var cmd tea.Cmd
	m.formInputs[m.focusedField], cmd = m.formInputs[m.focusedField].Update(msg)
	return m, cmd
}

// loadVersions returns a command to load versions for a binary
func loadVersions(dbService *repository.Service, binaryID int64) tea.Cmd {
	return func() tea.Msg {
		installations, err := getVersionsForBinary(dbService, binaryID)
		return versionsLoadedMsg{installations: installations, err: err}
	}
}

// createFormInputs creates text input fields for the binary form
func createFormInputs(parsed *parsedBinaryConfig) []textinput.Model {
	inputs := make([]textinput.Model, 8)

	// User ID
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Binary user ID"
	inputs[0].SetValue(parsed.userID)
	inputs[0].CharLimit = 64
	inputs[0].Width = 40

	// Name
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Binary name"
	inputs[1].SetValue(parsed.name)
	inputs[1].CharLimit = 64
	inputs[1].Width = 40

	// Provider (read-only)
	inputs[2] = textinput.New()
	inputs[2].SetValue(parsed.provider)
	inputs[2].CharLimit = 64
	inputs[2].Width = 40

	// Path
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "owner/repo"
	inputs[3].SetValue(parsed.path)
	inputs[3].CharLimit = 128
	inputs[3].Width = 40

	// Format
	inputs[4] = textinput.New()
	inputs[4].SetValue(parsed.format)
	inputs[4].CharLimit = 16
	inputs[4].Width = 40

	// Install Path (optional)
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Optional install path"
	inputs[5].SetValue(parsed.installPath)
	inputs[5].CharLimit = 256
	inputs[5].Width = 40

	// Asset Regex (optional)
	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Optional asset regex"
	inputs[6].SetValue(parsed.assetRegex)
	inputs[6].CharLimit = 256
	inputs[6].Width = 40

	// Release Regex (optional)
	inputs[7] = textinput.New()
	inputs[7].Placeholder = "Optional release regex"
	inputs[7].SetValue(parsed.releaseRegex)
	inputs[7].CharLimit = 256
	inputs[7].Width = 40

	return inputs
}

// saveBinary saves the binary configuration to the database
func saveBinary(m model) tea.Cmd {
	return func() tea.Msg {
		if m.parsedBinary == nil {
			return binarySavedMsg{err: fmt.Errorf("no binary data to save")}
		}

		// Get values from form inputs
		userID := m.formInputs[0].Value()
		name := m.formInputs[1].Value()
		provider := m.formInputs[2].Value()
		path := m.formInputs[3].Value()
		format := m.formInputs[4].Value()
		installPath := m.formInputs[5].Value()
		assetRegex := m.formInputs[6].Value()
		releaseRegex := m.formInputs[7].Value()

		// Validate required fields
		if userID == "" {
			return binarySavedMsg{err: fmt.Errorf("user ID is required")}
		}
		if name == "" {
			return binarySavedMsg{err: fmt.Errorf("name is required")}
		}
		if provider == "" {
			return binarySavedMsg{err: fmt.Errorf("provider is required")}
		}
		if path == "" {
			return binarySavedMsg{err: fmt.Errorf("path is required")}
		}
		if format == "" {
			return binarySavedMsg{err: fmt.Errorf("format is required")}
		}

		// Check if binary already exists
		existing, err := m.dbService.Binaries.GetByUserID(userID)
		if err != nil && err != database.ErrNotFound {
			return binarySavedMsg{err: fmt.Errorf("error checking for existing binary: %w", err)}
		}
		if existing != nil {
			return binarySavedMsg{err: fmt.Errorf("binary with ID '%s' already exists", userID)}
		}

		// Compute digest
		configDigest := crypto.ComputeDigest(
			userID, name, "", provider, path,
			installPath, format, assetRegex, releaseRegex,
		)

		// Create binary
		binary := &database.Binary{
			UserID:        userID,
			Name:          name,
			Provider:      provider,
			ProviderPath:  path,
			Format:        format,
			ConfigDigest:  configDigest,
			ConfigVersion: ConfigVersionManual, // Not from config file
		}

		// Set optional fields
		if installPath != "" {
			binary.InstallPath = &installPath
		}
		if assetRegex != "" {
			binary.AssetRegex = &assetRegex
		}
		if releaseRegex != "" {
			binary.ReleaseRegex = &releaseRegex
		}

		err = m.dbService.Binaries.Create(binary)
		if err != nil {
			return binarySavedMsg{err: fmt.Errorf("failed to create binary: %w", err)}
		}

		return binarySavedMsg{binary: binary, err: nil}
	}
}

// updatePlaceholderView handles updates for placeholder views (Downloads, Configuration, Help)
func (m model) updatePlaceholderView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check for tab switching
	if view, ok := getTabForKey(msg.String()); ok {
		m.currentView = view
		return m, nil
	}

	// Handle tab cycling
	switch msg.String() {
	case keyShiftTab:
		// Cycle to next tab
		m.currentView = getNextTab(m.currentView)
		return m, nil
	case keyCtrlShiftTab:
		// Cycle to previous tab
		m.currentView = getPreviousTab(m.currentView)
		return m, nil
	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	return m, nil
}
