package tui

import (
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

type model struct {
	// Services
	dbService *repository.Service
	config    *config.Config

	// View state
	currentView viewState

	// Binaries list view state
	binaries      []BinaryWithMetadata
	selectedIndex int
	loading       bool

	// Versions view state
	selectedBinary *database.Binary
	installations  []*database.Installation

	// Add binary view state
	urlInput      string
	parsedBinary  *parsedBinaryConfig
	formFields    map[string]string
	focusedField  int

	// Error state
	errorMessage string
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
	return model{
		dbService:   dbService,
		config:      cfg,
		currentView: viewBinariesList,
		loading:     true,
		formFields:  make(map[string]string),
	}
}
