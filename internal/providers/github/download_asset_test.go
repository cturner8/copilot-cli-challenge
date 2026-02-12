package github

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadAsset_Success(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("binary-content"))
	}))
	defer srv.Close()

	path, err := DownloadAsset(srv.URL+"/asset", "tool.tar.gz", false)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	expectedPath := filepath.Join(cacheHome, "binmate", "tool.tar.gz")
	if path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, path)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(got) != "binary-content" {
		t.Fatalf("unexpected file content: %q", string(got))
	}
}

func TestDownloadAsset_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := DownloadAsset(srv.URL+"/missing", "tool.tar.gz", false)
	if err == nil || !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("expected status error, got: %v", err)
	}
}

func TestDownloadAsset_AuthenticatedRequiresToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	_, err := DownloadAsset("https://example.com/file", "tool.tar.gz", true)
	if err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Fatalf("expected missing token error, got: %v", err)
	}
}

func TestDownloadAsset_AuthenticatedAddsBearerToken(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)
	t.Setenv("GITHUB_TOKEN", "secret-token")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	path, err := DownloadAsset(srv.URL+"/asset", "auth-tool.tar.gz", true)
	if err != nil {
		t.Fatalf("expected success with auth, got error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected downloaded file to exist, stat error: %v", err)
	}
}
