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

	// Search state
	searchMode       bool
	searchQuery      string
	searchTextInput  textinput.Model
	filteredBinaries []BinaryWithMetadata

	// Filter and sort state
	filterPanelOpen bool
	activeFilters   map[string]string // e.g., "provider": "github", "format": ".tar.gz"
	sortMode        string            // "name", "provider", "updated", "count"
	sortAscending   bool

	// Bulk operations state
	bulkSelectMode          bool
	selectedBinaries        map[int]bool // Map of selected indices in the current display
	bulkRemoveCount         int          // Count of binaries to remove in bulk mode
	bulkOperationInProgress bool         // Track if a bulk operation is running

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
	installReturnView    viewState // Track which view to return to after install

	// Remove confirmation state
	confirmingRemove bool
	removeBinaryID   string
	removeWithFiles  bool

	// Import binary view state
	importPathInput    textinput.Model
	importNameInput    textinput.Model
	importURLInput     textinput.Model
	importVersionInput textinput.Model
	importFocusIdx     int

	// GitHub views state
	githubReleaseInfo           *githubReleaseInfo
	githubAvailableVers         []githubReleaseInfo
	selectedAvailableVersionIdx int
	githubRepoInfo              *githubRepositoryInfo
	githubLoading               bool
	githubError                 string

	// Error state
	errorMessage   string
	successMessage string
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

// githubReleaseInfo holds GitHub release information for TUI display
type githubReleaseInfo struct {
	Name        string
	TagName     string
	Body        string
	Prerelease  bool
	PublishedAt string
	HTMLURL     string
}

// githubRepositoryInfo holds GitHub repository information for TUI display
type githubRepositoryInfo struct {
	Name        string
	FullName    string
	Description string
	Stars       int
	Forks       int
	HTMLURL     string
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

	importURLInput := textinput.New()
	importURLInput.Placeholder = "https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz (optional)"
	importURLInput.CharLimit = 256
	importURLInput.Width = 80

	importVersionInput := textinput.New()
	importVersionInput.Placeholder = "v1.0.0 (optional, auto-extracted from URL)"
	importVersionInput.CharLimit = 64
	importVersionInput.Width = 50

	// Create search text input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search by name (regex supported)..."
	searchInput.CharLimit = 128
	searchInput.Width = 60

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
		importURLInput:      importURLInput,
		importVersionInput:  importVersionInput,
		searchTextInput:     searchInput,
		searchMode:          false,
		filteredBinaries:    []BinaryWithMetadata{},
		activeFilters:       make(map[string]string),
		sortMode:            "name",
		sortAscending:       true,
		bulkSelectMode:      false,
		selectedBinaries:    make(map[int]bool),
	}
}
