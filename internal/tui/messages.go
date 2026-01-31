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

	// errorMsg represents an error state
	errorMsg struct {
		err error
	}
)
