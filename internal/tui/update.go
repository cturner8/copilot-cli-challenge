package tui

import (
	"fmt"

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
			m.successMessage = fmt.Sprintf("Binary %s added successfully", msg.binary.Name)
			m.loading = true
			return m, loadBinaries(m.dbService)
		}
		return m, nil

	case binaryInstalledMsg:
		m.installingInProgress = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.successMessage = ""
		} else {
			// Return to binaries list and reload
			m.currentView = viewBinariesList
			m.errorMessage = ""
			m.successMessage = fmt.Sprintf("Successfully installed %s version %s", msg.binary.Name, msg.installation.Version)
			m.loading = true
			return m, loadBinaries(m.dbService)
		}
		return m, nil

	case binaryUpdatedMsg:
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.successMessage = ""
		} else {
			m.errorMessage = ""
			if msg.oldVersion == msg.newVersion {
				m.successMessage = fmt.Sprintf("%s is already up to date (%s)", msg.binaryID, msg.newVersion)
			} else {
				m.successMessage = fmt.Sprintf("Updated %s from %s to %s", msg.binaryID, msg.oldVersion, msg.newVersion)
			}
			// Reload binaries list
			m.loading = true
			return m, loadBinaries(m.dbService)
		}
		return m, nil

	case binaryRemovedMsg:
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.successMessage = ""
		} else {
			m.errorMessage = ""
			m.successMessage = fmt.Sprintf("Binary %s removed successfully", msg.binaryID)
			// Reload binaries list
			m.loading = true
			return m, loadBinaries(m.dbService)
		}
		return m, nil

	case updateCheckMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.successMessage = ""
		} else {
			m.errorMessage = ""
			if msg.hasUpdate {
				m.successMessage = fmt.Sprintf("⬆ Update available for %s: %s → %s", msg.binaryID, msg.currentVersion, msg.latestVersion)
			} else {
				m.successMessage = fmt.Sprintf("✓ %s is up to date (%s)", msg.binaryID, msg.currentVersion)
			}
		}
		return m, nil

	case configSyncedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.successMessage = ""
		} else {
			m.errorMessage = ""
			m.successMessage = "Configuration synced to database successfully"
			// Reload binaries list
			return m, loadBinaries(m.dbService)
		}
		return m, nil

	case versionSwitchedMsg:
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.successMessage = ""
		} else {
			m.errorMessage = ""
			m.successMessage = fmt.Sprintf("Switched to version %s", msg.installation.Version)
			// Reload versions to update active indicator
			m.loading = true
			return m, loadVersions(m.dbService, m.selectedBinary.ID)
		}
		return m, nil

	case versionDeletedMsg:
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.successMessage = ""
		} else {
			m.errorMessage = ""
			m.successMessage = "Version deleted successfully"
			// Reload versions list
			m.loading = true
			return m, loadVersions(m.dbService, m.selectedBinary.ID)
		}
		return m, nil

	case successMsg:
		m.successMessage = msg.message
		m.errorMessage = ""
		return m, nil

	case errorMsg:
		m.errorMessage = msg.err.Error()
		m.successMessage = ""
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
		case viewInstallBinary:
			return m.updateInstallBinary(msg)
		case viewImportBinary:
			return m.updateImportBinary(msg)
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
	// Handle remove confirmation if active
	if m.confirmingRemove {
		switch msg.String() {
		case "y":
			// Remove without files
			binaryID := m.removeBinaryID
			m.confirmingRemove = false
			m.removeBinaryID = ""
			return m, removeBinary(m.dbService, binaryID, false)
		case "Y":
			// Remove with files
			binaryID := m.removeBinaryID
			m.confirmingRemove = false
			m.removeBinaryID = ""
			return m, removeBinary(m.dbService, binaryID, true)
		case "n", keyEsc:
			// Cancel
			m.confirmingRemove = false
			m.removeBinaryID = ""
			return m, nil
		}
		return m, nil
	}

	// Check for tab switching
	if view, ok := getTabForKey(msg.String()); ok {
		m.currentView = view
		return m, nil
	}

	// Handle tab cycling
	if updatedModel, handled := handleTabCycling(m, msg.String()); handled {
		return updatedModel, nil
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
		m.successMessage = ""

	case keyInstall:
		// Transition to install binary view
		if len(m.binaries) > 0 && m.selectedIndex < len(m.binaries) {
			m.currentView = viewInstallBinary
			m.installBinaryID = m.binaries[m.selectedIndex].Binary.UserID
			m.installVersionInput.Focus()
			m.errorMessage = ""
			m.successMessage = ""
		}

	case keyUpdate:
		// Update selected binary to latest version
		if len(m.binaries) > 0 && m.selectedIndex < len(m.binaries) {
			selectedBinary := m.binaries[m.selectedIndex]
			m.errorMessage = ""
			m.successMessage = ""
			return m, updateBinary(m.dbService, selectedBinary.Binary.UserID)
		}

	case keyUpdateAll:
		// Update all binaries to latest version
		if len(m.binaries) > 0 {
			m.errorMessage = ""
			m.successMessage = ""
			m.loading = true
			return m, updateAllBinaries(m.dbService, m.binaries)
		}

	case keyRemove:
		// Show remove confirmation for selected binary
		if len(m.binaries) > 0 && m.selectedIndex < len(m.binaries) {
			m.confirmingRemove = true
			m.removeBinaryID = m.binaries[m.selectedIndex].Binary.UserID
			m.errorMessage = ""
			m.successMessage = ""
		}

	case keyCheck:
		// Check for updates for selected binary
		if len(m.binaries) > 0 && m.selectedIndex < len(m.binaries) {
			selectedBinary := m.binaries[m.selectedIndex]
			m.errorMessage = ""
			m.successMessage = ""
			m.loading = true
			return m, checkForUpdates(m.dbService, selectedBinary.Binary.UserID)
		}

	case keyImport:
		// Transition to import binary view
		m.currentView = viewImportBinary
		m.importPathInput.Reset()
		m.importNameInput.Reset()
		m.importFocusIdx = 0
		m.importPathInput.Focus()
		m.errorMessage = ""
		m.successMessage = ""

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	return m, nil
}

