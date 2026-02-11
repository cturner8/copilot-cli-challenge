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
	keySync      = "s" // 's' for sync (in config view)
	keySearch    = "/" // '/' for search
	keyFilter    = "f" // 'f' for filter
	keySortOrder = "o" // 'o' for order/sort
	keyNextSort  = "n" // 'n' for next sort mode
	keySpace     = " " // space for toggle selection in bulk mode
	keyBulkMode  = "b" // 'b' for bulk selection mode

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
		return "↑/↓: navigate • enter: versions • /: search • f: filter • o: sort • b: bulk mode • space: select (bulk) • a: add • i: install • u: update • r: remove • q: quit"
	case viewVersions:
		return "↑/↓: navigate • s/enter: switch version • i: install new version • u: update • c: check updates • d/delete: delete version • esc: back to list • q: quit"
	case viewAddBinaryURL:
		return "Type URL • enter: parse • esc: cancel • q: quit"
	case viewAddBinaryForm:
		return "tab/shift+tab: navigate fields • ctrl+s: save • esc: cancel • q: quit"
	case viewDownloads:
		return "1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case viewConfiguration:
		return "s: sync config to database • 1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case viewHelp:
		return "1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	default:
		return "q: quit"
	}
}
