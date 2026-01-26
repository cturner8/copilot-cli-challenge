package github

import (
	"testing"
)

func TestFilterAssets(t *testing.T) {
	// Sample test assets
	assets := []ReleaseAsset{
		{Id: 1, Name: "binary-linux-amd64.tar.gz"},
		{Id: 2, Name: "binary-darwin-arm64.zip"},
		{Id: 3, Name: "binary-windows-amd64.exe"},
		{Id: 4, Name: "binary-linux-arm64.tar.gz"},
		{Id: 5, Name: "checksums.txt"},
	}

	tests := []struct {
		name        string
		filter      AssetFilter
		wantCount   int
		wantAssetId int // ID of expected asset (0 if multiple expected)
		wantErr     bool
	}{
		{
			name: "filter by linux and amd64",
			filter: AssetFilter{
				OS:   "linux",
				Arch: "amd64",
			},
			wantCount:   1,
			wantAssetId: 1,
			wantErr:     false,
		},
		{
			name: "filter by darwin and arm64",
			filter: AssetFilter{
				OS:   "darwin",
				Arch: "arm64",
			},
			wantCount:   1,
			wantAssetId: 2,
			wantErr:     false,
		},
		{
			name: "filter by extension tar.gz",
			filter: AssetFilter{
				Extension: ".tar.gz",
			},
			wantCount:   2,
			wantAssetId: 0, // Multiple results
			wantErr:     false,
		},
		{
			name: "filter by linux, arm64, and tar.gz",
			filter: AssetFilter{
				OS:        "linux",
				Arch:      "arm64",
				Extension: ".tar.gz",
			},
			wantCount:   1,
			wantAssetId: 4,
			wantErr:     false,
		},
		{
			name: "filter with prefix",
			filter: AssetFilter{
				Prefix: "binary",
			},
			wantCount:   4,
			wantAssetId: 0,
			wantErr:     false,
		},
		{
			name: "no matches",
			filter: AssetFilter{
				OS:   "freebsd",
				Arch: "amd64",
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "filter with regex",
			filter: AssetFilter{
				AssetRegex: `binary-linux-.*\.tar\.gz`,
			},
			wantCount:   2,
			wantAssetId: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterAssets(assets, tt.filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("FilterAssets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("FilterAssets() got %d assets, want %d", len(result), tt.wantCount)
				return
			}

			if tt.wantAssetId != 0 && len(result) > 0 {
				if result[0].Id != tt.wantAssetId {
					t.Errorf("FilterAssets() got asset ID %d, want %d", result[0].Id, tt.wantAssetId)
				}
			}
		})
	}
}

func TestFilterByOS(t *testing.T) {
	assets := []ReleaseAsset{
		{Id: 1, Name: "app-linux-amd64"},
		{Id: 2, Name: "app-Linux-amd64"},
		{Id: 3, Name: "app-darwin-arm64"},
		{Id: 4, Name: "app-macOS-arm64"},
		{Id: 5, Name: "app-windows-amd64.exe"},
	}

	tests := []struct {
		name      string
		os        string
		wantCount int
	}{
		{"linux variations", "linux", 2},
		{"darwin/macOS variations", "darwin", 2},
		{"windows variations", "windows", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterByOS(assets, tt.os)
			if len(result) != tt.wantCount {
				t.Errorf("filterByOS() got %d assets, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestFilterByArch(t *testing.T) {
	assets := []ReleaseAsset{
		{Id: 1, Name: "app-amd64"},
		{Id: 2, Name: "app-x86_64"},
		{Id: 3, Name: "app-x64"},
		{Id: 4, Name: "app-arm64"},
		{Id: 5, Name: "app-aarch64"},
		{Id: 6, Name: "app-i386"},
		{Id: 7, Name: "app-i686"},
	}

	tests := []struct {
		name      string
		arch      string
		wantCount int
	}{
		{"amd64 variations", "amd64", 3},
		{"arm64 variations", "arm64", 2},
		{"386 variations", "386", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterByArch(assets, tt.arch)
			if len(result) != tt.wantCount {
				t.Errorf("filterByArch() got %d assets, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestFilterByExtension(t *testing.T) {
	assets := []ReleaseAsset{
		{Id: 1, Name: "app.tar.gz"},
		{Id: 2, Name: "app.tgz"},
		{Id: 3, Name: "app.zip"},
		{Id: 4, Name: "app.tar.xz"},
	}

	tests := []struct {
		name      string
		ext       string
		wantCount int
	}{
		{"tar.gz extension", ".tar.gz", 1},
		{"zip extension", ".zip", 1},
		{"tgz extension", ".tgz", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterByExtension(assets, tt.ext)
			if len(result) != tt.wantCount {
				t.Errorf("filterByExtension() got %d assets, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestSelectBestAsset(t *testing.T) {
	tests := []struct {
		name    string
		assets  []ReleaseAsset
		wantId  int
		wantErr bool
	}{
		{
			name:    "empty assets",
			assets:  []ReleaseAsset{},
			wantErr: true,
		},
		{
			name: "single asset",
			assets: []ReleaseAsset{
				{Id: 1, Name: "app.zip"},
			},
			wantId:  1,
			wantErr: false,
		},
		{
			name: "prefer tar.gz over zip",
			assets: []ReleaseAsset{
				{Id: 1, Name: "app.zip"},
				{Id: 2, Name: "app.tar.gz"},
			},
			wantId:  2,
			wantErr: false,
		},
		{
			name: "prefer tgz",
			assets: []ReleaseAsset{
				{Id: 1, Name: "app.bin"},
				{Id: 2, Name: "app.tgz"},
			},
			wantId:  2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SelectBestAsset(tt.assets)

			if (err != nil) != tt.wantErr {
				t.Errorf("SelectBestAsset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result.Id != tt.wantId {
				t.Errorf("SelectBestAsset() got asset ID %d, want %d", result.Id, tt.wantId)
			}
		})
	}
}

func TestNewAssetFilter(t *testing.T) {
	filter := NewAssetFilter()

	if filter.OS == "" {
		t.Error("NewAssetFilter() OS should not be empty")
	}

	if filter.Arch == "" {
		t.Error("NewAssetFilter() Arch should not be empty")
	}
}
