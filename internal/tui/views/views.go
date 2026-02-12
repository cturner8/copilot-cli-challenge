package views

import "cturner8/binmate/internal/database"

// ViewState represents the current view in the TUI
type ViewState int

const (
	BinariesList ViewState = iota
	Versions
	AddBinaryURL
	AddBinaryForm
	InstallBinary
	ImportBinary
	Downloads
	Configuration
	Help
)

// String returns the string representation of the view state
func (v ViewState) String() string {
	switch v {
	case BinariesList:
		return "Binaries List"
	case Versions:
		return "Versions"
	case AddBinaryURL:
		return "Add Binary - URL"
	case AddBinaryForm:
		return "Add Binary - Configuration"
	case InstallBinary:
		return "Install Binary"
	case ImportBinary:
		return "Import Binary"
	case Downloads:
		return "Downloads"
	case Configuration:
		return "Configuration"
	case Help:
		return "Help"
	default:
		return "Unknown"
	}
}

// Renderer defines the interface for types that can provide rendering context
type Renderer interface {
	// Styling
	RenderTitle(text string) string
	RenderError(text string) string
	RenderSuccess(text string) string
	RenderHeader(text string) string
	RenderLoading(text string) string
	RenderEmptyState(text string) string
	RenderHelp(text string) string

	// Helpers
	RenderTabs() string
	GetHelpText(view ViewState) string

	// State access
	GetCurrentView() ViewState
	GetWidth() int
	GetHeight() int
	GetErrorMessage() string
	GetSuccessMessage() string
	GetLoading() bool
	GetBinaries() []interface{}
	GetSelectedIndex() int
	GetConfirmingRemove() bool
	GetRemoveBinaryID() string
	GetSelectedBinary() *database.Binary
	GetInstallations() []*database.Installation
	GetSelectedVersionIdx() int
}
