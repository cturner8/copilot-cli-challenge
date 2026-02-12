package importcmd

import (
	"strings"
	"testing"
)

func TestImportCommand_ValidationAndErrors(t *testing.T) {
	t.Run("requires name flag", func(t *testing.T) {
		cmd := NewCommand()
		cmd.SetArgs([]string{"/usr/local/bin/gh"})

		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "required flag(s) \"name\" not set") {
			t.Fatalf("expected required name flag error, got: %v", err)
		}
	})

	t.Run("returns not implemented import error", func(t *testing.T) {
		cmd := NewCommand()
		cmd.SetArgs([]string{"/usr/local/bin/gh", "--name", "gh"})

		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "failed to import binary") {
			t.Fatalf("expected import failure, got: %v", err)
		}
		if !strings.Contains(err.Error(), "not yet implemented") {
			t.Fatalf("expected not implemented reason, got: %v", err)
		}
	})
}
