package tui

// viewState represents the current view in the TUI
type viewState int

const (
	viewBinariesList viewState = iota
	viewVersions
	viewAddBinaryURL
	viewAddBinaryForm
	viewDownloads
	viewConfiguration
	viewHelp
)

// String returns the string representation of the view state
func (v viewState) String() string {
	switch v {
	case viewBinariesList:
		return "Binaries List"
	case viewVersions:
		return "Versions"
	case viewAddBinaryURL:
		return "Add Binary - URL"
	case viewAddBinaryForm:
		return "Add Binary - Configuration"
	case viewDownloads:
		return "Downloads"
	case viewConfiguration:
		return "Configuration"
	case viewHelp:
		return "Help"
	default:
		return "Unknown"
	}
}
