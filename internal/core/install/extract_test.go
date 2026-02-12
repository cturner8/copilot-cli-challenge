package install

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"cturner8/binmate/internal/database"
)

type tarFixtureEntry struct {
	name     string
	content  []byte
	mode     int64
	typeflag byte
	linkname string
}

func createTarGzFixture(t *testing.T, path string, entries []tarFixtureEntry) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create tar fixture: %v", err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	tw := tar.NewWriter(gzw)

	for _, entry := range entries {
		hdr := &tar.Header{
			Name:     entry.name,
			Mode:     entry.mode,
			Size:     int64(len(entry.content)),
			Typeflag: entry.typeflag,
			Linkname: entry.linkname,
		}
		if entry.typeflag == 0 {
			hdr.Typeflag = tar.TypeReg
		}
		if hdr.Typeflag != tar.TypeReg {
			hdr.Size = 0
		}

		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write tar header: %v", err)
		}
		if hdr.Typeflag == tar.TypeReg {
			if _, err := tw.Write(entry.content); err != nil {
				t.Fatalf("failed to write tar content: %v", err)
			}
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
}

type zipFixtureEntry struct {
	name    string
	content []byte
	mode    os.FileMode
}

func createZipFixture(t *testing.T, path string, entries []zipFixtureEntry) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create zip fixture: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for _, entry := range entries {
		hdr := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		hdr.SetMode(entry.mode)
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			t.Fatalf("failed to create zip header: %v", err)
		}
		if _, err := w.Write(entry.content); err != nil {
			t.Fatalf("failed to write zip content: %v", err)
		}
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
}

func createTarGzBytes(t *testing.T, binaryName string, content []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)
	hdr := &tar.Header{Name: binaryName, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
	return buf.Bytes()
}

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

func TestExtractTar(t *testing.T) {
	t.Run("extracts binary from nested path", func(t *testing.T) {
		tmp := t.TempDir()
		archivePath := filepath.Join(tmp, "test.tar.gz")
		createTarGzFixture(t, archivePath, []tarFixtureEntry{
			{name: "bin/testbin", content: []byte("tar-binary"), mode: 0o755},
		})

		destDir := filepath.Join(tmp, "dest")
		got, err := extractTar(archivePath, destDir, "testbin")
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if got != filepath.Join(destDir, "testbin") {
			t.Fatalf("unexpected extracted path: %s", got)
		}

		content, err := os.ReadFile(got)
		if err != nil {
			t.Fatalf("failed to read extracted file: %v", err)
		}
		if string(content) != "tar-binary" {
			t.Fatalf("unexpected extracted content: %q", content)
		}
	})

	t.Run("validates required arguments and archive shape", func(t *testing.T) {
		tmp := t.TempDir()
		archivePath := filepath.Join(tmp, "invalid.tar.gz")
		if err := os.WriteFile(archivePath, []byte("not-gzip"), 0o644); err != nil {
			t.Fatalf("failed to write invalid fixture: %v", err)
		}

		if _, err := extractTar(archivePath, "", "testbin"); err == nil || !strings.Contains(err.Error(), "destination directory is required") {
			t.Fatalf("expected destination error, got: %v", err)
		}
		if _, err := extractTar(archivePath, tmp, ""); err == nil || !strings.Contains(err.Error(), "binary name is required") {
			t.Fatalf("expected binary name error, got: %v", err)
		}
		if _, err := extractTar(archivePath, tmp, "testbin"); err == nil || !strings.Contains(err.Error(), "create gzip reader") {
			t.Fatalf("expected gzip error, got: %v", err)
		}
	})

	t.Run("returns not found when binary is absent", func(t *testing.T) {
		tmp := t.TempDir()
		archivePath := filepath.Join(tmp, "test.tar.gz")
		createTarGzFixture(t, archivePath, []tarFixtureEntry{
			{name: "bin/other", content: []byte("other"), mode: 0o755},
		})

		_, err := extractTar(archivePath, filepath.Join(tmp, "dest"), "testbin")
		if err == nil || !strings.Contains(err.Error(), "not found in archive") {
			t.Fatalf("expected not found error, got: %v", err)
		}
	})
}

func TestExtractTarBinary_RejectsSymlinks(t *testing.T) {
	err := extractTarBinary(nil, &tar.Header{
		Name:     "testbin",
		Typeflag: tar.TypeSymlink,
		Linkname: "/tmp/target",
	}, "/tmp/ignored")
	if err == nil || !strings.Contains(err.Error(), "symlinks are not supported") {
		t.Fatalf("expected symlink rejection, got: %v", err)
	}
}

