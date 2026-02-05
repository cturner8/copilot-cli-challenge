package tui

import (
	"cturner8/binmate/internal/database/repository"
	tea "github.com/charmbracelet/bubbletea"
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
