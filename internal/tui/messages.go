package tui

import "cturner8/binmate/internal/database"

// Custom messages for Bubble Tea
type (
	// binariesLoadedMsg is sent when binaries data is loaded
	binariesLoadedMsg struct {
		binaries []BinaryWithMetadata
		err      error
	}

	// versionsLoadedMsg is sent when versions data is loaded
	versionsLoadedMsg struct {
		installations []*database.Installation
		err           error
	}

	// binarySavedMsg is sent when a new binary is saved
	binarySavedMsg struct {
		binary *database.Binary
		err    error
	}

	// versionSwitchedMsg is sent when active version is switched
	versionSwitchedMsg struct {
		installation *database.Installation
		err          error
	}

	// versionDeletedMsg is sent when a version is deleted
	versionDeletedMsg struct {
		installationID int64
		err            error
	}

	// binaryInstalledMsg is sent when a binary version is installed
	binaryInstalledMsg struct {
		binary       *database.Binary
		installation *database.Installation
		err          error
	}

	// binaryUpdatedMsg is sent when a binary is updated
	binaryUpdatedMsg struct {
		binaryID   string
		oldVersion string
		newVersion string
		err        error
	}

	// binaryRemovedMsg is sent when a binary is removed
	binaryRemovedMsg struct {
		binaryID string
		err      error
	}

	// binaryImportedMsg is sent when a binary is imported
	binaryImportedMsg struct {
		binary *database.Binary
		err    error
	}

	// updateCheckMsg is sent when update check is complete
	updateCheckMsg struct {
		binaryID        string
		currentVersion  string
		latestVersion   string
		hasUpdate       bool
		latestInstalled bool // true if latest version is installed but not active
		err             error
	}

	// configSyncedMsg is sent when config sync is complete
	configSyncedMsg struct {
		err error
	}

	// errorMsg represents an error state
	errorMsg struct {
		err error
	}

	// successMsg represents a success notification
	successMsg struct {
		message string
	}
)
