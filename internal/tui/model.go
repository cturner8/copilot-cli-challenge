package tui

import (
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
	"cturner8/binmate/internal/tui/views"

	"github.com/charmbracelet/bubbles/textinput"
)

type Model struct {
	// Services
	DbService *repository.Service
	Config    *config.Config

	// View state
	CurrentView views.ViewState

	// Window dimensions
	Width  int
	Height int

	// Binaries list view state
	Binaries      []BinaryWithMetadata
	SelectedIndex int
	Loading       bool

	// Versions view state
	SelectedBinary     *database.Binary
	Installations      []*database.Installation
	SelectedVersionIdx int

	// Add binary view state - URL input
	UrlTextInput textinput.Model

	// Add binary view state - Form
	ParsedBinary *parsedBinaryConfig
	FormInputs   []textinput.Model
	FocusedField int

	// Install binary view state
	InstallBinaryID      string
	InstallVersionInput  textinput.Model
	InstallingInProgress bool
	InstallReturnView    views.ViewState // Track which view to return to after install

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
}

// parsedBinaryConfig represents a binary configuration parsed from a URL
type parsedBinaryConfig struct {
	userID        string
	name          string
	provider      string
	path          string
	format        string
	version       string
	assetName     string
	installPath   string
	assetRegex    string
	releaseRegex  string
	authenticated bool
}

func initialModel(dbService *repository.Service, cfg *config.Config) Model {
	// Create URL text input
	urlInput := textinput.New()
	urlInput.Placeholder = "https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz"
	urlInput.CharLimit = 256
	urlInput.Width = 80

	// Create version text input for install view
	versionInput := textinput.New()
	versionInput.Placeholder = "latest"
	versionInput.CharLimit = 64
	versionInput.Width = 40

	// Create text inputs for import view
	importPathInput := textinput.New()
	importPathInput.Placeholder = "/usr/local/bin/binary"
	importPathInput.CharLimit = 256
	importPathInput.Width = 60

	importNameInput := textinput.New()
	importNameInput.Placeholder = "binary-name"
	importNameInput.CharLimit = 64
	importNameInput.Width = 40

	return Model{
		DbService:           dbService,
		Config:              cfg,
		CurrentView:         views.BinariesList,
		Loading:             true,
		UrlTextInput:        urlInput,
		FormInputs:          []textinput.Model{},
		InstallVersionInput: versionInput,
		ImportPathInput:     importPathInput,
		ImportNameInput:     importNameInput,
	}
}
