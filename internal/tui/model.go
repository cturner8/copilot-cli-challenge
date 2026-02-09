package tui

import (
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"

	"github.com/charmbracelet/bubbles/textinput"
)

type model struct {
	// Services
	dbService *repository.Service
	config    *config.Config

	// View state
	currentView viewState

	// Window dimensions
	width  int
	height int

	// Binaries list view state
	binaries      []BinaryWithMetadata
	selectedIndex int
	loading       bool

	// Versions view state
	selectedBinary     *database.Binary
	installations      []*database.Installation
	selectedVersionIdx int

	// Add binary view state - URL input
	urlTextInput textinput.Model

	// Add binary view state - Form
	parsedBinary *parsedBinaryConfig
	formInputs   []textinput.Model
	focusedField int

	// Install binary view state
	installBinaryID      string
	installVersionInput  textinput.Model
	installingInProgress bool

	// Remove confirmation state
	confirmingRemove bool
	removeBinaryID   string
	removeWithFiles  bool

	// Import binary view state
	importPathInput textinput.Model
	importNameInput textinput.Model
	importFocusIdx  int

	// Error state
	errorMessage   string
	successMessage string
}

// parsedBinaryConfig represents a binary configuration parsed from a URL
type parsedBinaryConfig struct {
	userID       string
	name         string
	provider     string
	path         string
	format       string
	version      string
	assetName    string
	installPath  string
	assetRegex   string
	releaseRegex string
}

func initialModel(dbService *repository.Service, cfg *config.Config) model {
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

	return model{
		dbService:           dbService,
		config:              cfg,
		currentView:         viewBinariesList,
		loading:             true,
		urlTextInput:        urlInput,
		formInputs:          []textinput.Model{},
		installVersionInput: versionInput,
		importPathInput:     importPathInput,
		importNameInput:     importNameInput,
	}
}
