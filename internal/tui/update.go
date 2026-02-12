package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	binarySvc "cturner8/binmate/internal/core/binary"
	configPkg "cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/core/crypto"
	installSvc "cturner8/binmate/internal/core/install"
	urlparser "cturner8/binmate/internal/core/url"
	versionSvc "cturner8/binmate/internal/core/version"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
	"cturner8/binmate/internal/providers/github"
	"cturner8/binmate/internal/tui/views"
)

const (
	// ConfigVersionManual indicates a binary was added manually via TUI, not from config file
	ConfigVersionManual = 0
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case binariesLoadedMsg:
		m.Loading = false
		if msg.err != nil {
			m.ErrorMessage = msg.err.Error()
		} else {
			m.Binaries = msg.binaries
			m.ErrorMessage = ""
		}
		return m, nil

	case versionsLoadedMsg:
		m.Loading = false
		if msg.err != nil {
			m.ErrorMessage = msg.err.Error()
		} else {
			m.Installations = msg.installations
			m.ErrorMessage = ""
		}
		return m, nil

	case binarySavedMsg:
		if msg.err != nil {
			m.ErrorMessage = msg.err.Error()
		} else {
			// Return to binaries list and reload
			m.CurrentView = views.BinariesList
			m.ParsedBinary = nil
			m.FormInputs = []textinput.Model{}
			m.ErrorMessage = ""
			m.SuccessMessage = fmt.Sprintf("Binary %s added successfully", msg.binary.Name)
			m.Loading = true
			return m, loadBinaries(m.DbService)
		}
		return m, nil

	case binaryInstalledMsg:
		m.InstallingInProgress = false
		if msg.err != nil {
			m.ErrorMessage = msg.err.Error()
			m.SuccessMessage = ""
		} else {
			// Return to the view we came from (binaries list or versions)
			returnView := m.InstallReturnView
			if returnView == views.ViewState(0) {
				returnView = views.BinariesList // Default to binaries list
			}
			m.CurrentView = returnView
			m.InstallVersionInput.Reset()
			m.InstallBinaryID = ""
			m.ErrorMessage = ""
			m.SuccessMessage = fmt.Sprintf("Successfully installed %s version %s", msg.binary.Name, msg.installation.Version)

			// Reload appropriate data based on return view
			m.Loading = true
			if returnView == views.Versions && m.SelectedBinary != nil {
				return m, loadVersions(m.DbService, m.SelectedBinary.ID)
			}
			return m, loadBinaries(m.DbService)
		}
		return m, nil

	case binaryUpdatedMsg:
		if msg.err != nil {
			m.ErrorMessage = msg.err.Error()
			m.SuccessMessage = ""
		} else {
			m.ErrorMessage = ""
			if msg.oldVersion == msg.newVersion {
				m.SuccessMessage = fmt.Sprintf("%s is already up to date (%s)", msg.binaryID, msg.newVersion)
			} else {
				m.SuccessMessage = fmt.Sprintf("Updated %s from %s to %s", msg.binaryID, msg.oldVersion, msg.newVersion)
			}
			// Reload appropriate data based on current view
			m.Loading = true
			if m.CurrentView == views.Versions && m.SelectedBinary != nil {
				return m, loadVersions(m.DbService, m.SelectedBinary.ID)
			}
			return m, loadBinaries(m.DbService)
		}
		return m, nil

	case binaryRemovedMsg:
		if msg.err != nil {
			m.ErrorMessage = msg.err.Error()
			m.SuccessMessage = ""
		} else {
			m.ErrorMessage = ""
			m.SuccessMessage = fmt.Sprintf("Binary %s removed successfully", msg.binaryID)
			// Reload binaries list
			m.Loading = true
			return m, loadBinaries(m.DbService)
		}
		return m, nil

	case updateCheckMsg:
		m.Loading = false
		if msg.err != nil {
			m.ErrorMessage = msg.err.Error()
			m.SuccessMessage = ""
		} else {
			m.ErrorMessage = ""
			if msg.hasUpdate {
				m.SuccessMessage = fmt.Sprintf("⬆ Update available for %s: %s → %s", msg.binaryID, msg.currentVersion, msg.latestVersion)
			} else if msg.latestInstalled {
				m.SuccessMessage = fmt.Sprintf("✓ %s: latest version (%s) installed but not active (current: %s)", msg.binaryID, msg.latestVersion, msg.currentVersion)
			} else {
				m.SuccessMessage = fmt.Sprintf("✓ %s is up to date (%s)", msg.binaryID, msg.currentVersion)
			}
		}
		return m, nil

	case configSyncedMsg:
		m.Loading = false
		if msg.err != nil {
			m.ErrorMessage = msg.err.Error()
			m.SuccessMessage = ""
		} else {
			m.ErrorMessage = ""
			m.SuccessMessage = "Configuration synced to database successfully"
			// Reload binaries list
			return m, loadBinaries(m.DbService)
		}
		return m, nil

	case versionSwitchedMsg:
		if msg.err != nil {
			m.ErrorMessage = msg.err.Error()
			m.SuccessMessage = ""
		} else {
			m.ErrorMessage = ""
			m.SuccessMessage = fmt.Sprintf("Switched to version %s", msg.installation.Version)
			// Reload versions to update active indicator
			m.Loading = true
			return m, loadVersions(m.DbService, m.SelectedBinary.ID)
		}
		return m, nil

	case versionDeletedMsg:
		if msg.err != nil {
			m.ErrorMessage = msg.err.Error()
			m.SuccessMessage = ""
		} else {
			m.ErrorMessage = ""
			m.SuccessMessage = "Version deleted successfully"
			// Reload versions list
			m.Loading = true
			return m, loadVersions(m.DbService, m.SelectedBinary.ID)
		}
		return m, nil

	case successMsg:
		m.SuccessMessage = msg.message
		m.ErrorMessage = ""
		return m, nil

	case errorMsg:
		m.ErrorMessage = msg.err.Error()
		m.SuccessMessage = ""
		return m, nil

	case tea.KeyMsg:
		switch m.CurrentView {
		case views.BinariesList:
			return m.updateBinariesList(msg)
		case views.Versions:
			return m.updateVersions(msg)
		case views.AddBinaryURL:
			return m.updateAddBinaryURL(msg)
		case views.AddBinaryForm:
			return m.updateAddBinaryForm(msg)
		case views.InstallBinary:
			return m.updateInstallBinary(msg)
		case views.ImportBinary:
			return m.updateImportBinary(msg)
		case views.Downloads:
			return m.updatePlaceholderView(msg)
		case views.Configuration:
			return m.updatePlaceholderView(msg)
		case views.Help:
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
func (m Model) updateBinariesList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle remove confirmation if active
	if m.ConfirmingRemove {
		switch msg.String() {
		case "y":
			// Remove without files
			binaryID := m.RemoveBinaryID
			m.ConfirmingRemove = false
			m.RemoveBinaryID = ""
			return m, removeBinary(m.DbService, binaryID, false)
		case "Y":
			// Remove with files
			binaryID := m.RemoveBinaryID
			m.ConfirmingRemove = false
			m.RemoveBinaryID = ""
			return m, removeBinary(m.DbService, binaryID, true)
		case "n", keyEsc:
			// Cancel
			m.ConfirmingRemove = false
			m.RemoveBinaryID = ""
			return m, nil
		}
		return m, nil
	}

	// Check for tab switching
	if view, ok := getTabForKey(msg.String()); ok {
		m.CurrentView = view
		return m, nil
	}

	// Handle tab cycling
	if updatedModel, handled := handleTabCycling(m, msg.String()); handled {
		return updatedModel, nil
	}

	switch msg.String() {
	case keyUp:
		if m.SelectedIndex > 0 {
			m.SelectedIndex--
		}

	case keyDown:
		if m.SelectedIndex < len(m.Binaries)-1 {
			m.SelectedIndex++
		}

	case keyEnter:
		// Transition to versions view
		if len(m.Binaries) > 0 && m.SelectedIndex < len(m.Binaries) {
			m.CurrentView = views.Versions
			m.SelectedBinary = m.Binaries[m.SelectedIndex].Binary
			m.Loading = true
			return m, loadVersions(m.DbService, m.SelectedBinary.ID)
		}

	case keyAdd:
		// Transition to add binary view
		m.CurrentView = views.AddBinaryURL
		m.UrlTextInput.Reset()
		m.UrlTextInput.Focus()
		m.ErrorMessage = ""
		m.SuccessMessage = ""

	case keyInstall:
		// Transition to install binary view
		if len(m.Binaries) > 0 && m.SelectedIndex < len(m.Binaries) {
			m.CurrentView = views.InstallBinary
			m.InstallBinaryID = m.Binaries[m.SelectedIndex].Binary.UserID
			m.InstallReturnView = views.BinariesList
			m.InstallVersionInput.Focus()
			m.ErrorMessage = ""
			m.SuccessMessage = ""
		}

	case keyUpdate:
		// Update selected binary to latest version
		if len(m.Binaries) > 0 && m.SelectedIndex < len(m.Binaries) {
			selectedBinary := m.Binaries[m.SelectedIndex]
			m.ErrorMessage = ""
			m.SuccessMessage = ""
			return m, updateBinary(m.DbService, selectedBinary.Binary.UserID)
		}

	case keyUpdateAll:
		// Update all binaries to latest version
		if len(m.Binaries) > 0 {
			m.ErrorMessage = ""
			m.SuccessMessage = ""
			m.Loading = true
			return m, updateAllBinaries(m.DbService, m.Binaries)
		}

	case keyRemove:
		// Show remove confirmation for selected binary
		if len(m.Binaries) > 0 && m.SelectedIndex < len(m.Binaries) {
			m.ConfirmingRemove = true
			m.RemoveBinaryID = m.Binaries[m.SelectedIndex].Binary.UserID
			m.ErrorMessage = ""
			m.SuccessMessage = ""
		}

	case keyCheck:
		// Check for updates for selected binary
		if len(m.Binaries) > 0 && m.SelectedIndex < len(m.Binaries) {
			selectedBinary := m.Binaries[m.SelectedIndex]
			m.ErrorMessage = ""
			m.SuccessMessage = ""
			m.Loading = true
			return m, checkForUpdates(m.DbService, selectedBinary.Binary.UserID)
		}

	case keyImport:
		// Transition to import binary view
		m.CurrentView = views.ImportBinary
		m.ImportPathInput.Reset()
		m.ImportNameInput.Reset()
		m.ImportFocusIdx = 0
		m.ImportPathInput.Focus()
		m.ErrorMessage = ""
		m.SuccessMessage = ""

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	return m, nil
}

// updateVersions handles updates for the versions view
func (m Model) updateVersions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp:
		if m.SelectedVersionIdx > 0 {
			m.SelectedVersionIdx--
		}

	case keyDown:
		if m.SelectedVersionIdx < len(m.Installations)-1 {
			m.SelectedVersionIdx++
		}

	case keySwitch, keyEnter:
		// Switch to selected version
		if len(m.Installations) > 0 && m.SelectedVersionIdx < len(m.Installations) {
			selectedInstallation := m.Installations[m.SelectedVersionIdx]
			return m, switchVersion(m.DbService, m.SelectedBinary, selectedInstallation)
		}

	case keyInstall:
		// Install a new version
		if m.SelectedBinary != nil {
			m.CurrentView = views.InstallBinary
			m.InstallBinaryID = m.SelectedBinary.UserID
			m.InstallReturnView = views.Versions
			m.InstallVersionInput.Focus()
			m.ErrorMessage = ""
			m.SuccessMessage = ""
		}

	case keyUpdate:
		// Update to latest version
		if m.SelectedBinary != nil {
			m.ErrorMessage = ""
			m.SuccessMessage = ""
			return m, updateBinary(m.DbService, m.SelectedBinary.UserID)
		}

	case keyCheck:
		// Check for updates
		if m.SelectedBinary != nil {
			m.ErrorMessage = ""
			m.SuccessMessage = ""
			m.Loading = true
			return m, checkForUpdates(m.DbService, m.SelectedBinary.UserID)
		}

	case keyDelete, keyDelete2:
		// Delete selected version
		if len(m.Installations) > 0 && m.SelectedVersionIdx < len(m.Installations) {
			selectedInstallation := m.Installations[m.SelectedVersionIdx]

			// Check if this is the active version
			activeVersion, _ := getActiveVersion(m.DbService, m.SelectedBinary.ID)
			if activeVersion != nil && activeVersion.ID == selectedInstallation.ID {
				m.ErrorMessage = "Cannot delete active version. Switch to another version first."
				m.SuccessMessage = ""
				return m, nil
			}

			return m, deleteVersion(m.DbService, selectedInstallation)
		}

	case keyEsc:
		// Return to binaries list
		m.CurrentView = views.BinariesList
		m.SelectedBinary = nil
		m.Installations = nil
		m.SelectedVersionIdx = 0
		m.ErrorMessage = ""
		m.SuccessMessage = ""

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	return m, nil
}

// updateAddBinaryURL handles updates for the add binary URL view
func (m Model) updateAddBinaryURL(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		// Cancel and return to binaries list
		m.CurrentView = views.BinariesList
		m.UrlTextInput.Reset()
		m.ErrorMessage = ""
		return m, nil

	case keyEnter:
		// Parse URL and transition to form view
		url := m.UrlTextInput.Value()
		if url == "" {
			m.ErrorMessage = "Please enter a URL"
			return m, nil
		}

		parsed, err := urlparser.ParseGitHubReleaseURL(url)
		if err != nil {
			m.ErrorMessage = fmt.Sprintf("Invalid URL: %v", err)
			return m, nil
		}

		// Create parsed binary config
		m.ParsedBinary = &parsedBinaryConfig{
			userID:    urlparser.GenerateBinaryID(parsed.AssetName),
			name:      urlparser.GenerateBinaryName(parsed.AssetName),
			provider:  "github",
			path:      fmt.Sprintf("%s/%s", parsed.Owner, parsed.Repo),
			format:    parsed.Format,
			version:   parsed.Version,
			assetName: parsed.AssetName,
		}

		// Create form inputs
		m.FormInputs = createFormInputs(m.ParsedBinary)
		m.FocusedField = 0
		m.FormInputs[0].Focus()

		// Transition to form view
		m.CurrentView = views.AddBinaryForm
		m.ErrorMessage = ""
		return m, nil

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	// Handle text input
	var cmd tea.Cmd
	m.UrlTextInput, cmd = m.UrlTextInput.Update(msg)
	return m, cmd
}

// updateAddBinaryForm handles updates for the add binary form view
func (m Model) updateAddBinaryForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		// Cancel and return to binaries list
		m.CurrentView = views.BinariesList
		m.ParsedBinary = nil
		m.FormInputs = []textinput.Model{}
		m.ErrorMessage = ""
		return m, nil

	case keyTab:
		// Move to next field
		m.FormInputs[m.FocusedField].Blur()
		m.FocusedField = (m.FocusedField + 1) % len(m.FormInputs)
		m.FormInputs[m.FocusedField].Focus()
		return m, nil

	case "shift+tab":
		// Move to previous field
		m.FormInputs[m.FocusedField].Blur()
		m.FocusedField = (m.FocusedField - 1 + len(m.FormInputs)) % len(m.FormInputs)
		m.FormInputs[m.FocusedField].Focus()
		return m, nil

	case keySave:
		// Save binary to database
		return m, saveBinary(m)

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	// Handle text input for focused field
	var cmd tea.Cmd
	m.FormInputs[m.FocusedField], cmd = m.FormInputs[m.FocusedField].Update(msg)
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
	inputs := make([]textinput.Model, 9)

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

	// Authenticated (boolean as string)
	inputs[8] = textinput.New()
	inputs[8].Placeholder = "true/false (default: false)"
	if parsed.authenticated {
		inputs[8].SetValue("true")
	} else {
		inputs[8].SetValue("false")
	}
	inputs[8].CharLimit = 5
	inputs[8].Width = 40

	return inputs
}