func TestExtractZip(t *testing.T) {
	t.Run("extracts binary from nested path", func(t *testing.T) {
		tmp := t.TempDir()
		archivePath := filepath.Join(tmp, "test.zip")
		createZipFixture(t, archivePath, []zipFixtureEntry{
			{name: "bin/testbin", content: []byte("zip-binary"), mode: 0o755},
		})

		destDir := filepath.Join(tmp, "dest")
		got, err := extractZip(archivePath, destDir, "testbin")
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if got != filepath.Join(destDir, "testbin") {
			t.Fatalf("unexpected extracted path: %s", got)
		}

		content, err := os.ReadFile(got)
		if err != nil {
			t.Fatalf("failed to read extracted file: %v", err)
		}
		if string(content) != "zip-binary" {
			t.Fatalf("unexpected extracted content: %q", content)
		}
	})

	t.Run("validates required arguments and archive shape", func(t *testing.T) {
		tmp := t.TempDir()
		archivePath := filepath.Join(tmp, "invalid.zip")
		if err := os.WriteFile(archivePath, []byte("not-zip"), 0o644); err != nil {
			t.Fatalf("failed to write invalid fixture: %v", err)
		}

		if _, err := extractZip(archivePath, "", "testbin"); err == nil || !strings.Contains(err.Error(), "destination directory is required") {
			t.Fatalf("expected destination error, got: %v", err)
		}
		if _, err := extractZip(archivePath, tmp, ""); err == nil || !strings.Contains(err.Error(), "binary name is required") {
			t.Fatalf("expected binary name error, got: %v", err)
		}
		if _, err := extractZip(archivePath, tmp, "testbin"); err == nil || !strings.Contains(err.Error(), "open zip") {
			t.Fatalf("expected zip error, got: %v", err)
		}
	})

	t.Run("returns not found when binary is absent", func(t *testing.T) {
		tmp := t.TempDir()
		archivePath := filepath.Join(tmp, "test.zip")
		createZipFixture(t, archivePath, []zipFixtureEntry{
			{name: "bin/other", content: []byte("other"), mode: 0o755},
		})

		_, err := extractZip(archivePath, filepath.Join(tmp, "dest"), "testbin")
		if err == nil || !strings.Contains(err.Error(), "not found in archive") {
			t.Fatalf("expected not found error, got: %v", err)
		}
	})
}

func TestExtractZip_RejectsSymlinkAsset(t *testing.T) {
	tmp := t.TempDir()
	archivePath := filepath.Join(tmp, "symlink.zip")
	createZipFixture(t, archivePath, []zipFixtureEntry{
		{name: "testbin", content: []byte("link-target"), mode: os.ModeSymlink | 0o777},
	})

	_, err := extractZip(archivePath, filepath.Join(tmp, "dest"), "testbin")
	if err == nil || !strings.Contains(err.Error(), "symlinks are not supported") {
		t.Fatalf("expected symlink rejection, got: %v", err)
	}
}

func TestExtractPathHelpers(t *testing.T) {
	t.Run("getLocalDataDir uses XDG_DATA_HOME when set", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_DATA_HOME", tmp)
		got, err := getLocalDataDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != tmp {
			t.Fatalf("expected %q, got %q", tmp, got)
		}
	})

	t.Run("getLocalDataDir falls back to HOME on unix", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("unix-specific fallback path")
		}
		home := t.TempDir()
		orig, wasSet := os.LookupEnv("XDG_DATA_HOME")
		if err := os.Unsetenv("XDG_DATA_HOME"); err != nil {
			t.Fatalf("failed to unset XDG_DATA_HOME: %v", err)
		}
		t.Cleanup(func() {
			if wasSet {
				_ = os.Setenv("XDG_DATA_HOME", orig)
			} else {
				_ = os.Unsetenv("XDG_DATA_HOME")
			}
		})
		t.Setenv("HOME", home)
		got, err := getLocalDataDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(home, ".local", "share")
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("getExtractPath builds expected version path", func(t *testing.T) {
		base := t.TempDir()
		t.Setenv("XDG_DATA_HOME", base)
		got, err := getExtractPath("bin-id", "v1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(base, "binmate", "versions", "bin-id", "v1.0.0")
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

func TestExtractAsset(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))

	tarArchive := filepath.Join(tmp, "asset.tar.gz")
	createTarGzFixture(t, tarArchive, []tarFixtureEntry{
		{name: "bin/tool", content: []byte("tar-content"), mode: 0o755},
	})
	tarBinary := &database.Binary{UserID: "tool-id", Name: "tool", Format: ".tar.gz"}

	tarPath, err := ExtractAsset(tarArchive, tarBinary, "v1.0.0")
	if err != nil {
		t.Fatalf("expected tar extraction success, got: %v", err)
	}
	tarContent, err := os.ReadFile(tarPath)
	if err != nil {
		t.Fatalf("failed to read tar extracted binary: %v", err)
	}
	if string(tarContent) != "tar-content" {
		t.Fatalf("unexpected tar extracted content: %q", tarContent)
	}

	zipArchive := filepath.Join(tmp, "asset.zip")
	createZipFixture(t, zipArchive, []zipFixtureEntry{
		{name: "bin/tool", content: []byte("zip-content"), mode: 0o755},
	})
	zipBinary := &database.Binary{UserID: "tool-id", Name: "tool", Format: ".zip"}

	zipPath, err := ExtractAsset(zipArchive, zipBinary, "v2.0.0")
	if err != nil {
		t.Fatalf("expected zip extraction success, got: %v", err)
	}
	zipContent, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("failed to read zip extracted binary: %v", err)
	}
	if string(zipContent) != "zip-content" {
		t.Fatalf("unexpected zip extracted content: %q", zipContent)
	}

	unknownBinary := &database.Binary{UserID: "tool-id", Name: "tool", Format: ".unknown"}
	if _, err := ExtractAsset(zipArchive, unknownBinary, "v3.0.0"); err == nil || !strings.Contains(err.Error(), "unsupported asset format") {
		t.Fatalf("expected unsupported format error, got: %v", err)
	}
}

