package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"
	"testing"

	"cturner8/binmate/internal/database"
)

type githubRewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (t *githubRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.target.Scheme
	clone.URL.Host = t.target.Host
	return t.base.RoundTrip(clone)
}

func routeGitHubAPIToServer(t *testing.T, srv *httptest.Server) {
	t.Helper()
	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}

	original := http.DefaultTransport
	http.DefaultTransport = &githubRewriteTransport{
		target: target,
		base:   srv.Client().Transport,
	}

	t.Cleanup(func() {
		http.DefaultTransport = original
	})
}

func makeTestBinary() *database.Binary {
	return &database.Binary{
		ProviderPath: "owner/repo",
		Format:       ".tar.gz",
	}
}

func TestFetchReleaseAsset_LatestSuccess(t *testing.T) {
	assetName := "tool-" + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases/latest" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v1.2.3",
			Assets: []ReleaseAsset{
				{Id: 101, Name: assetName, BrowserDownloadUrl: "https://example.com/tool.tar.gz"},
				{Id: 102, Name: "tool-" + runtime.GOOS + "-" + runtime.GOARCH + ".zip", BrowserDownloadUrl: "https://example.com/tool.zip"},
			},
		})
	}))
	defer srv.Close()

	routeGitHubAPIToServer(t, srv)

	release, asset, err := FetchReleaseAsset(makeTestBinary(), "latest")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if release.TagName != "v1.2.3" {
		t.Fatalf("expected tag v1.2.3, got %q", release.TagName)
	}
	if asset.Id != 101 {
		t.Fatalf("expected asset 101, got %d", asset.Id)
	}
}

func TestFetchReleaseAsset_ReleaseRegexPrefixesVersion(t *testing.T) {
	assetName := "tool-" + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases/tags/v1.2.3" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v1.2.3",
			Assets: []ReleaseAsset{
				{Id: 201, Name: assetName, BrowserDownloadUrl: "https://example.com/tool.tar.gz"},
			},
		})
	}))
	defer srv.Close()

	routeGitHubAPIToServer(t, srv)

	releaseRegex := "v"
	binary := makeTestBinary()
	binary.ReleaseRegex = &releaseRegex

	_, asset, err := FetchReleaseAsset(binary, "1.2.3")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if asset.Id != 201 {
		t.Fatalf("expected asset 201, got %d", asset.Id)
	}
}

func TestFetchReleaseAsset_AuthenticatedRequiresToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	binary := makeTestBinary()
	binary.Authenticated = true

	_, _, err := FetchReleaseAsset(binary, "latest")
	if err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Fatalf("expected missing token error, got: %v", err)
	}
}

func TestFetchReleaseAsset_Errors(t *testing.T) {
	assetName := "tool-" + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz"
	tests := []struct {
		name        string
		status      int
		contentType string
		body        string
		binary      *database.Binary
		version     string
		wantErr     string
	}{
		{
			name:        "invalid release regex",
			binary:      func() *database.Binary { b := makeTestBinary(); rr := "["; b.ReleaseRegex = &rr; return b }(),
			version:     "1.0.0",
			wantErr:     "invalid releaseRegex pattern",
			status:      http.StatusOK,
			contentType: "application/json",
			body:        `{}`,
		},
		{
			name:        "non 200 status",
			status:      http.StatusTooManyRequests,
			contentType: "application/json",
			body:        `{}`,
			binary:      makeTestBinary(),
			version:     "latest",
			wantErr:     "unexpected status",
		},
		{
			name:        "invalid content type",
			status:      http.StatusOK,
			contentType: "text/plain",
			body:        `{}`,
			binary:      makeTestBinary(),
			version:     "latest",
			wantErr:     "invalid release response content",
		},
		{
			name:        "invalid json",
			status:      http.StatusOK,
			contentType: "application/json",
			body:        `{`,
			binary:      makeTestBinary(),
			version:     "latest",
			wantErr:     "failed to parse JSON",
		},
		{
			name:        "no assets",
			status:      http.StatusOK,
			contentType: "application/json",
			body:        `{"tag_name":"v1.0.0","assets":[]}`,
			binary:      makeTestBinary(),
			version:     "latest",
			wantErr:     "no release assets",
		},
		{
			name:        "no matching assets",
			status:      http.StatusOK,
			contentType: "application/json",
			body:        `{"tag_name":"v1.0.0","assets":[{"id":1,"name":"` + assetName + `.zip","browser_download_url":"https://example.com/tool.zip"}]}`,
			binary:      makeTestBinary(),
			version:     "latest",
			wantErr:     "no matching assets found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			routeGitHubAPIToServer(t, srv)

			_, _, err := FetchReleaseAsset(tt.binary, tt.version)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}
