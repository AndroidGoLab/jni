//go:build !android

package e2e_grpc_test

import (
	"os/exec"
	"strings"
	"testing"
)

var jnicliBin = "../../cmd/jnicli"

// TestJnicliHelp verifies the root command exists and lists expected subcommands.
func TestJnicliHelp(t *testing.T) {
	out := runJnicliHelp(t)
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

// TestJnicliCommandCount verifies the expected number of leaf commands.
func TestJnicliCommandCount(t *testing.T) {
	cmd := exec.Command("go", "run", jnicliBin, "list-commands")
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

func runJnicliHelp(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("go", "run", jnicliBin, "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jnicli --help: %v\n%s", err, out)
	}
	return string(out)
}
