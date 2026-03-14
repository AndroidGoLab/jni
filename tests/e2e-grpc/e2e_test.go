//go:build !android

package e2e_grpc_test

import (
	"os/exec"
	"strings"
	"testing"
)

var jnictlBin = "../../cmd/jnictl"

// TestJnictlHelp verifies the root command exists and lists expected subcommands.
func TestJnictlHelp(t *testing.T) {
	out := runJnictlHelp(t)
	requiredCommands := []string{
		"alarm", "bluetooth", "camera", "location", "notification",
		"power", "vibrator", "wifi", "jni", "handle",
	}
	for _, cmd := range requiredCommands {
		if !strings.Contains(out, cmd) {
			t.Errorf("missing subcommand %q in help output", cmd)
		}
	}
}

// TestJnictlCommandCount verifies the expected number of leaf commands.
func TestJnictlCommandCount(t *testing.T) {
	cmd := exec.Command("go", "run", jnictlBin, "list-commands")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list-commands: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 1800 {
		t.Errorf("expected >= 1800 leaf commands, got %d", len(lines))
	}
	t.Logf("total leaf commands: %d", len(lines))
}

func runJnictlHelp(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("go", "run", jnictlBin, "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jnictl --help: %v\n%s", err, out)
	}
	return string(out)
}
