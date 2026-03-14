//go:build !android

package e2e_grpc_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// These tests require a running jniservice on an Android emulator.
// Set JNICTL_E2E_ADDR=<host:port> to enable them.
// Run with: JNICTL_E2E_ADDR=localhost:50051 go test -v -run TestEmulator

func skipIfNoEmulator(t *testing.T) {
	t.Helper()
	if os.Getenv("JNICTL_E2E_ADDR") == "" {
		t.Skip("JNICTL_E2E_ADDR not set; skipping emulator tests")
	}
}

func runLiveJnictl(t *testing.T, args ...string) string {
	t.Helper()
	addr := os.Getenv("JNICTL_E2E_ADDR")
	fullArgs := append([]string{"--addr", addr, "--insecure"}, args...)
	cmd := exec.Command("go", append([]string{"run", jnictlBin}, fullArgs...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jnictl %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func TestEmulator_PowerIsInteractive(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnictl(t, "power", "power-manager", "is-interactive")
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	t.Logf("power is-interactive: %s", out)
}

func TestEmulator_BuildInfo(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnictl(t, "build", "build", "get-manufacturer")
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	t.Logf("manufacturer: %s", out)
}

func TestEmulator_JNIGetVersion(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnictl(t, "jni", "get-version")
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if _, ok := resp["version"]; !ok {
		t.Error("missing 'version' field in response")
	}
	t.Logf("JNI version: %s", out)
}

func TestEmulator_LocationProviderEnabled(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnictl(t, "location", "location-manager", "is-provider-enabled", "--arg0", "gps")
	t.Logf("location provider enabled: %s", out)
}

func TestEmulator_WiFiEnabled(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnictl(t, "wifi", "wifi-manager", "is-enabled")
	t.Logf("wifi enabled: %s", out)
}

func TestEmulator_BluetoothEnabled(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnictl(t, "bluetooth", "bluetooth-adapter", "is-enabled")
	t.Logf("bluetooth enabled: %s", out)
}

func TestEmulator_NotificationsEnabled(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnictl(t, "notification", "notification-manager", "are-notifications-enabled")
	t.Logf("notifications enabled: %s", out)
}

func TestEmulator_KeyguardLocked(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnictl(t, "keyguard", "keyguard-manager", "is-device-locked")
	t.Logf("device locked: %s", out)
}

func TestEmulator_BatteryStatus(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnictl(t, "battery", "battery-manager", "get-status")
	t.Logf("battery status: %s", out)
}

func TestEmulator_RawJNI_FindClassAndCallMethod(t *testing.T) {
	skipIfNoEmulator(t)
	// Find java.lang.System class
	out := runLiveJnictl(t, "jni", "class", "find", "--name", "java/lang/System")
	var findResp map[string]interface{}
	if err := json.Unmarshal([]byte(out), &findResp); err != nil {
		t.Fatalf("FindClass: invalid JSON: %v\n%s", err, out)
	}
	classHandle, ok := findResp["classHandle"]
	if !ok {
		t.Fatal("FindClass: missing classHandle")
	}
	t.Logf("System class handle: %v", classHandle)
}
