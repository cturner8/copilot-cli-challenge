package tui

import (
	"testing"

	"cturner8/binmate/internal/core/config"
	"cturner8/binmate/internal/database/repository"
)

func TestInitialModel(t *testing.T) {
	// Create a mock service and config
	dbService := &repository.Service{}
	cfg := &config.Config{
		Version: 1,
	}

	m := initialModel(dbService, cfg)

	// Test that model is initialized with correct defaults
	if m.dbService != dbService {
		t.Errorf("initialModel() dbService = %v, want %v", m.dbService, dbService)
	}

	if m.config != cfg {
		t.Errorf("initialModel() config = %v, want %v", m.config, cfg)
	}

	if m.currentView != viewBinariesList {
		t.Errorf("initialModel() currentView = %v, want %v", m.currentView, viewBinariesList)
	}

	if !m.loading {
		t.Errorf("initialModel() loading = %v, want true", m.loading)
	}

	if m.selectedIndex != 0 {
		t.Errorf("initialModel() selectedIndex = %d, want 0", m.selectedIndex)
	}

	// Test that text inputs are initialized
	if m.urlTextInput.CharLimit != 256 {
		t.Errorf("initialModel() urlTextInput.CharLimit = %d, want 256", m.urlTextInput.CharLimit)
	}

	if m.installVersionInput.CharLimit != 64 {
		t.Errorf("initialModel() installVersionInput.CharLimit = %d, want 64", m.installVersionInput.CharLimit)
	}

	if m.importPathInput.CharLimit != 256 {
		t.Errorf("initialModel() importPathInput.CharLimit = %d, want 256", m.importPathInput.CharLimit)
	}

	if m.importNameInput.CharLimit != 64 {
		t.Errorf("initialModel() importNameInput.CharLimit = %d, want 64", m.importNameInput.CharLimit)
	}

	// Test initial state of slices
	if m.binaries != nil {
		t.Errorf("initialModel() binaries = %v, want nil", m.binaries)
	}

	if m.installations != nil {
		t.Errorf("initialModel() installations = %v, want nil", m.installations)
	}

	if len(m.formInputs) != 0 {
		t.Errorf("initialModel() len(formInputs) = %d, want 0", len(m.formInputs))
	}
}

func TestParsedBinaryConfig(t *testing.T) {
	// Test that parsedBinaryConfig struct can be created
	parsed := &parsedBinaryConfig{
		userID:        "test-binary",
		name:          "Test Binary",
		provider:      "github",
		path:          "owner/repo",
		format:        "tar.gz",
		version:       "v1.0.0",
		assetName:     "binary.tar.gz",
		installPath:   "/usr/local/bin",
		assetRegex:    "binary-.*",
		releaseRegex:  "v.*",
		authenticated: false,
	}

	if parsed.userID != "test-binary" {
		t.Errorf("userID = %s, want test-binary", parsed.userID)
	}

	if parsed.name != "Test Binary" {
		t.Errorf("name = %s, want Test Binary", parsed.name)
	}

	if parsed.provider != "github" {
		t.Errorf("provider = %s, want github", parsed.provider)
	}

	if parsed.authenticated != false {
		t.Errorf("authenticated = %v, want false", parsed.authenticated)
	}
}

func TestViewStateString(t *testing.T) {
	tests := []struct {
		name     string
		state    viewState
		expected string
	}{
		{
			name:     "binaries list view",
			state:    viewBinariesList,
			expected: "Binaries List",
		},
		{
			name:     "versions view",
			state:    viewVersions,
			expected: "Versions",
		},
		{
			name:     "add binary URL view",
			state:    viewAddBinaryURL,
			expected: "Add Binary - URL",
		},
		{
			name:     "add binary form view",
			state:    viewAddBinaryForm,
			expected: "Add Binary - Configuration",
		},
		{
			name:     "install binary view",
			state:    viewInstallBinary,
			expected: "Install Binary",
		},
		{
			name:     "import binary view",
			state:    viewImportBinary,
			expected: "Import Binary",
		},
		{
			name:     "downloads view",
			state:    viewDownloads,
			expected: "Downloads",
		},
		{
			name:     "configuration view",
			state:    viewConfiguration,
			expected: "Configuration",
		},
		{
			name:     "help view",
			state:    viewHelp,
			expected: "Help",
		},
		{
			name:     "unknown view",
			state:    viewState(999),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("viewState.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCreateFormInputs(t *testing.T) {
	parsed := &parsedBinaryConfig{
		userID:        "test-binary",
		name:          "Test Binary",
		provider:      "github",
		path:          "owner/repo",
		format:        "tar.gz",
		version:       "v1.0.0",
		assetName:     "binary.tar.gz",
		installPath:   "/usr/local/bin",
		assetRegex:    "binary-.*",
		releaseRegex:  "v.*",
		authenticated: true,
	}

	inputs := createFormInputs(parsed)

	// Test that we have the correct number of inputs
	expectedInputs := 9
	if len(inputs) != expectedInputs {
		t.Fatalf("createFormInputs() returned %d inputs, want %d", len(inputs), expectedInputs)
	}

	// Test that inputs are pre-populated with correct values
	if inputs[0].Value() != parsed.userID {
		t.Errorf("input[0] (userID) = %q, want %q", inputs[0].Value(), parsed.userID)
	}

	if inputs[1].Value() != parsed.name {
		t.Errorf("input[1] (name) = %q, want %q", inputs[1].Value(), parsed.name)
	}

	if inputs[2].Value() != parsed.provider {
		t.Errorf("input[2] (provider) = %q, want %q", inputs[2].Value(), parsed.provider)
	}

	if inputs[3].Value() != parsed.path {
		t.Errorf("input[3] (path) = %q, want %q", inputs[3].Value(), parsed.path)
	}

	if inputs[4].Value() != parsed.format {
		t.Errorf("input[4] (format) = %q, want %q", inputs[4].Value(), parsed.format)
	}

	if inputs[5].Value() != parsed.installPath {
		t.Errorf("input[5] (installPath) = %q, want %q", inputs[5].Value(), parsed.installPath)
	}

	if inputs[6].Value() != parsed.assetRegex {
		t.Errorf("input[6] (assetRegex) = %q, want %q", inputs[6].Value(), parsed.assetRegex)
	}

	if inputs[7].Value() != parsed.releaseRegex {
		t.Errorf("input[7] (releaseRegex) = %q, want %q", inputs[7].Value(), parsed.releaseRegex)
	}

	if inputs[8].Value() != "true" {
		t.Errorf("input[8] (authenticated) = %q, want %q", inputs[8].Value(), "true")
	}
}

func TestCreateFormInputs_Unauthenticated(t *testing.T) {
	parsed := &parsedBinaryConfig{
		userID:        "test-binary",
		name:          "Test Binary",
		provider:      "github",
		path:          "owner/repo",
		format:        "tar.gz",
		authenticated: false,
	}

	inputs := createFormInputs(parsed)

	// Test authenticated field when false
	if inputs[8].Value() != "false" {
		t.Errorf("input[8] (authenticated) = %q, want %q", inputs[8].Value(), "false")
	}
}
