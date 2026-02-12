package views

// getHelpText returns context-specific help text for each view
func getHelpText(view ViewState) string {
	switch view {
	case BinariesList:
		return "↑/↓: navigate • enter: versions • a: add • i: install • u/U: update • r: remove • c: check • m: import • 1-4/tab: switch tabs • q: quit"
	case Versions:
		return "↑/↓: navigate • s/enter: switch • d: delete • esc: back • q: quit"
	case AddBinaryURL:
		return "enter: parse URL • esc: cancel • ctrl+c: quit"
	case AddBinaryForm:
		return "↑/↓/tab: navigate fields • enter: save • esc: cancel"
	case InstallBinary:
		return "enter: confirm install • esc: cancel"
	case ImportBinary:
		return "↑/↓/tab: navigate fields • enter: import • esc: cancel"
	case Downloads:
		return "1-4/tab: switch tabs • q: quit"
	case Configuration:
		return "s: sync config • 1-4/tab: switch tabs • q: quit"
	case Help:
		return "1-4/tab: switch tabs • q: quit"
	default:
		return "q: quit"
	}
}
