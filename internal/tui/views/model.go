package views

import (
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"

	"github.com/charmbracelet/bubbles/textinput"
)

// Model represents the TUI model data needed for rendering views
// This mirrors the main TUI model but is defined here to avoid circular imports
type Model struct {
	// Services
	DbService interface{} // Using interface{} to avoid importing repository
	Config    *config.Config

	// View state
	CurrentView ViewState

	// Window dimensions
	Width  int
	Height int

	// Binaries list view state
	Binaries      []BinaryWithMetadata
	SelectedIndex int
	Loading       bool

	// Versions view state
	SelectedBinary       *database.Binary
	Installations        []*database.Installation
	SelectedVersionIdx   int
	ActiveInstallationID int64 // ID of the active installation for the selected binary

	// Add binary view state - URL input
	UrlTextInput textinput.Model

	// Add binary view state - Form
	ParsedBinary *ParsedBinaryConfig
	FormInputs   []textinput.Model
	FocusedField int

	// Install binary view state
	InstallBinaryID      string
	InstallVersionInput  textinput.Model
	InstallingInProgress bool
	InstallReturnView    ViewState

	// Remove confirmation state
	ConfirmingRemove bool
	RemoveBinaryID   string
	RemoveWithFiles  bool

	// Import binary view state
	ImportPathInput textinput.Model
	ImportNameInput textinput.Model
	ImportFocusIdx  int

	// Error state
	ErrorMessage   string
	SuccessMessage string

	// Helpers
	RenderTabsFn  func() string
	GetHelpTextFn func(ViewState) string
}

// BinaryWithMetadata represents a binary with its metadata
type BinaryWithMetadata struct {
	Binary        *database.Binary
	ActiveVersion string
	InstallCount  int
}

// ParsedBinaryConfig represents a binary configuration parsed from a URL
type ParsedBinaryConfig struct {
	UserID        string
	Name          string
	Provider      string
	Path          string
	Format        string
	Version       string
	AssetName     string
	InstallPath   string
	AssetRegex    string
	ReleaseRegex  string
	Authenticated bool
}
