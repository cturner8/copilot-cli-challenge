package tui

import "cturner8/binmate/internal/tui/views"

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

	// Tab cycling keys
	tab             = "tab"
	keyShiftTab     = "shift+tab"
	keyCtrlShiftTab = "ctrl+shift+tab"

	// Special keys
	keyCtrlC = "ctrl+c"
)

// getHelpText returns context-sensitive help text for the current view
func getHelpText(view views.ViewState) string {
	switch view {
	case views.BinariesList:
		return "↑/↓: navigate • enter: view versions • a: add binary • i: install • u: update • r: remove • c: check updates • m: import • 1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case views.Versions:
		return "↑/↓: navigate • s/enter: switch version • i: install new version • u: update • c: check updates • d/delete: delete version • esc: back to list • q: quit"
	case views.AddBinaryURL:
		return "Type URL • enter: parse • esc: cancel • q: quit"
	case views.AddBinaryForm:
		return "tab/shift+tab: navigate fields • ctrl+s: save • esc: cancel • q: quit"
	case views.Downloads:
		return "1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case views.Configuration:
		return "s: sync config to database • 1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	case views.Help:
		return "1-4/shift+tab/ctrl+shift+tab: switch tabs • q: quit"
	default:
		return "q: quit"
	}
}
