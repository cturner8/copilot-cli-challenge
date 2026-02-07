package tui

func (m model) View() string {
	switch m.currentView {
	case viewBinariesList:
		return m.renderBinariesList()
	case viewVersions:
		return m.renderVersions()
	case viewAddBinaryURL:
		return m.renderAddBinaryURL()
	case viewAddBinaryForm:
		return m.renderAddBinaryForm()
	case viewInstallBinary:
		return m.renderInstallBinary()
	case viewDownloads:
		return m.renderDownloads()
	case viewConfiguration:
		return m.renderConfiguration()
	case viewHelp:
		return m.renderHelp()
	default:
		return "Unknown view"
	}
}
