package tui

import "cturner8/binmate/internal/tui/views"

func (m Model) View() string {
	// Convert tui.Model to views.Model
	vm := m.toViewModel()

	switch m.CurrentView {
	case views.BinariesList:
		return vm.RenderBinariesList()
	case views.Versions:
		return vm.RenderVersions()
	case views.AddBinaryURL:
		return vm.RenderAddBinaryURL()
	case views.AddBinaryForm:
		return vm.RenderAddBinaryForm()
	case views.InstallBinary:
		return vm.RenderInstallBinary()
	case views.ImportBinary:
		return vm.RenderImportBinary()
	case views.Downloads:
		return vm.RenderDownloads()
	case views.Configuration:
		return vm.RenderConfiguration()
	case views.Help:
		return vm.RenderHelp()
	default:
		return "Unknown view"
	}
}

// toViewModel converts the tui Model to a views Model
func (m Model) toViewModel() views.Model {
	// Get active installation ID if we have a selected binary
	var activeInstallationID int64
	if m.SelectedBinary != nil {
		activeVersion, _ := getActiveVersion(m.DbService, m.SelectedBinary.ID)
		if activeVersion != nil {
			activeInstallationID = activeVersion.ID
		}
	}

	// Convert binaries to views format
	viewBinaries := make([]views.BinaryWithMetadata, len(m.Binaries))
	for i, b := range m.Binaries {
		viewBinaries[i] = views.BinaryWithMetadata{
			Binary:        b.Binary,
			ActiveVersion: b.ActiveVersion,
			InstallCount:  b.InstallCount,
		}
	}

	// Convert parsed binary if it exists
	var parsedBinary *views.ParsedBinaryConfig
	if m.ParsedBinary != nil {
		parsedBinary = &views.ParsedBinaryConfig{
			UserID:        m.ParsedBinary.userID,
			Name:          m.ParsedBinary.name,
			Provider:      m.ParsedBinary.provider,
			Path:          m.ParsedBinary.path,
			Format:        m.ParsedBinary.format,
			Version:       m.ParsedBinary.version,
			AssetName:     m.ParsedBinary.assetName,
			InstallPath:   m.ParsedBinary.installPath,
			AssetRegex:    m.ParsedBinary.assetRegex,
			ReleaseRegex:  m.ParsedBinary.releaseRegex,
			Authenticated: m.ParsedBinary.authenticated,
		}
	}

	return views.Model{
		DbService:            m.DbService,
		Config:               m.Config,
		CurrentView:          m.CurrentView,
		Width:                m.Width,
		Height:               m.Height,
		Binaries:             viewBinaries,
		SelectedIndex:        m.SelectedIndex,
		Loading:              m.Loading,
		SelectedBinary:       m.SelectedBinary,
		Installations:        m.Installations,
		SelectedVersionIdx:   m.SelectedVersionIdx,
		ActiveInstallationID: activeInstallationID,
		UrlTextInput:         m.UrlTextInput,
		ParsedBinary:         parsedBinary,
		FormInputs:           m.FormInputs,
		FocusedField:         m.FocusedField,
		InstallBinaryID:      m.InstallBinaryID,
		InstallVersionInput:  m.InstallVersionInput,
		InstallingInProgress: m.InstallingInProgress,
		InstallReturnView:    m.InstallReturnView,
		ConfirmingRemove:     m.ConfirmingRemove,
		RemoveBinaryID:       m.RemoveBinaryID,
		RemoveWithFiles:      m.RemoveWithFiles,
		ImportPathInput:      m.ImportPathInput,
		ImportNameInput:      m.ImportNameInput,
		ImportFocusIdx:       m.ImportFocusIdx,
		ErrorMessage:         m.ErrorMessage,
		SuccessMessage:       m.SuccessMessage,
		RenderTabsFn:         m.renderTabs,
		GetHelpTextFn:        getHelpText,
	}
}
