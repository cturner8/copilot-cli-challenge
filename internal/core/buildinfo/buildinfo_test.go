package buildinfo

import (
	"runtime/debug"
	"testing"
)

func TestResolve_UsesLinkerValues(t *testing.T) {
	originalReader := readBuildInfo
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v9.9.9"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "fallback-commit"},
				{Key: "vcs.time", Value: "fallback-date"},
				{Key: "vcs.modified", Value: "true"},
			},
		}, true
	}
	defer func() { readBuildInfo = originalReader }()

	got := Resolve("v1.2.3", "abc123", "2026-02-14T00:00:00Z")

	if got.Version != "v1.2.3" {
		t.Fatalf("version = %q, want %q", got.Version, "v1.2.3")
	}
	if got.Commit != "abc123" {
		t.Fatalf("commit = %q, want %q", got.Commit, "abc123")
	}
	if got.Date != "2026-02-14T00:00:00Z" {
		t.Fatalf("date = %q, want %q", got.Date, "2026-02-14T00:00:00Z")
	}
}

func TestResolve_FallsBackToRuntimeBuildInfo(t *testing.T) {
	originalReader := readBuildInfo
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v2.0.1"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "deadbeef"},
				{Key: "vcs.time", Value: "2026-02-14T11:00:00Z"},
				{Key: "vcs.modified", Value: "true"},
			},
		}, true
	}
	defer func() { readBuildInfo = originalReader }()

	got := Resolve("dev", "unknown", "unknown")

	if got.Version != "v2.0.1" {
		t.Fatalf("version = %q, want %q", got.Version, "v2.0.1")
	}
	if got.Commit != "deadbeef" {
		t.Fatalf("commit = %q, want %q", got.Commit, "deadbeef")
	}
	if got.Date != "2026-02-14T11:00:00Z" {
		t.Fatalf("date = %q, want %q", got.Date, "2026-02-14T11:00:00Z")
	}
	if !got.Modified {
		t.Fatal("modified = false, want true")
	}
}

func TestResolve_UsesDefaultsWhenMissing(t *testing.T) {
	originalReader := readBuildInfo
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return nil, false
	}
	defer func() { readBuildInfo = originalReader }()

	got := Resolve("", "", "")

	if got.Version != DefaultVersion {
		t.Fatalf("version = %q, want %q", got.Version, DefaultVersion)
	}
	if got.Commit != DefaultUnknown {
		t.Fatalf("commit = %q, want %q", got.Commit, DefaultUnknown)
	}
	if got.Date != DefaultUnknown {
		t.Fatalf("date = %q, want %q", got.Date, DefaultUnknown)
	}
	if got.Modified {
		t.Fatal("modified = true, want false")
	}
}