// saveBinary saves the binary configuration to the database
func saveBinary(m Model) tea.Cmd {
	return func() tea.Msg {
		if m.ParsedBinary == nil {
			return binarySavedMsg{err: fmt.Errorf("no binary data to save")}
		}

		// Get values from form inputs
		userID := m.FormInputs[0].Value()
		name := m.FormInputs[1].Value()
		provider := m.FormInputs[2].Value()
		path := m.FormInputs[3].Value()
		format := m.FormInputs[4].Value()
		installPath := m.FormInputs[5].Value()
		assetRegex := m.FormInputs[6].Value()
		releaseRegex := m.FormInputs[7].Value()
		authenticatedStr := m.FormInputs[8].Value()

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

		// Parse authenticated boolean
		authenticated := false
		if authenticatedStr != "" {
			authenticatedStr = strings.ToLower(strings.TrimSpace(authenticatedStr))
			authenticated = authenticatedStr == "true" || authenticatedStr == "yes" || authenticatedStr == "1"
		}

		// Check if binary already exists
		existing, err := m.DbService.Binaries.GetByUserID(userID)
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
			Authenticated: authenticated,
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

		err = m.DbService.Binaries.Create(binary)
		if err != nil {
			return binarySavedMsg{err: fmt.Errorf("failed to create binary: %w", err)}
		}

		return binarySavedMsg{binary: binary, err: nil}
	}
}