func TestInstallBinary_SuccessAndAlreadyInstalledPath(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	binary := createTestBinary(t, dbService, "test")
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "cache"))
	t.Setenv("HOME", filepath.Join(tmp, "home"))

	assetName := fmt.Sprintf("testbin-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	archive := createTarGzBytes(t, "testbin", []byte("#!/bin/sh\necho ok\n"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/releases/latest":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"tag_name":"v1.2.3","assets":[{"id":1,"name":"%s","browser_download_url":"https://api.github.com/download/%s"}]}`, assetName, assetName)
		case "/download/" + assetName:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	routeGitHubAPIToServer(t, server)

	result, err := InstallBinary("test", "latest", dbService)
	if err != nil {
		t.Fatalf("expected successful install, got: %v", err)
	}
	if result.Version != "v1.2.3" {
		t.Fatalf("expected resolved version v1.2.3, got %s", result.Version)
	}
	if _, err := os.Stat(result.Installation.InstalledPath); err != nil {
		t.Fatalf("expected extracted binary to exist: %v", err)
	}

	activeVersion, err := dbService.Versions.Get(binary.ID)
	if err != nil {
		t.Fatalf("expected active version record: %v", err)
	}
	target, err := os.Readlink(activeVersion.SymlinkPath)
	if err != nil {
		t.Fatalf("failed reading symlink: %v", err)
	}
	if target != result.Installation.InstalledPath {
		t.Fatalf("expected symlink target %q, got %q", result.Installation.InstalledPath, target)
	}

	// Installing latest again should take the "already installed" branch.
	again, err := InstallBinary("test", "latest", dbService)
	if err != nil {
		t.Fatalf("expected successful reinstall on existing version, got: %v", err)
	}
	if again.Installation.ID != result.Installation.ID {
		t.Fatalf("expected existing installation to be reused, got %d != %d", again.Installation.ID, result.Installation.ID)
	}

	installations, err := dbService.Installations.ListByBinary(binary.ID)
	if err != nil {
		t.Fatalf("failed to list installations: %v", err)
	}
	if len(installations) != 1 {
		t.Fatalf("expected exactly one installation record, got %d", len(installations))
	}
}

func TestInstallBinary_ChecksumVerificationFailure(t *testing.T) {
	dbService, cleanup := setupTestDB(t)
	defer cleanup()

	createTestBinary(t, dbService, "test")
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "cache"))
	t.Setenv("HOME", filepath.Join(tmp, "home"))

	assetName := fmt.Sprintf("testbin-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	archive := createTarGzBytes(t, "testbin", []byte("not-the-expected-checksum"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/releases/latest":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"tag_name":"v1.2.3","assets":[{"id":1,"name":"%s","digest":"sha256:deadbeef","browser_download_url":"https://api.github.com/download/%s"}]}`, assetName, assetName)
		case "/download/" + assetName:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	routeGitHubAPIToServer(t, server)

	_, err := InstallBinary("test", "latest", dbService)
	if err == nil || !strings.Contains(err.Error(), "checksum verification failed") {
		t.Fatalf("expected checksum verification error, got: %v", err)
	}
}
