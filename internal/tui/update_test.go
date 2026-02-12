package tui

import (
	"testing"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := model{
		dbService: &repository.Service{},
		config:    &config.Config{},
	}

	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}

	updatedModel, _ := m.Update(msg)
	m2 := updatedModel.(model)

	if m2.width != 120 {
		t.Errorf("Update(WindowSizeMsg) width = %d, want 120", m2.width)
	}

	if m2.height != 40 {
		t.Errorf("Update(WindowSizeMsg) height = %d, want 40", m2.height)
	}
}

func TestUpdateBinariesList_NavigateUp(t *testing.T) {
	m := model{
		currentView: viewBinariesList,
		binaries: []BinaryWithMetadata{
			{Binary: &database.Binary{ID: 1, Name: "Binary1"}},
			{Binary: &database.Binary{ID: 2, Name: "Binary2"}},
			{Binary: &database.Binary{ID: 3, Name: "Binary3"}},
		},
		selectedIndex: 2,
		dbService:     &repository.Service{},
		config:        &config.Config{},
	}

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.updateBinariesList(msg)
	m2 := updatedModel.(model)

	if m2.selectedIndex != 1 {
		t.Errorf("updateBinariesList(up) selectedIndex = %d, want 1", m2.selectedIndex)
	}
}

func TestUpdateBinariesList_NavigateUpAtTop(t *testing.T) {
	m := model{
		currentView: viewBinariesList,
		binaries: []BinaryWithMetadata{
			{Binary: &database.Binary{ID: 1, Name: "Binary1"}},
			{Binary: &database.Binary{ID: 2, Name: "Binary2"}},
		},
		selectedIndex: 0,
		dbService:     &repository.Service{},
		config:        &config.Config{},
	}

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.updateBinariesList(msg)
	m2 := updatedModel.(model)

	if m2.selectedIndex != 0 {
		t.Errorf("updateBinariesList(up at top) selectedIndex = %d, want 0", m2.selectedIndex)
	}
}

