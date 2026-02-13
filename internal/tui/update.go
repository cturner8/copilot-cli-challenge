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
			// Return to the view we came from (binaries list or versions)
			returnView := m.installReturnView
			if returnView == viewState(0) {
				returnView = viewBinariesList // Default to binaries list
			}
			m.currentView = returnView
			m.installVersionInput.Reset()
			m.installBinaryID = ""
			m.errorMessage = ""
			m.successMessage = fmt.Sprintf("Successfully installed %s version %s", msg.binary.Name, msg.installation.Version)

			// Reload appropriate data based on return view
			m.loading = true
			if returnView == viewVersions && m.selectedBinary != nil {
				return m, loadVersions(m.dbService, m.selectedBinary.ID)
			}
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
			// Reload appropriate data based on current view
			m.loading = true
			if m.currentView == viewVersions && m.selectedBinary != nil {
				return m, loadVersions(m.dbService, m.selectedBinary.ID)
			}
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

	case binaryImportedMsg:
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.successMessage = ""
		} else {
			m.errorMessage = ""
			m.successMessage = fmt.Sprintf("Binary %s imported successfully", msg.binary.UserID)
			// Reset form and return to binaries list
			m.importPathInput.Reset()
			m.importNameInput.Reset()
			m.importURLInput.Reset()
			m.importVersionInput.Reset()
			m.importFocusIdx = 0
			m.currentView = viewBinariesList
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
			} else if msg.latestInstalled {
				m.successMessage = fmt.Sprintf("✓ %s: latest version (%s) installed but not active (current: %s)", msg.binaryID, msg.latestVersion, msg.currentVersion)
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

	case githubRepoInfoMsg:
		m.githubLoading = false
		if msg.err != nil {
			m.githubError = msg.err.Error()
		} else {
			m.githubRepoInfo = msg.info
		}
		return m, nil

	case githubAvailableVersionsMsg:
		m.githubLoading = false
		if msg.err != nil {
			m.githubError = msg.err.Error()
		} else {
			m.githubAvailableVers = msg.versions
		}
		return m, nil

	case githubReleaseNotesMsg:
		m.githubLoading = false
		if msg.err != nil {
			m.githubError = msg.err.Error()
		} else {
			m.githubReleaseInfo = msg.release
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
		case viewReleaseNotes, viewAvailableVersions, viewRepositoryInfo:
			return m.updateGitHubView(msg)
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
			// Check if bulk mode and multiple selections
			if m.bulkSelectMode && len(m.selectedBinaries) > 0 {
				binariesToShow := getDisplayBinaries(m.binaries, m.activeFilters, m.searchQuery, m.sortMode, m.sortAscending)
				var cmds []tea.Cmd
				for idx := range m.selectedBinaries {
					if idx < len(binariesToShow) {
						binaryID := binariesToShow[idx].Binary.UserID
						cmds = append(cmds, removeBinary(m.dbService, binaryID, false))
					}
				}
				m.confirmingRemove = false
				m.removeBinaryID = ""
				m.selectedBinaries = make(map[int]bool)
				m.bulkSelectMode = false
				if len(cmds) > 0 {
					return m, tea.Batch(cmds...)
				}
				return m, nil
			} else {
				// Single remove
				binaryID := m.removeBinaryID
				m.confirmingRemove = false
				m.removeBinaryID = ""
				return m, removeBinary(m.dbService, binaryID, false)
			}
		case "Y":
			// Remove with files
			// Check if bulk mode and multiple selections
			if m.bulkSelectMode && len(m.selectedBinaries) > 0 {
				binariesToShow := getDisplayBinaries(m.binaries, m.activeFilters, m.searchQuery, m.sortMode, m.sortAscending)
				var cmds []tea.Cmd
				for idx := range m.selectedBinaries {
					if idx < len(binariesToShow) {
						binaryID := binariesToShow[idx].Binary.UserID
						cmds = append(cmds, removeBinary(m.dbService, binaryID, true))
					}
				}
				m.confirmingRemove = false
				m.removeBinaryID = ""
				m.selectedBinaries = make(map[int]bool)
				m.bulkSelectMode = false
				if len(cmds) > 0 {
					return m, tea.Batch(cmds...)
				}
				return m, nil
			} else {
				// Single remove
				binaryID := m.removeBinaryID
				m.confirmingRemove = false
				m.removeBinaryID = ""
				return m, removeBinary(m.dbService, binaryID, true)
			}
		case "n", keyEsc:
			// Cancel
			m.confirmingRemove = false
			m.removeBinaryID = ""
			return m, nil
		}
		return m, nil
	}

	// Handle filter panel if open
	if m.filterPanelOpen {
		switch msg.String() {
		case "1":
			// Toggle provider filter (github only for now)
			if _, ok := m.activeFilters["provider"]; ok {
				delete(m.activeFilters, "provider")
			} else {
				m.activeFilters["provider"] = "github"
			}
			// Clear selections when filters change
			if m.bulkSelectMode {
				m.selectedBinaries = make(map[int]bool)
			}
			return m, nil
		case "2":
			// Cycle format filter (.tar.gz -> .zip -> none)
			if format, ok := m.activeFilters["format"]; ok {
				if format == ".tar.gz" {
					m.activeFilters["format"] = ".zip"
				} else {
					delete(m.activeFilters, "format")
				}
			} else {
				m.activeFilters["format"] = ".tar.gz"
			}
			// Clear selections when filters change
			if m.bulkSelectMode {
				m.selectedBinaries = make(map[int]bool)
			}
			return m, nil
		case "3":
			// Cycle status filter (installed -> not-installed -> none)
			if status, ok := m.activeFilters["status"]; ok {
				if status == "installed" {
					m.activeFilters["status"] = "not-installed"
				} else {
					delete(m.activeFilters, "status")
				}
			} else {
				m.activeFilters["status"] = "installed"
			}
			// Clear selections when filters change
			if m.bulkSelectMode {
				m.selectedBinaries = make(map[int]bool)
			}
			return m, nil
		case "c":
			// Clear all filters
			m.activeFilters = make(map[string]string)
			// Clear selections when filters change
			if m.bulkSelectMode {
				m.selectedBinaries = make(map[int]bool)
			}
			m.successMessage = "All filters cleared"
			return m, nil
		case keyEsc:
			// Close filter panel
			m.filterPanelOpen = false
			return m, nil
		}
		return m, nil
	}

	// Handle search mode - only process text input when actively typing
	if m.searchMode {
		switch msg.String() {
		case keyEsc:
			// Exit search mode
			m.searchMode = false
			m.searchTextInput.Reset()
			m.searchTextInput.Blur()
			m.searchQuery = ""
			m.filteredBinaries = []BinaryWithMetadata{}
			m.selectedIndex = 0
			// Clear selections when search changes
			if m.bulkSelectMode {
				m.selectedBinaries = make(map[int]bool)
			}
			return m, nil
		case keyEnter:
			// Apply search and exit input mode (but keep filtered view)
			m.searchQuery = m.searchTextInput.Value()
			m.searchTextInput.Blur()
			m.searchMode = false // Exit search mode to allow normal navigation
			m.filteredBinaries = filterBinaries(m.binaries, m.searchQuery)
			m.selectedIndex = 0
			// Clear selections when search changes
			if m.bulkSelectMode {
				m.selectedBinaries = make(map[int]bool)
			}
			return m, nil
		default:
			// Update search input
			var cmd tea.Cmd
			m.searchTextInput, cmd = m.searchTextInput.Update(msg)
			return m, cmd
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

	// Handle bulk selection toggle with space key BEFORE the main switch
	// This prevents space from falling through to other handlers
	// Check both msg.String() and the actual rune to be absolutely sure
	keyStr := msg.String()
	if keyStr == keySpace || keyStr == " " {
		if m.bulkSelectMode {
			binariesToShow := getDisplayBinaries(m.binaries, m.activeFilters, m.searchQuery, m.sortMode, m.sortAscending)
			if len(binariesToShow) > 0 && m.selectedIndex < len(binariesToShow) {
				// Toggle selection
				if m.selectedBinaries[m.selectedIndex] {
					delete(m.selectedBinaries, m.selectedIndex)
				} else {
					m.selectedBinaries[m.selectedIndex] = true
				}
			}
			return m, nil
		}
		// If not in bulk mode, space does nothing (prevents accidental actions)
		return m, nil
	}

	switch msg.String() {
	case keyUp:
		// Navigate display list
		binariesToShow := getDisplayBinaries(m.binaries, m.activeFilters, m.searchQuery, m.sortMode, m.sortAscending)
		if m.selectedIndex > 0 && len(binariesToShow) > 0 {
			m.selectedIndex--
		}

	case keyDown:
		// Navigate display list
		binariesToShow := getDisplayBinaries(m.binaries, m.activeFilters, m.searchQuery, m.sortMode, m.sortAscending)
		if len(binariesToShow) > 0 && m.selectedIndex < len(binariesToShow)-1 {
			m.selectedIndex++
		}

	case keyEnter:
		// Transition to versions view using display list
		binariesToShow := getDisplayBinaries(m.binaries, m.activeFilters, m.searchQuery, m.sortMode, m.sortAscending)
		if len(binariesToShow) > 0 && m.selectedIndex < len(binariesToShow) {
			m.currentView = viewVersions
			m.selectedBinary = binariesToShow[m.selectedIndex].Binary
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
		binariesToShow := getDisplayBinaries(m.binaries, m.activeFilters, m.searchQuery, m.sortMode, m.sortAscending)
		if len(binariesToShow) > 0 && m.selectedIndex < len(binariesToShow) {
			m.currentView = viewInstallBinary
			m.installBinaryID = binariesToShow[m.selectedIndex].Binary.UserID
			m.installReturnView = viewBinariesList
			m.installVersionInput.Focus()
			m.errorMessage = ""
			m.successMessage = ""
		}

	case keyUpdate:
		// Update selected binary(ies) to latest version
		binariesToShow := getDisplayBinaries(m.binaries, m.activeFilters, m.searchQuery, m.sortMode, m.sortAscending)

		// If in bulk mode and items are selected, update all selected
		if m.bulkSelectMode && len(m.selectedBinaries) > 0 {
			var selectedBinariesList []BinaryWithMetadata
			for idx := range m.selectedBinaries {
				if idx < len(binariesToShow) {
					selectedBinariesList = append(selectedBinariesList, binariesToShow[idx])
				}
			}
			if len(selectedBinariesList) > 0 {
				m.errorMessage = ""
				m.successMessage = ""
				m.loading = true
				return m, updateAllBinaries(m.dbService, selectedBinariesList)
			}
		} else if len(binariesToShow) > 0 && m.selectedIndex < len(binariesToShow) {
			// Single update
			selectedBinary := binariesToShow[m.selectedIndex]
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
		// Show remove confirmation for selected binary(ies)
		binariesToShow := getDisplayBinaries(m.binaries, m.activeFilters, m.searchQuery, m.sortMode, m.sortAscending)

		// If in bulk mode and items are selected, prepare to remove all selected
		if m.bulkSelectMode && len(m.selectedBinaries) > 0 {
			// For bulk remove, track the count separately
			m.confirmingRemove = true
			m.bulkRemoveCount = len(m.selectedBinaries)
			m.removeBinaryID = "" // Clear single binary ID for bulk operations
			m.errorMessage = ""
			m.successMessage = ""
		} else if len(binariesToShow) > 0 && m.selectedIndex < len(binariesToShow) {
			// Single remove
			m.confirmingRemove = true
			m.removeBinaryID = binariesToShow[m.selectedIndex].Binary.UserID
			m.bulkRemoveCount = 0 // Clear bulk count for single operations
			m.errorMessage = ""
			m.successMessage = ""
		}

	case keyCheck:
		// Check for updates for selected binary
		binariesToShow := getDisplayBinaries(m.binaries, m.activeFilters, m.searchQuery, m.sortMode, m.sortAscending)
		if len(binariesToShow) > 0 && m.selectedIndex < len(binariesToShow) {
			selectedBinary := binariesToShow[m.selectedIndex]
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

	case keySearch:
		// Enter search mode
		m.searchMode = true
		// Preserve current search value if one exists
		if m.searchQuery != "" {
			m.searchTextInput.SetValue(m.searchQuery)
		} else {
			m.searchTextInput.Reset()
		}
		m.searchTextInput.Focus()
		m.errorMessage = ""
		m.successMessage = ""

	case keyFilter:
		// Toggle filter panel
		m.filterPanelOpen = !m.filterPanelOpen

	case keyNextSort:
		// Cycle through sort modes
		switch m.sortMode {
		case "name":
			m.sortMode = "provider"
		case "provider":
			m.sortMode = "count"
		case "count":
			m.sortMode = "updated"
		case "updated":
			m.sortMode = "name"
		default:
			m.sortMode = "name"
		}
		m.successMessage = fmt.Sprintf("Sort by: %s", m.sortMode)

	case keySortOrder:
		// Toggle sort order
		m.sortAscending = !m.sortAscending
		direction := "ascending"
		if !m.sortAscending {
			direction = "descending"
		}
		m.successMessage = fmt.Sprintf("Sort order: %s", direction)

	case keyBulkMode:
		// Toggle bulk selection mode
		m.bulkSelectMode = !m.bulkSelectMode
		if !m.bulkSelectMode {
			// Clear selections when exiting bulk mode
			m.selectedBinaries = make(map[int]bool)
			m.successMessage = "Bulk mode: OFF"
		} else {
			m.successMessage = "Bulk mode: ON (use Space to select, U to update all, R to remove all)"
		}

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

	case keyInstall:
		// Install a new version
		if m.selectedBinary != nil {
			m.currentView = viewInstallBinary
			m.installBinaryID = m.selectedBinary.UserID
			m.installReturnView = viewVersions
			m.installVersionInput.Focus()
			m.errorMessage = ""
			m.successMessage = ""
		}

	case keyUpdate:
		// Update to latest version
		if m.selectedBinary != nil {
			m.errorMessage = ""
			m.successMessage = ""
			return m, updateBinary(m.dbService, m.selectedBinary.UserID)
		}

	case keyCheck:
		// Check for updates
		if m.selectedBinary != nil {
			m.errorMessage = ""
			m.successMessage = ""
			m.loading = true
			return m, checkForUpdates(m.dbService, m.selectedBinary.UserID)
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

	case keyReleaseNotes:
		// View release notes for selected version
		if m.selectedBinary != nil && m.selectedBinary.Provider == "github" {
			// Get the active version or selected version
			version := "latest"
			if len(m.installations) > 0 && m.selectedVersionIdx < len(m.installations) {
				version = m.installations[m.selectedVersionIdx].Version
			}
			m.currentView = viewReleaseNotes
			m.githubLoading = true
			m.githubError = ""
			return m, fetchReleaseNotes(m.selectedBinary, version)
		}

	case keyRepoInfo:
		// View GitHub repository information
		if m.selectedBinary != nil && m.selectedBinary.Provider == "github" {
			m.currentView = viewRepositoryInfo
			m.githubLoading = true
			m.githubError = ""
			return m, fetchRepositoryInfo(m.selectedBinary)
		}

	case keyAvailVersions:
		// View available versions from GitHub
		if m.selectedBinary != nil && m.selectedBinary.Provider == "github" {
			m.currentView = viewAvailableVersions
			m.githubLoading = true
			m.githubError = ""
			return m, fetchAvailableVersions(m.selectedBinary)
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
		authenticatedStr := m.formInputs[8].Value()

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
		// Cancel and return to the view we came from
		returnView := m.installReturnView
		if returnView == viewState(0) {
			returnView = viewBinariesList
		}
		m.currentView = returnView
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

// importBinary imports an existing binary from the file system
func importBinary(dbService *repository.Service, path string, name string) tea.Cmd {
	return importBinaryWithOptions(dbService, path, name, "", "")
}

// importBinaryWithOptions imports an existing binary with optional URL and version
func importBinaryWithOptions(dbService *repository.Service, path string, name string, url string, version string) tea.Cmd {
	return func() tea.Msg {
		// Use the binary service to import the binary
		binary, err := binarySvc.ImportBinaryWithOptions(path, name, url, version, false, false, dbService)
		if err != nil {
			return binaryImportedMsg{
				err: fmt.Errorf("failed to import binary: %w", err),
			}
		}

		return binaryImportedMsg{
			binary: binary,
			err:    nil,
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
func (m model) updateImportBinary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		// Cancel and return to binaries list
		m.currentView = viewBinariesList
		m.importPathInput.Reset()
		m.importNameInput.Reset()
		m.importURLInput.Reset()
		m.importVersionInput.Reset()
		m.importFocusIdx = 0
		m.errorMessage = ""
		m.successMessage = ""
		return m, nil

	case keyTab:
		// Move to next field (0 -> 1 -> 2 -> 3 -> 0)
		m.importPathInput.Blur()
		m.importNameInput.Blur()
		m.importURLInput.Blur()
		m.importVersionInput.Blur()

		m.importFocusIdx = (m.importFocusIdx + 1) % 4

		switch m.importFocusIdx {
		case 0:
			m.importPathInput.Focus()
		case 1:
			m.importNameInput.Focus()
		case 2:
			m.importURLInput.Focus()
		case 3:
			m.importVersionInput.Focus()
		}
		return m, nil

	case keyShiftTab:
		// Move to previous field (3 -> 2 -> 1 -> 0 -> 3)
		m.importPathInput.Blur()
		m.importNameInput.Blur()
		m.importURLInput.Blur()
		m.importVersionInput.Blur()

		m.importFocusIdx = (m.importFocusIdx - 1 + 4) % 4

		switch m.importFocusIdx {
		case 0:
			m.importPathInput.Focus()
		case 1:
			m.importNameInput.Focus()
		case 2:
			m.importURLInput.Focus()
		case 3:
			m.importVersionInput.Focus()
		}
		return m, nil

	case keyEnter:
		// Attempt import
		path := m.importPathInput.Value()
		name := m.importNameInput.Value()
		url := m.importURLInput.Value()
		version := m.importVersionInput.Value()

		if path == "" {
			m.errorMessage = "Binary path is required"
			m.successMessage = ""
			return m, nil
		}
		if name == "" && url == "" {
			m.errorMessage = "Either binary name or GitHub URL is required"
			m.successMessage = ""
			return m, nil
		}

		// Clear messages and trigger import
		m.errorMessage = ""
		m.successMessage = ""
		return m, importBinaryWithOptions(m.dbService, path, name, url, version)

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	// Handle text input for focused field
	var cmd tea.Cmd
	switch m.importFocusIdx {
	case 0:
		m.importPathInput, cmd = m.importPathInput.Update(msg)
	case 1:
		m.importNameInput, cmd = m.importNameInput.Update(msg)
	case 2:
		m.importURLInput, cmd = m.importURLInput.Update(msg)
	case 3:
		m.importVersionInput, cmd = m.importVersionInput.Update(msg)
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

// updateGitHubView handles updates for GitHub-related views
func (m model) updateGitHubView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		// Return to versions view
		m.currentView = viewVersions
		m.githubReleaseInfo = nil
		m.githubAvailableVers = nil
		m.githubRepoInfo = nil
		m.githubError = ""
		return m, nil

	case keyQuit, keyCtrlC:
		return m, tea.Quit
	}

	return m, nil
}

// Message types for GitHub data fetching
type githubRepoInfoMsg struct {
	info *githubRepositoryInfo
	err  error
}

type githubAvailableVersionsMsg struct {
	versions []githubReleaseInfo
	err      error
}

type githubReleaseNotesMsg struct {
	release *githubReleaseInfo
	err     error
}

// fetchRepositoryInfo fetches repository information from GitHub
func fetchRepositoryInfo(binary *database.Binary) tea.Cmd {
	return func() tea.Msg {
		repoInfo, err := github.GetRepositoryInfo(binary)
		if err != nil {
			return githubRepoInfoMsg{err: err}
		}

		return githubRepoInfoMsg{
			info: &githubRepositoryInfo{
				Name:        repoInfo.Name,
				FullName:    repoInfo.FullName,
				Description: repoInfo.Description,
				Stars:       repoInfo.StargazersCount,
				Forks:       repoInfo.ForksCount,
				HTMLURL:     repoInfo.HTMLURL,
			},
		}
	}
}

// fetchAvailableVersions fetches available versions from GitHub
func fetchAvailableVersions(binary *database.Binary) tea.Cmd {
	return func() tea.Msg {
		releases, err := github.ListAvailableVersions(binary, 20)
		if err != nil {
			return githubAvailableVersionsMsg{err: err}
		}

		var versions []githubReleaseInfo
		for _, release := range releases {
			versions = append(versions, githubReleaseInfo{
				Name:        release.Name,
				TagName:     release.TagName,
				Body:        release.Body,
				Prerelease:  release.Prerelease,
				PublishedAt: release.PublishedAt.Format("2006-01-02 15:04"),
				HTMLURL:     release.HTMLURL,
			})
		}

		return githubAvailableVersionsMsg{versions: versions}
	}
}

// fetchReleaseNotes fetches release notes from GitHub for a specific version
func fetchReleaseNotes(binary *database.Binary, version string) tea.Cmd {
	return func() tea.Msg {
		releaseInfo, err := github.FetchReleaseNotes(binary, version)
		if err != nil {
			return githubReleaseNotesMsg{err: err}
		}

		return githubReleaseNotesMsg{
			release: &githubReleaseInfo{
				Name:        releaseInfo.Name,
				TagName:     releaseInfo.TagName,
				Body:        releaseInfo.Body,
				Prerelease:  releaseInfo.Prerelease,
				PublishedAt: releaseInfo.PublishedAt.Format("2006-01-02 15:04"),
				HTMLURL:     releaseInfo.HTMLURL,
			},
		}
	}
}
