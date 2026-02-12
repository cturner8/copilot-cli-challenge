package integration

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	binarySvc "cturner8/binmate/internal/core/binary"
	configPkg "cturner8/binmate/internal/core/config"
	installSvc "cturner8/binmate/internal/core/install"
	versionSvc "cturner8/binmate/internal/core/version"
	"cturner8/binmate/internal/database"
	"cturner8/binmate/internal/database/repository"
)

func setupIntegrationDB(t *testing.T) (*repository.Service, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "integration.db")
	db, err := database.Initialize(dbPath)
	if err != nil {
		t.Fatalf("failed to initialise database: %v", err)
	}

	svc := repository.NewService(db)
	cleanup := func() {
		_ = svc.Close()
	}
	return svc, cleanup
}

func createTarArchive(t *testing.T, fileName string, content []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)
	hdr := &tar.Header{Name: fileName, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}
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

type githubTransportRewrite struct {
	target *url.URL
	base   http.RoundTripper
}

func (t *githubTransportRewrite) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.target.Scheme
	clone.URL.Host = t.target.Host
	return t.base.RoundTrip(clone)
}

func routeGitHubTrafficToServer(t *testing.T, srv *httptest.Server) {
	t.Helper()
	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}

	original := http.DefaultTransport
	http.DefaultTransport = &githubTransportRewrite{
		target: target,
		base:   srv.Client().Transport,
	}
	t.Cleanup(func() {
		http.DefaultTransport = original
	})
}

