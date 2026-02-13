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

	// GitHub navigation keys (used in versions view)
	keyReleaseNotes  = "l" // 'l' for release notes/logs
	keyRepoInfo      = "g" // 'g' for GitHub repo info
	keyAvailVersions = "v" // 'v' for view available versions

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
		return "↑/↓: navigate • enter: versions • /: search • f: filter • o: sort order • n: next sort • a: add • i: install • u: update • r: remove • q: quit"
	case viewVersions:
		return "↑/↓: navigate • s/enter: switch • i: install • u: update • c: check • d: delete • l: release notes • g: repo info • v: versions • esc: back • q: quit"
	case viewAddBinaryURL:
		return "Type URL • enter: parse • esc: cancel • q: quit"
	case viewAddBinaryForm:
		return "tab/shift+tab: navigate fields • ctrl+s: save • esc: cancel • q: quit"
	case viewDownloads:
		return "1-3/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case viewConfiguration:
		return "s: sync config to database • 1-3/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case viewHelp:
		return "1-3/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case viewReleaseNotes:
		return "esc: back • q: quit"
	case viewAvailableVersions:
		return "↑/↓: navigate • enter/l: release notes • i: install selected • esc: back • q: quit"
	case viewRepositoryInfo:
		return "s: star repository • esc: back • q: quit"
	default:
		return "q: quit"
	}
}
