package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"cturner8/binmate/internal/database/repository"
)

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("binmate"),
		loadBinaries(m.dbService),
	)
}

// loadBinaries returns a command to load binaries from the database
func loadBinaries(dbService *repository.Service) tea.Cmd {
	return func() tea.Msg {
		binaries, err := getBinariesWithMetadata(dbService)
		return binariesLoadedMsg{binaries: binaries, err: err}
	}
}