// updatePlaceholderView handles updates for placeholder views (Downloads, Configuration, Help)
func (m Model) updatePlaceholderView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Configuration view: handle sync action
	if m.CurrentView == views.Configuration {
		switch msg.String() {
		case keySync:
			// Trigger sync
			m.ErrorMessage = ""
			m.SuccessMessage = ""
			m.Loading = true
			return m, syncConfig(m.DbService, m.Config)
		}
	}

	// Check for tab switching
	if view, ok := getTabForKey(msg.String()); ok {
		m.CurrentView = view
		return m, nil
	}

	// Handle tab cycling
	if updatedModel, handled := handleTabCycling(m, msg.String()); handled {
		return updatedModel, nil
	}

	switch msg.String() {
	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	return m, nil
}

// switchVersion switches the active version of a binary
func switchVersion(dbService *repository.Service, binary *database.Binary, installation *database.Installation) tea.Cmd {
	return func() tea.Msg {
		// Handle optional InstallPath
		customInstallPath := ""
		if binary.InstallPath != nil {
			customInstallPath = *binary.InstallPath
		}

		// Update the symlink
		symlinkPath, err := versionSvc.SetActiveVersion(installation.InstalledPath, customInstallPath, binary.Name, binary.Alias)
		if err != nil {
			return versionSwitchedMsg{err: fmt.Errorf("failed to set active version: %w", err)}
		}

		// Update the versions table
		if err := dbService.Versions.Set(binary.ID, installation.ID, symlinkPath); err != nil {
			return versionSwitchedMsg{err: fmt.Errorf("failed to update version record: %w", err)}
		}

		return versionSwitchedMsg{installation: installation, err: nil}
	}
}

