package configcmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"cturner8/binmate/internal/core/config"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	_ = w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read captured stdout: %v", err)
	}
	return buf.String()
}

func TestConfigCommand_OutputModes(t *testing.T) {
	Config = &config.Config{
		Version: 1,
		Binaries: []config.Binary{
			{Id: "gh", Name: "GitHub CLI", Provider: "github", Path: "cli/cli"},
		},
	}

	t.Run("json output", func(t *testing.T) {
		cmd := NewCommand()
		cmd.SetArgs([]string{"--json"})

		out := captureStdout(t, func() {
			if err := cmd.Execute(); err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
		})

		if !strings.Contains(out, `"Id": "gh"`) {
			t.Fatalf("expected JSON binary ID in output, got: %s", out)
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewCommand()
		cmd.SetArgs([]string{})

		out := captureStdout(t, func() {
			if err := cmd.Execute(); err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
		})

		if !strings.Contains(out, "binmate Configuration") || !strings.Contains(out, "gh") {
			t.Fatalf("expected table output in stdout, got: %s", out)
		}
	})
}