// updateVersions handles updates for the versions view
func (m model) updateVersions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp:
		if m.selectedVersionIdx > 0 {
			m.selectedVersionIdx--
		}

	case keyDown:
		if m.selectedVersionIdx < len(m.installations)-1 {
			m.selectedVersionIdx++
		}

	case keySwitch, keyEnter:
		// Switch to selected version
		if len(m.installations) > 0 && m.selectedVersionIdx < len(m.installations) {
			selectedInstallation := m.installations[m.selectedVersionIdx]
			return m, switchVersion(m.dbService, m.selectedBinary, selectedInstallation)
		}

	case keyDelete, keyDelete2:
		// Delete selected version
		if len(m.installations) > 0 && m.selectedVersionIdx < len(m.installations) {
			selectedInstallation := m.installations[m.selectedVersionIdx]

			// Check if this is the active version
			activeVersion, _ := getActiveVersion(m.dbService, m.selectedBinary.ID)
			if activeVersion != nil && activeVersion.ID == selectedInstallation.ID {
				m.errorMessage = "Cannot delete active version. Switch to another version first."
				m.successMessage = ""
				return m, nil
			}

			return m, deleteVersion(m.dbService, selectedInstallation)
		}

	case keyEsc:
		// Return to binaries list
		m.currentView = viewBinariesList
		m.selectedBinary = nil
		m.installations = nil
		m.selectedVersionIdx = 0
		m.errorMessage = ""
		m.successMessage = ""

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
			userID:    urlparser.GenerateBinaryID(parsed.AssetName),
			name:      urlparser.GenerateBinaryName(parsed.AssetName),
			provider:  "github",
			path:      fmt.Sprintf("%s/%s", parsed.Owner, parsed.Repo),
			format:    parsed.Format,
			version:   parsed.Version,
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
	// Configuration view: handle sync action
	if m.currentView == viewConfiguration {
		switch msg.String() {
		case keySync:
			// Trigger sync
			m.errorMessage = ""
			m.successMessage = ""
			m.loading = true
			return m, syncConfig(m.dbService, m.config)
		}
	}

	// Check for tab switching
	if view, ok := getTabForKey(msg.String()); ok {
		m.currentView = view
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
func (m model) updateInstallBinary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Don't process keys if installing
	if m.installingInProgress {
		switch msg.String() {
		case keyQuit, keyCtrlC:
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg.String() {
	case keyEsc:
		// Cancel and return to binaries list
		m.currentView = viewBinariesList
		m.installVersionInput.Reset()
		m.installBinaryID = ""
		m.errorMessage = ""
		m.successMessage = ""
		return m, nil

	case keyEnter:
		// Start installation
		version := m.installVersionInput.Value()
		if version == "" {
			version = "latest"
		}

		m.installingInProgress = true
		m.errorMessage = ""
		m.successMessage = ""
		return m, installBinary(m.dbService, m.installBinaryID, version)

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	// Handle text input
	var cmd tea.Cmd
	m.installVersionInput, cmd = m.installVersionInput.Update(msg)
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

		// Get current active version
		currentVersion := "none"
		activeVersion, err := dbService.Versions.Get(binaryConfig.ID)
		if err == nil && activeVersion != nil {
			installation, err := dbService.Installations.GetByID(activeVersion.InstallationID)
			if err == nil {
				currentVersion = installation.Version
			}
		}

		latestVersion := release.TagName
		hasUpdate := currentVersion != latestVersion && currentVersion != "none"

		return updateCheckMsg{
			binaryID:       binaryID,
			currentVersion: currentVersion,
			latestVersion:  latestVersion,
			hasUpdate:      hasUpdate,
			err:            nil,
		}
	}
}

// updateImportBinary handles updates for the import binary view
func (m model) updateImportBinary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		// Cancel and return to binaries list
		m.currentView = viewBinariesList
		m.importPathInput.Reset()
		m.importNameInput.Reset()
		m.importFocusIdx = 0
		m.errorMessage = ""
		m.successMessage = ""
		return m, nil

	case keyTab:
		// Move to next field
		if m.importFocusIdx == 0 {
			m.importPathInput.Blur()
			m.importNameInput.Focus()
			m.importFocusIdx = 1
		} else {
			m.importNameInput.Blur()
			m.importPathInput.Focus()
			m.importFocusIdx = 0
		}
		return m, nil

	case keyEnter:
		// Attempt import
		path := m.importPathInput.Value()
		name := m.importNameInput.Value()

		if path == "" {
			m.errorMessage = "Binary path is required"
			m.successMessage = ""
			return m, nil
		}
		if name == "" {
			m.errorMessage = "Binary name is required"
			m.successMessage = ""
			return m, nil
		}

		// Show message that import is not yet fully implemented
		m.errorMessage = ""
		m.successMessage = "Import functionality is pending service layer implementation"
		return m, nil

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	// Handle text input for focused field
	var cmd tea.Cmd
	if m.importFocusIdx == 0 {
		m.importPathInput, cmd = m.importPathInput.Update(msg)
	} else {
		m.importNameInput, cmd = m.importNameInput.Update(msg)
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