// deleteVersion deletes an installation
func deleteVersion(dbService *repository.Service, installation *database.Installation) tea.Cmd {
	return func() tea.Msg {
		// Delete from database
		if err := dbService.Installations.Delete(installation.ID); err != nil {
			return versionDeletedMsg{err: fmt.Errorf("failed to delete version: %w", err)}
		}

		// Optionally delete files (for now we'll leave the files)
		// In a future enhancement, we can add a confirmation dialog

		return versionDeletedMsg{installationID: installation.ID, err: nil}
	}
}

// updateInstallBinary handles updates for the install binary view
func (m Model) updateInstallBinary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Don't process keys if installing
	if m.InstallingInProgress {
		switch msg.String() {
		case keyQuit, keyCtrlC:
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg.String() {
	case keyEsc:
		// Cancel and return to the view we came from
		returnView := m.InstallReturnView
		if returnView == views.ViewState(0) {
			returnView = views.BinariesList
		}
		m.CurrentView = returnView
		m.InstallVersionInput.Reset()
		m.InstallBinaryID = ""
		m.ErrorMessage = ""
		m.SuccessMessage = ""
		return m, nil

	case keyEnter:
		// Start installation
		version := m.InstallVersionInput.Value()
		if version == "" {
			version = "latest"
		}

		m.InstallingInProgress = true
		m.ErrorMessage = ""
		m.SuccessMessage = ""
		return m, installBinary(m.DbService, m.InstallBinaryID, version)

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	// Handle text input
	var cmd tea.Cmd
	m.InstallVersionInput, cmd = m.InstallVersionInput.Update(msg)
	return m, cmd
}

// installBinary installs a binary version
func installBinary(dbService *repository.Service, binaryID string, version string) tea.Cmd {
	return func() tea.Msg {
		// Use the install service to install the binary
		result, err := installSvc.InstallBinary(binaryID, version, dbService)
		if err != nil {
			return binaryInstalledMsg{err: fmt.Errorf("failed to install binary: %w", err)}
		}

		return binaryInstalledMsg{
			binary:       result.Binary,
			installation: result.Installation,
			err:          nil,
		}
	}
}

// updateBinary updates a binary to the latest version
func updateBinary(dbService *repository.Service, binaryID string) tea.Cmd {
	return func() tea.Msg {
		// Get current active version if any
		currentVersion := "none"
		binary, err := dbService.Binaries.GetByUserID(binaryID)
		if err == nil {
			activeVer, err := dbService.Versions.Get(binary.ID)
			if err == nil && activeVer != nil {
				installation, err := dbService.Installations.GetByID(activeVer.InstallationID)
				if err == nil {
					currentVersion = installation.Version
				}
			}
		}

		// Use the install service to update to latest
		result, err := installSvc.UpdateToLatest(binaryID, dbService)
		if err != nil {
			return binaryUpdatedMsg{
				binaryID: binaryID,
				err:      fmt.Errorf("failed to update binary: %w", err),
			}
		}

		return binaryUpdatedMsg{
			binaryID:   binaryID,
			oldVersion: currentVersion,
			newVersion: result.Version,
			err:        nil,
		}
	}
}

// updateAllBinaries updates all binaries to their latest versions
func updateAllBinaries(dbService *repository.Service, binaries []BinaryWithMetadata) tea.Cmd {
	return func() tea.Msg {
		updatedCount := 0
		failedCount := 0

		for _, b := range binaries {
			_, err := installSvc.UpdateToLatest(b.Binary.UserID, dbService)
			if err != nil {
				failedCount++
			} else {
				updatedCount++
			}
		}

		// Return a success message with the summary
		return successMsg{
			message: fmt.Sprintf("Updated %d binaries (%d failed)", updatedCount, failedCount),
		}
	}
}

// removeBinary removes a binary from the database and optionally from disk
func removeBinary(dbService *repository.Service, binaryID string, removeFiles bool) tea.Cmd {
	return func() tea.Msg {
		// Use the binary service to remove the binary
		err := binarySvc.RemoveBinary(binaryID, dbService, removeFiles)
		if err != nil {
			return binaryRemovedMsg{
				binaryID: binaryID,
				err:      fmt.Errorf("failed to remove binary: %w", err),
			}
		}

		return binaryRemovedMsg{
			binaryID: binaryID,
			err:      nil,
		}
	}
}

// checkForUpdates checks if updates are available for a binary
func checkForUpdates(dbService *repository.Service, binaryID string) tea.Cmd {
	return func() tea.Msg {
		// Get the binary configuration
		binaryConfig, err := dbService.Binaries.GetByUserID(binaryID)
		if err != nil {
			return updateCheckMsg{
				binaryID: binaryID,
				err:      fmt.Errorf("binary not found: %w", err),
			}
		}

		if binaryConfig.Provider != "github" {
			return updateCheckMsg{
				binaryID: binaryID,
				err:      fmt.Errorf("only github provider is currently supported"),
			}
		}

		// Import github provider
		// Fetch latest release
		release, _, err := github.FetchReleaseAsset(binaryConfig, "latest")
		if err != nil {
			return updateCheckMsg{
				binaryID: binaryID,
				err:      fmt.Errorf("failed to fetch latest release: %w", err),
			}
		}

		latestVersion := release.TagName

		// Get current active version
		currentVersion := "none"
		activeVersion, err := dbService.Versions.Get(binaryConfig.ID)
		if err == nil && activeVersion != nil {
			installation, err := dbService.Installations.GetByID(activeVersion.InstallationID)
			if err == nil {
				currentVersion = installation.Version
			}
		}

		// Check if latest version is already installed (even if not active)
		_, err = dbService.Installations.Get(binaryConfig.ID, latestVersion)
		isLatestInstalled := err == nil

		// Determine if update is needed: latest not installed, or current is not latest
		hasUpdate := !isLatestInstalled && currentVersion != latestVersion && currentVersion != "none"

		return updateCheckMsg{
			binaryID:        binaryID,
			currentVersion:  currentVersion,
			latestVersion:   latestVersion,
			hasUpdate:       hasUpdate,
			latestInstalled: isLatestInstalled && currentVersion != latestVersion,
			err:             nil,
		}
	}
}

// updateImportBinary handles updates for the import binary view
func (m Model) updateImportBinary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		// Cancel and return to binaries list
		m.CurrentView = views.BinariesList
		m.ImportPathInput.Reset()
		m.ImportNameInput.Reset()
		m.ImportFocusIdx = 0
		m.ErrorMessage = ""
		m.SuccessMessage = ""
		return m, nil

	case keyTab:
		// Move to next field
		if m.ImportFocusIdx == 0 {
			m.ImportPathInput.Blur()
			m.ImportNameInput.Focus()
			m.ImportFocusIdx = 1
		} else {
			m.ImportNameInput.Blur()
			m.ImportPathInput.Focus()
			m.ImportFocusIdx = 0
		}
		return m, nil

	case keyEnter:
		// Attempt import
		path := m.ImportPathInput.Value()
		name := m.ImportNameInput.Value()

		if path == "" {
			m.ErrorMessage = "Binary path is required"
			m.SuccessMessage = ""
			return m, nil
		}
		if name == "" {
			m.ErrorMessage = "Binary name is required"
			m.SuccessMessage = ""
			return m, nil
		}

		// Show message that import is not yet fully implemented
		m.ErrorMessage = ""
		m.SuccessMessage = "Import functionality is pending service layer implementation"
		return m, nil

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	// Handle text input for focused field
	var cmd tea.Cmd
	if m.ImportFocusIdx == 0 {
		m.ImportPathInput, cmd = m.ImportPathInput.Update(msg)
	} else {
		m.ImportNameInput, cmd = m.ImportNameInput.Update(msg)
	}
	return m, cmd
}

// syncConfig syncs the configuration file to the database
func syncConfig(dbService *repository.Service, cfg *configPkg.Config) tea.Cmd {
	return func() tea.Msg {
		err := configPkg.SyncToDatabase(*cfg, dbService)
		if err != nil {
			return configSyncedMsg{err: fmt.Errorf("failed to sync config: %w", err)}
		}

		return configSyncedMsg{err: nil}
	}
}
