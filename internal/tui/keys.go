package tui

// Key binding constants
const (
	// Navigation keys
	keyUp    = "up"
	keyDown  = "down"
	keyLeft  = "left"
	keyRight = "right"
	keyEnter = "enter"
	keyEsc   = "esc"
	keyTab   = "tab"

	// Action keys
	keyAdd  = "a"
	keyQuit = "q"
	keySave = "ctrl+s"

	// Tab cycling keys
	keyShiftTab     = "shift+tab"
	keyCtrlShiftTab = "ctrl+shift+tab"

	// Special keys
	keyCtrlC = "ctrl+c"
)

// getHelpText returns context-sensitive help text for the current view
func getHelpText(view viewState) string {
	switch view {
	case viewBinariesList:
		return "↑/↓: navigate • enter: view versions • a: add binary • 1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case viewVersions:
		return "↑/↓: navigate • esc: back to list • q: quit"
	case viewAddBinaryURL:
		return "Type URL • enter: parse • esc: cancel • q: quit"
	case viewAddBinaryForm:
		return "tab/shift+tab: navigate fields • ctrl+s: save • esc: cancel • q: quit"
	case viewDownloads:
		return "1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case viewConfiguration:
		return "1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case viewHelp:
		return "1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	default:
		return "q: quit"
	}
}
