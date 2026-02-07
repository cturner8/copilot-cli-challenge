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
	keyAdd       = "a"
	keyQuit      = "q"
	keySave      = "ctrl+s"
	keySwitch    = "s"
	keyDelete    = "d"
	keyDelete2   = "delete"
	keyInstall   = "i"
	keyUpdate    = "u"
	keyUpdateAll = "U"
	keyRemove    = "r"
	keyCheck     = "c"
	keyImport    = "m" // 'm' for iMport

	// Tab cycling keys
	tab             = "tab"
	keyShiftTab     = "shift+tab"
	keyCtrlShiftTab = "ctrl+shift+tab"

	// Special keys
	keyCtrlC = "ctrl+c"
)

// getHelpText returns context-sensitive help text for the current view
func getHelpText(view viewState) string {
	switch view {
	case viewBinariesList:
		return "↑/↓: navigate • enter: view versions • a: add binary • i: install • u: update • r: remove • c: check updates • m: import • 1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case viewVersions:
		return "↑/↓: navigate • s/enter: switch version • d/delete: delete version • esc: back to list • q: quit"
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