func TestUpdateBinariesList_NavigateDown(t *testing.T) {
	m := model{
		currentView: viewBinariesList,
		binaries: []BinaryWithMetadata{
			{Binary: &database.Binary{ID: 1, Name: "Binary1"}},
			{Binary: &database.Binary{ID: 2, Name: "Binary2"}},
			{Binary: &database.Binary{ID: 3, Name: "Binary3"}},
		},
		selectedIndex: 0,
		dbService:     &repository.Service{},
		config:        &config.Config{},
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := m.updateBinariesList(msg)
	m2 := updatedModel.(model)

	if m2.selectedIndex != 1 {
		t.Errorf("updateBinariesList(down) selectedIndex = %d, want 1", m2.selectedIndex)
	}
}

func TestUpdateBinariesList_NavigateDownAtBottom(t *testing.T) {
	m := model{
		currentView: viewBinariesList,
		binaries: []BinaryWithMetadata{
			{Binary: &database.Binary{ID: 1, Name: "Binary1"}},
			{Binary: &database.Binary{ID: 2, Name: "Binary2"}},
		},
		selectedIndex: 1,
		dbService:     &repository.Service{},
		config:        &config.Config{},
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := m.updateBinariesList(msg)
	m2 := updatedModel.(model)

	if m2.selectedIndex != 1 {
		t.Errorf("updateBinariesList(down at bottom) selectedIndex = %d, want 1", m2.selectedIndex)
	}
}

func TestUpdateBinariesList_EnterViewVersions(t *testing.T) {
	m := model{
		currentView: viewBinariesList,
		binaries: []BinaryWithMetadata{
			{Binary: &database.Binary{ID: 1, Name: "Binary1", UserID: "binary1"}},
			{Binary: &database.Binary{ID: 2, Name: "Binary2", UserID: "binary2"}},
		},
		selectedIndex: 0,
		dbService:     &repository.Service{},
		config:        &config.Config{},
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := m.updateBinariesList(msg)
	m2 := updatedModel.(model)

	if m2.currentView != viewVersions {
		t.Errorf("updateBinariesList(enter) currentView = %v, want %v", m2.currentView, viewVersions)
	}

	if m2.selectedBinary == nil {
		t.Fatal("updateBinariesList(enter) selectedBinary should not be nil")
	}

	if m2.selectedBinary.ID != 1 {
		t.Errorf("updateBinariesList(enter) selectedBinary.ID = %d, want 1", m2.selectedBinary.ID)
	}

	if !m2.loading {
		t.Errorf("updateBinariesList(enter) loading = %v, want true", m2.loading)
	}

	if cmd == nil {
		t.Error("updateBinariesList(enter) should return a command to load versions")
	}
}

func TestUpdateBinariesList_PressAddKey(t *testing.T) {
	m := initialModel(&repository.Service{}, &config.Config{})
	m.currentView = viewBinariesList

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, _ := m.updateBinariesList(msg)
	m2 := updatedModel.(model)

	if m2.currentView != viewAddBinaryURL {
		t.Errorf("updateBinariesList('a') currentView = %v, want %v", m2.currentView, viewAddBinaryURL)
	}
}

func TestUpdateBinariesList_RemoveConfirmation(t *testing.T) {
	m := initialModel(&repository.Service{}, &config.Config{})
	m.currentView = viewBinariesList
	m.binaries = []BinaryWithMetadata{
		{Binary: &database.Binary{ID: 1, Name: "Binary1", UserID: "binary1"}},
	}
	m.selectedIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, _ := m.updateBinariesList(msg)
	m2 := updatedModel.(model)

	if !m2.confirmingRemove {
		t.Errorf("updateBinariesList('r') confirmingRemove = %v, want true", m2.confirmingRemove)
	}

	if m2.removeBinaryID != "binary1" {
		t.Errorf("updateBinariesList('r') removeBinaryID = %q, want %q", m2.removeBinaryID, "binary1")
	}
}

func TestUpdateBinariesList_RemoveConfirmYes(t *testing.T) {
	m := model{
		currentView:      viewBinariesList,
		confirmingRemove: true,
		removeBinaryID:   "binary1",
		dbService:        &repository.Service{},
		config:           &config.Config{},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	updatedModel, cmd := m.updateBinariesList(msg)
	m2 := updatedModel.(model)

	if m2.confirmingRemove {
		t.Errorf("updateBinariesList('y') confirmingRemove = %v, want false", m2.confirmingRemove)
	}

	if m2.removeBinaryID != "" {
		t.Errorf("updateBinariesList('y') removeBinaryID = %q, want empty", m2.removeBinaryID)
	}

	if cmd == nil {
		t.Error("updateBinariesList('y') should return a command to remove binary")
	}
}

func TestUpdateBinariesList_RemoveConfirmCancel(t *testing.T) {
	m := model{
		currentView:      viewBinariesList,
		confirmingRemove: true,
		removeBinaryID:   "binary1",
		dbService:        &repository.Service{},
		config:           &config.Config{},
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := m.updateBinariesList(msg)
	m2 := updatedModel.(model)

	if m2.confirmingRemove {
		t.Errorf("updateBinariesList(esc) confirmingRemove = %v, want false", m2.confirmingRemove)
	}

	if m2.removeBinaryID != "" {
		t.Errorf("updateBinariesList(esc) removeBinaryID = %q, want empty", m2.removeBinaryID)
	}
}

func TestUpdateVersions_NavigateUp(t *testing.T) {
	m := model{
		currentView: viewVersions,
		installations: []*database.Installation{
			{ID: 1, Version: "1.0.0"},
			{ID: 2, Version: "2.0.0"},
			{ID: 3, Version: "3.0.0"},
		},
		selectedVersionIdx: 2,
		dbService:          &repository.Service{},
		config:             &config.Config{},
	}

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.updateVersions(msg)
	m2 := updatedModel.(model)

	if m2.selectedVersionIdx != 1 {
		t.Errorf("updateVersions(up) selectedVersionIdx = %d, want 1", m2.selectedVersionIdx)
	}
}

func TestUpdateVersions_NavigateDown(t *testing.T) {
	m := model{
		currentView: viewVersions,
		installations: []*database.Installation{
			{ID: 1, Version: "1.0.0"},
			{ID: 2, Version: "2.0.0"},
		},
		selectedVersionIdx: 0,
		dbService:          &repository.Service{},
		config:             &config.Config{},
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := m.updateVersions(msg)
	m2 := updatedModel.(model)

	if m2.selectedVersionIdx != 1 {
		t.Errorf("updateVersions(down) selectedVersionIdx = %d, want 1", m2.selectedVersionIdx)
	}
}

func TestUpdateVersions_EscapeToList(t *testing.T) {
	m := model{
		currentView: viewVersions,
		selectedBinary: &database.Binary{
			ID:   1,
			Name: "Binary1",
		},
		installations: []*database.Installation{
			{ID: 1, Version: "1.0.0"},
		},
		dbService: &repository.Service{},
		config:    &config.Config{},
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := m.updateVersions(msg)
	m2 := updatedModel.(model)

	if m2.currentView != viewBinariesList {
		t.Errorf("updateVersions(esc) currentView = %v, want %v", m2.currentView, viewBinariesList)
	}

	if m2.selectedBinary != nil {
		t.Errorf("updateVersions(esc) selectedBinary should be nil")
	}

	if m2.installations != nil {
		t.Errorf("updateVersions(esc) installations should be nil")
	}
}

func TestGetTabForKey(t *testing.T) {
	tests := []struct {
		key      string
		expected viewState
		ok       bool
	}{
		{"1", viewBinariesList, true},
		{"2", viewDownloads, true},
		{"3", viewConfiguration, true},
		{"4", viewHelp, true},
		{"5", viewBinariesList, false},
		{"a", viewBinariesList, false},
		{"q", viewBinariesList, false},
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			view, ok := getTabForKey(tt.key)
			if ok != tt.ok {
				t.Errorf("getTabForKey(%q) ok = %v, want %v", tt.key, ok, tt.ok)
			}
			if ok && view != tt.expected {
				t.Errorf("getTabForKey(%q) view = %v, want %v", tt.key, view, tt.expected)
			}
		})
	}
}

func TestGetNextTab(t *testing.T) {
	tests := []struct {
		current  viewState
		expected viewState
	}{
		{viewBinariesList, viewDownloads},
		{viewDownloads, viewConfiguration},
		{viewConfiguration, viewHelp},
		{viewHelp, viewBinariesList}, // Wraps around
	}

	for _, tt := range tests {
		t.Run(tt.current.String(), func(t *testing.T) {
			result := getNextTab(tt.current)
			if result != tt.expected {
				t.Errorf("getNextTab(%v) = %v, want %v", tt.current, result, tt.expected)
			}
		})
	}
}

func TestGetPreviousTab(t *testing.T) {
	tests := []struct {
		current  viewState
		expected viewState
	}{
		{viewBinariesList, viewHelp}, // Wraps around to end
		{viewDownloads, viewBinariesList},
		{viewConfiguration, viewDownloads},
		{viewHelp, viewConfiguration},
	}

	for _, tt := range tests {
		t.Run(tt.current.String(), func(t *testing.T) {
			result := getPreviousTab(tt.current)
			if result != tt.expected {
				t.Errorf("getPreviousTab(%v) = %v, want %v", tt.current, result, tt.expected)
			}
		})
	}
}

func TestHandleTabCycling_Tab(t *testing.T) {
	m := initialModel(&repository.Service{}, &config.Config{})
	m.currentView = viewBinariesList

	updatedModel, handled := handleTabCycling(m, keyTab)

	if !handled {
		t.Error("handleTabCycling(tab) should return handled = true")
	}

	if updatedModel.currentView != viewDownloads {
		t.Errorf("handleTabCycling(tab) currentView = %v, want %v", updatedModel.currentView, viewDownloads)
	}
}

func TestHandleTabCycling_ShiftTab(t *testing.T) {
	m := initialModel(&repository.Service{}, &config.Config{})
	m.currentView = viewBinariesList

	updatedModel, handled := handleTabCycling(m, keyShiftTab)

	if !handled {
		t.Error("handleTabCycling(shift+tab) should return handled = true")
	}

	if updatedModel.currentView != viewHelp {
		t.Errorf("handleTabCycling(shift+tab) currentView = %v, want %v", updatedModel.currentView, viewHelp)
	}
}

func TestHandleTabCycling_NonTabKey(t *testing.T) {
	m := initialModel(&repository.Service{}, &config.Config{})
	m.currentView = viewBinariesList

	updatedModel, handled := handleTabCycling(m, "a")

	if handled {
		t.Error("handleTabCycling('a') should return handled = false")
	}

	if updatedModel.currentView != viewBinariesList {
		t.Errorf("handleTabCycling('a') should not change currentView")
	}
}

func TestUpdateBinariesList_TabSwitch(t *testing.T) {
	m := initialModel(&repository.Service{}, &config.Config{})
	m.currentView = viewBinariesList

	// Test switching to downloads with key "2"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	updatedModel, _ := m.updateBinariesList(msg)
	m2 := updatedModel.(model)

	if m2.currentView != viewDownloads {
		t.Errorf("updateBinariesList('2') currentView = %v, want %v", m2.currentView, viewDownloads)
	}
}

func TestUpdate_BinariesLoadedMsg_Success(t *testing.T) {
	m := model{
		loading:   true,
		dbService: &repository.Service{},
		config:    &config.Config{},
	}

	msg := binariesLoadedMsg{
		binaries: []BinaryWithMetadata{
			{Binary: &database.Binary{ID: 1, Name: "Binary1"}},
		},
		err: nil,
	}

	updatedModel, _ := m.Update(msg)
	m2 := updatedModel.(model)

	if m2.loading {
		t.Error("Update(binariesLoadedMsg) loading should be false")
	}

	if len(m2.binaries) != 1 {
		t.Errorf("Update(binariesLoadedMsg) len(binaries) = %d, want 1", len(m2.binaries))
	}

	if m2.errorMessage != "" {
		t.Errorf("Update(binariesLoadedMsg) errorMessage = %q, want empty", m2.errorMessage)
	}
}

func TestUpdate_BinariesLoadedMsg_Error(t *testing.T) {
	m := model{
		loading:   true,
		dbService: &repository.Service{},
		config:    &config.Config{},
	}

	msg := binariesLoadedMsg{
		binaries: nil,
		err:      database.ErrNotFound,
	}

	updatedModel, _ := m.Update(msg)
	m2 := updatedModel.(model)

	if m2.loading {
		t.Error("Update(binariesLoadedMsg error) loading should be false")
	}

	if m2.errorMessage == "" {
		t.Error("Update(binariesLoadedMsg error) errorMessage should not be empty")
	}
}

func TestUpdate_SuccessMsg(t *testing.T) {
	m := model{
		dbService: &repository.Service{},
		config:    &config.Config{},
	}

	msg := successMsg{
		message: "Operation successful",
	}

	updatedModel, _ := m.Update(msg)
	m2 := updatedModel.(model)

	if m2.successMessage != "Operation successful" {
		t.Errorf("Update(successMsg) successMessage = %q, want %q", m2.successMessage, "Operation successful")
	}

	if m2.errorMessage != "" {
		t.Errorf("Update(successMsg) errorMessage should be cleared")
	}
}

func TestUpdate_ErrorMsg(t *testing.T) {
	m := model{
		dbService: &repository.Service{},
		config:    &config.Config{},
	}

	msg := errorMsg{
		err: database.ErrNotFound,
	}

	updatedModel, _ := m.Update(msg)
	m2 := updatedModel.(model)

	if m2.errorMessage == "" {
		t.Error("Update(errorMsg) errorMessage should not be empty")
	}

	if m2.successMessage != "" {
		t.Errorf("Update(errorMsg) successMessage should be cleared")
	}
}
