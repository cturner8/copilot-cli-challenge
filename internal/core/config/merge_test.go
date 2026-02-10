package config

import (
	"testing"
)

func TestMergeBinaryWithGlobal(t *testing.T) {
	tests := []struct {
		name     string
		binary   Binary
		global   GlobalConfig
		expected Binary
	}{
		{
			name: "binary with no install path uses global",
			binary: Binary{
				Id:       "test",
				Name:     "test",
				Provider: "github",
				Path:     "owner/repo",
				Format:   ".tar.gz",
			},
			global: GlobalConfig{
				InstallPath: "/usr/local/bin",
			},
			expected: Binary{
				Id:          "test",
				Name:        "test",
				Provider:    "github",
				Path:        "owner/repo",
				Format:      ".tar.gz",
				InstallPath: "/usr/local/bin",
			},
		},
		{
			name: "binary install path overrides global",
			binary: Binary{
				Id:          "test",
				Name:        "test",
				Provider:    "github",
				Path:        "owner/repo",
				Format:      ".tar.gz",
				InstallPath: "/opt/bin",
			},
			global: GlobalConfig{
				InstallPath: "/usr/local/bin",
			},
			expected: Binary{
				Id:          "test",
				Name:        "test",
				Provider:    "github",
				Path:        "owner/repo",
				Format:      ".tar.gz",
				InstallPath: "/opt/bin",
			},
		},
		{
			name: "provider authenticated applies when binary not authenticated",
			binary: Binary{
				Id:            "test",
				Name:          "test",
				Provider:      "github",
				Path:          "owner/repo",
				Format:        ".tar.gz",
				Authenticated: false,
			},
			global: GlobalConfig{
				Providers: map[string]ProviderDefaults{
					"github": {
						Authenticated: true,
					},
				},
			},
			expected: Binary{
				Id:            "test",
				Name:          "test",
				Provider:      "github",
				Path:          "owner/repo",
				Format:        ".tar.gz",
				Authenticated: true,
			},
		},
		{
			name: "binary authenticated stays true when provider default is true",
			binary: Binary{
				Id:            "test",
				Name:          "test",
				Provider:      "github",
				Path:          "owner/repo",
				Format:        ".tar.gz",
				Authenticated: true,
			},
			global: GlobalConfig{
				Providers: map[string]ProviderDefaults{
					"github": {
						Authenticated: true,
					},
				},
			},
			expected: Binary{
				Id:            "test",
				Name:          "test",
				Provider:      "github",
				Path:          "owner/repo",
				Format:        ".tar.gz",
				Authenticated: true,
			},
		},
		{
			name: "no provider config means no change",
			binary: Binary{
				Id:            "test",
				Name:          "test",
				Provider:      "gitlab",
				Path:          "owner/repo",
				Format:        ".tar.gz",
				Authenticated: false,
			},
			global: GlobalConfig{
				Providers: map[string]ProviderDefaults{
					"github": {
						Authenticated: true,
					},
				},
			},
			expected: Binary{
				Id:            "test",
				Name:          "test",
				Provider:      "gitlab",
				Path:          "owner/repo",
				Format:        ".tar.gz",
				Authenticated: false,
			},
		},
		{
			name: "all global settings apply",
			binary: Binary{
				Id:       "test",
				Name:     "test",
				Provider: "github",
				Path:     "owner/repo",
				Format:   ".tar.gz",
			},
			global: GlobalConfig{
				InstallPath: "/usr/local/bin",
				Providers: map[string]ProviderDefaults{
					"github": {
						Authenticated: true,
					},
				},
			},
			expected: Binary{
				Id:            "test",
				Name:          "test",
				Provider:      "github",
				Path:          "owner/repo",
				Format:        ".tar.gz",
				InstallPath:   "/usr/local/bin",
				Authenticated: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeBinaryWithGlobal(tt.binary, tt.global)

			if result.InstallPath != tt.expected.InstallPath {
				t.Errorf("InstallPath = %v, expected %v", result.InstallPath, tt.expected.InstallPath)
			}
			if result.Authenticated != tt.expected.Authenticated {
				t.Errorf("Authenticated = %v, expected %v", result.Authenticated, tt.expected.Authenticated)
			}
		})
	}
}
