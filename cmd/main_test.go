package cmd

import (
	"bytes"
	"testing"
)

func TestRootVersionFlag_Long(t *testing.T) {
	SetBuildMetadata("v1.2.3", "abc123", "2026-02-14T11:00:00Z")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if got := buf.String(); got != "binmate v1.2.3\n" {
		t.Fatalf("output = %q, want %q", got, "binmate v1.2.3\n")
	}
}

func TestRootVersionFlag_Short(t *testing.T) {
	SetBuildMetadata("v1.2.3", "abc123", "2026-02-14T11:00:00Z")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"-v"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if got := buf.String(); got != "binmate v1.2.3\n" {
		t.Fatalf("output = %q, want %q", got, "binmate v1.2.3\n")
	}
}
