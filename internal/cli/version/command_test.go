package versioncmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand_DefaultOutput(t *testing.T) {
	originalVersion, originalCommit, originalDate := BuildVersion, BuildCommit, BuildDate
	BuildVersion = "v1.2.3"
	BuildCommit = "abc123"
	BuildDate = "2026-02-14T11:00:00Z"
	t.Cleanup(func() {
		BuildVersion = originalVersion
		BuildCommit = originalCommit
		BuildDate = originalDate
	})

	cmd := NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	output := buf.String()
	if output != "binmate v1.2.3\n" {
		t.Fatalf("output = %q, want %q", output, "binmate v1.2.3\n")
	}
}

func TestVersionCommand_VerboseOutput(t *testing.T) {
	originalVersion, originalCommit, originalDate := BuildVersion, BuildCommit, BuildDate
	BuildVersion = "v1.2.3"
	BuildCommit = "abc123"
	BuildDate = "2026-02-14T11:00:00Z"
	t.Cleanup(func() {
		BuildVersion = originalVersion
		BuildCommit = originalCommit
		BuildDate = originalDate
	})

	cmd := NewCommand()
	cmd.SetArgs([]string{"--verbose"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "version: v1.2.3") {
		t.Fatalf("expected version in output, got: %s", output)
	}
	if !strings.Contains(output, "commit: abc123") {
		t.Fatalf("expected commit in output, got: %s", output)
	}
	if !strings.Contains(output, "date: 2026-02-14T11:00:00Z") {
		t.Fatalf("expected date in output, got: %s", output)
	}
}
