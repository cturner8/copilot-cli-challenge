package tui

import (
	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database/repository"
	tea "github.com/charmbracelet/bubbletea"
)

func InitProgram(dbService *repository.Service, cfg *config.Config) *tea.Program {
	return tea.NewProgram(initialModel(dbService, cfg))
}
