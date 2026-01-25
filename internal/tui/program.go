package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func InitProgram() *tea.Program {
	return tea.NewProgram(initialModel())
}