func TestInstallUpdateAndSwitchWorkflow(t *testing.T) {
	svc, cleanup := setupIntegrationDB(t)
	defer cleanup()

	tmp := t.TempDir()
	t.Setenv("HOME", filepath.Join(tmp, "home"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "cache"))

	cfg := configPkg.Config{
		Version: 1,
		Binaries: []configPkg.Binary{
			{
				Id:       "workflow-bin",
				Name:     "wfbin",
				Provider: "github",
				Path:     "owner/repo",
				Format:   ".tar.gz",
			},
		},
	}
	if err := configPkg.SyncToDatabase(cfg, svc); err != nil {
		t.Fatalf("failed to sync initial config: %v", err)
	}

	archiveByTag := map[string][]byte{
		"v1.0.0": createTarArchive(t, "wfbin", []byte("binary-v1")),
		"v1.1.0": createTarArchive(t, "wfbin", []byte("binary-v2")),
	}
	latestTag := "v1.0.0"
	assetNameForTag := func(tag string) string {
		return fmt.Sprintf("wfbin-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, strings.TrimPrefix(tag, "v"))
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/owner/repo/releases/latest":
			assetName := assetNameForTag(latestTag)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"tag_name":"%s","assets":[{"id":1,"name":"%s","browser_download_url":"https://api.github.com/download/%s"}]}`, latestTag, assetName, assetName)
		case strings.HasPrefix(r.URL.Path, "/download/"):
			assetName := strings.TrimPrefix(r.URL.Path, "/download/")
			var tag string
			for candidate := range archiveByTag {
				if assetNameForTag(candidate) == assetName {
					tag = candidate
					break
				}
			}
			if tag == "" {
				http.NotFound(w, r)
				return
			}
			_, _ = w.Write(archiveByTag[tag])
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	routeGitHubTrafficToServer(t, srv)

	installed, err := installSvc.InstallBinary("workflow-bin", "latest", svc)
	if err != nil {
		t.Fatalf("install workflow failed: %v", err)
	}
	if installed.Version != "v1.0.0" {
		t.Fatalf("expected installed version v1.0.0, got %s", installed.Version)
	}

	latestTag = "v1.1.0"
	updated, err := installSvc.UpdateToLatest("workflow-bin", svc)
	if err != nil {
		t.Fatalf("update workflow failed: %v", err)
	}
	if updated.Version != "v1.1.0" {
		t.Fatalf("expected updated version v1.1.0, got %s", updated.Version)
	}

	if err := versionSvc.SwitchVersion("workflow-bin", "v1.0.0", svc); err != nil {
		t.Fatalf("switch workflow failed: %v", err)
	}

	bin, err := svc.Binaries.GetByUserID("workflow-bin")
	if err != nil {
		t.Fatalf("failed to load binary record: %v", err)
	}
	activeVersion, err := svc.Versions.Get(bin.ID)
	if err != nil {
		t.Fatalf("failed to load active version record: %v", err)
	}
	v1Install, err := svc.Installations.Get(bin.ID, "v1.0.0")
	if err != nil {
		t.Fatalf("failed to load v1 installation: %v", err)
	}
	if activeVersion.InstallationID != v1Install.ID {
		t.Fatalf("expected active installation ID %d, got %d", v1Install.ID, activeVersion.InstallationID)
	}
}

func TestRemoveWorkflowWithFiles(t *testing.T) {
	svc, cleanup := setupIntegrationDB(t)
	defer cleanup()

	tmp := t.TempDir()
	binary := &database.Binary{
		UserID:       "remove-bin",
		Name:         "removebin",
		Provider:     "github",
		ProviderPath: "owner/repo",
		Format:       ".tar.gz",
		Source:       "manual",
	}
	if err := svc.Binaries.Create(binary); err != nil {
		t.Fatalf("failed to create binary: %v", err)
	}

	installedPath := filepath.Join(tmp, "versions", "removebin", "v1.0.0", "removebin")
	if err := os.MkdirAll(filepath.Dir(installedPath), 0o755); err != nil {
		t.Fatalf("failed to create installation directory: %v", err)
	}
	if err := os.WriteFile(installedPath, []byte("installed"), 0o755); err != nil {
		t.Fatalf("failed to create installation file: %v", err)
	}

	installation := &database.Installation{
		BinaryID:          binary.ID,
		Version:           "v1.0.0",
		InstalledPath:     installedPath,
		SourceURL:         "https://example.com/removebin.tar.gz",
		FileSize:          9,
		Checksum:          "deadbeef",
		ChecksumAlgorithm: "SHA256",
	}
	if err := svc.Installations.Create(installation); err != nil {
		t.Fatalf("failed to create installation: %v", err)
	}

	symlinkPath := filepath.Join(tmp, "bin", "removebin")
	if err := os.MkdirAll(filepath.Dir(symlinkPath), 0o755); err != nil {
		t.Fatalf("failed to create symlink directory: %v", err)
	}
	if err := os.Symlink(installedPath, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}
	if err := svc.Versions.Set(binary.ID, installation.ID, symlinkPath); err != nil {
		t.Fatalf("failed to set active version: %v", err)
	}

	if err := binarySvc.RemoveBinary("remove-bin", svc, true); err != nil {
		t.Fatalf("remove workflow failed: %v", err)
	}

	if _, err := svc.Binaries.GetByUserID("remove-bin"); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected binary to be removed from DB, got: %v", err)
	}
	if _, err := os.Lstat(symlinkPath); !os.IsNotExist(err) {
		t.Fatalf("expected symlink to be removed, got stat error: %v", err)
	}
	if _, err := os.Stat(installedPath); !os.IsNotExist(err) {
		t.Fatalf("expected installation file to be removed, got stat error: %v", err)
	}
}

func TestConfigSyncWorkflowPreservesManualBinaries(t *testing.T) {
	svc, cleanup := setupIntegrationDB(t)
	defer cleanup()

	cfg := configPkg.Config{
		Version: 1,
		Global: configPkg.GlobalConfig{
			InstallPath: "/usr/local/bin",
		},
		Binaries: []configPkg.Binary{
			{
				Id:       "cfg-bin",
				Name:     "cfgbin",
				Provider: "github",
				Path:     "owner/repo",
				Format:   ".zip",
			},
		},
	}
	if err := configPkg.SyncToDatabase(cfg, svc); err != nil {
		t.Fatalf("failed to sync config: %v", err)
	}

	cfgBinary, err := svc.Binaries.GetByUserID("cfg-bin")
	if err != nil {
		t.Fatalf("expected config binary to exist: %v", err)
	}
	if cfgBinary.Source != "config" {
		t.Fatalf("expected config-managed source, got %q", cfgBinary.Source)
	}

	manual := &database.Binary{
		UserID:       "manual-bin",
		Name:         "manualbin",
		Provider:     "github",
		ProviderPath: "owner/other",
		Format:       ".tar.gz",
		Source:       "manual",
	}
	if err := svc.Binaries.Create(manual); err != nil {
		t.Fatalf("failed to create manual binary: %v", err)
	}

	cfg.Version = 2
	cfg.Binaries = nil
	if err := configPkg.SyncToDatabase(cfg, svc); err != nil {
		t.Fatalf("failed to sync updated config: %v", err)
	}

	if _, err := svc.Binaries.GetByUserID("cfg-bin"); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected config binary to be removed, got: %v", err)
	}
	if _, err := svc.Binaries.GetByUserID("manual-bin"); err != nil {
		t.Fatalf("expected manual binary to be preserved, got: %v", err)
	}
}
