//go:build !android

package e2e_grpc_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestScenario_GPSCommandsExist(t *testing.T) {
	assertCommandExists(t, "location", "location-manager", "get-last-known-location")
	assertCommandExists(t, "location", "location-manager", "is-provider-enabled")
	assertCommandExists(t, "location", "location-manager", "request-location-updates-raw")
}

func TestScenario_WiFiCommandsExist(t *testing.T) {
	assertCommandExists(t, "wifi", "wifi-manager", "is-enabled")
	assertCommandExists(t, "wifi", "wifi-manager", "get-connection-info-raw")
	assertCommandExists(t, "wifi", "wifi-manager", "get-scan-results-raw")
}

func TestScenario_CameraCommandsExist(t *testing.T) {
	assertCommandExists(t, "camera", "camera-manager", "set-torch-mode")
}

func TestScenario_RecorderCommandsExist(t *testing.T) {
	assertCommandExists(t, "recorder", "media-recorder", "set-audio-source")
	assertCommandExists(t, "recorder", "media-recorder", "prepare")
	assertCommandExists(t, "recorder", "media-recorder", "start")
	assertCommandExists(t, "recorder", "media-recorder", "stop")
	assertCommandExists(t, "recorder", "media-recorder", "release")
}

func TestScenario_NotificationCommandsExist(t *testing.T) {
	assertCommandExists(t, "notification", "notification-manager", "are-notifications-enabled")
	assertCommandExists(t, "notification", "notification-manager", "create-notification-channel")
	assertCommandExists(t, "notification", "notification-manager", "notify-raw")
	assertCommandExists(t, "notification", "notification-manager", "cancel")
}

func TestScenario_DeviceInfoCommandsExist(t *testing.T) {
	assertCommandExists(t, "build", "build", "get-manufacturer")
	assertCommandExists(t, "build", "build", "get-model")
	assertCommandExists(t, "build", "build", "get-sdk-int")
}

func TestScenario_BluetoothCommandsExist(t *testing.T) {
	assertCommandExists(t, "bluetooth", "bluetooth-adapter", "is-enabled")
	assertCommandExists(t, "bluetooth", "bluetooth-adapter", "get-name")
	assertCommandExists(t, "bluetooth", "bluetooth-adapter", "get-address")
}

func TestScenario_RawJNICommandsExist(t *testing.T) {
	assertCommandExists(t, "jni", "class", "find")
	assertCommandExists(t, "jni", "method", "get-static-id")
	assertCommandExists(t, "jni", "method", "call-static")
	assertCommandExists(t, "jni", "string", "new")
	assertCommandExists(t, "jni", "string", "get")
}

// assertCommandExists verifies a jnicli subcommand exists by checking its help output.
func assertCommandExists(t *testing.T, args ...string) {
	t.Helper()
	fullArgs := append(args, "--help")
	cmd := exec.Command("go", append([]string{"run", jnicliBin}, fullArgs...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("command jnicli %s: %v\n%s", strings.Join(args, " "), err, out)
		return
	}
	// Verify it's a real command (has "Usage:" section), not just a group
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("command jnicli %s produced no Usage section", strings.Join(args, " "))
	}
}
