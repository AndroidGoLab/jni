//go:build !android

package e2e_grpc_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func skipIfNoMTLS(t *testing.T) {
	t.Helper()
	if os.Getenv("JNICTL_E2E_ADDR") == "" {
		t.Skip("JNICTL_E2E_ADDR not set")
	}
	if os.Getenv("JNICTL_E2E_MTLS") != "1" {
		t.Skip("JNICTL_E2E_MTLS not set; skipping mTLS tests")
	}
}

// runJnicliAuth runs jnicli with mTLS cert flags against the live server.
func runJnicliAuth(t *testing.T, certDir string, args ...string) string {
	t.Helper()
	addr := os.Getenv("JNICTL_E2E_ADDR")
	fullArgs := []string{
		"--addr", addr,
		"--insecure",
		"--cert", filepath.Join(certDir, "client.crt"),
		"--key", filepath.Join(certDir, "client.key"),
		"--ca", filepath.Join(certDir, "ca.crt"),
	}
	fullArgs = append(fullArgs, args...)

	cmd := jnicliCommand(fullArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jnicli %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// runJnicliAuthExpectError runs jnicli with mTLS and expects failure.
func runJnicliAuthExpectError(t *testing.T, certDir string, args ...string) string {
	t.Helper()
	addr := os.Getenv("JNICTL_E2E_ADDR")
	fullArgs := []string{
		"--addr", addr,
		"--insecure",
		"--cert", filepath.Join(certDir, "client.crt"),
		"--key", filepath.Join(certDir, "client.key"),
		"--ca", filepath.Join(certDir, "ca.crt"),
	}
	fullArgs = append(fullArgs, args...)

	cmd := jnicliCommand(fullArgs...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected error from jnicli %s but got success:\n%s", strings.Join(args, " "), out)
	}
	return string(out)
}

// runAdmin runs jniserviceadmin with the test DB.
func runAdmin(t *testing.T, dbPath string, args ...string) string {
	t.Helper()
	adminBin := os.Getenv("JNISERVICEADMIN_BIN")
	if adminBin == "" {
		t.Skip("JNISERVICEADMIN_BIN not set")
	}
	fullArgs := append([]string{"--db", dbPath}, args...)
	cmd := exec.Command(adminBin, fullArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jniserviceadmin %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func TestE2E_Auth_RegisterAndDenied(t *testing.T) {
	skipIfNoMTLS(t)
	certDir := t.TempDir()

	// 1. Register a new client.
	addr := os.Getenv("JNICTL_E2E_ADDR")
	cmd := jnicliCommand(
		"--addr", addr, "--insecure",
		"auth", "register",
		"--cn", "test-e2e-client",
		"--key-out", filepath.Join(certDir, "client.key"),
		"--cert-out", filepath.Join(certDir, "client.crt"),
		"--ca-out", filepath.Join(certDir, "ca.crt"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("register: %v\n%s", err, out)
	}
	t.Logf("Registered: %s", strings.TrimSpace(string(out)))

	// Verify cert files exist.
	for _, f := range []string{"client.key", "client.crt", "ca.crt"} {
		if _, err := os.Stat(filepath.Join(certDir, f)); err != nil {
			t.Fatalf("missing %s: %v", f, err)
		}
	}

	// 2. Try to call FindClass — should be denied (no grants).
	errOut := runJnicliAuthExpectError(t, certDir, "jni", "class", "find", "--name", "java/lang/String")
	if !strings.Contains(errOut, "PermissionDenied") {
		t.Errorf("expected PermissionDenied, got: %s", errOut)
	}
	t.Log("FindClass correctly denied without grant")
}

func TestE2E_Auth_GrantAndAllow(t *testing.T) {
	skipIfNoMTLS(t)

	dbPath := os.Getenv("JNISERVICE_DB")
	if dbPath == "" {
		dbPath = "/data/local/tmp/jniservice.db"
	}
	adminBin := os.Getenv("JNISERVICEADMIN_BIN")
	if adminBin == "" {
		t.Skip("JNISERVICEADMIN_BIN not set")
	}

	certDir := t.TempDir()

	// 1. Register.
	addr := os.Getenv("JNICTL_E2E_ADDR")
	cmd := jnicliCommand(
		"--addr", addr, "--insecure",
		"auth", "register",
		"--cn", "test-grant-client",
		"--key-out", filepath.Join(certDir, "client.key"),
		"--cert-out", filepath.Join(certDir, "client.crt"),
		"--ca-out", filepath.Join(certDir, "ca.crt"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("register: %v\n%s", err, out)
	}

	// 2. Admin grants access.
	runAdmin(t, dbPath, "grants", "approve", "test-grant-client", "/jni_raw.JNIService/*")
	t.Log("Admin granted /jni_raw.JNIService/* to test-grant-client")

	// 3. Also grant handlestore (needed for handle release in tests).
	runAdmin(t, dbPath, "grants", "approve", "test-grant-client", "/handlestore.HandleStoreService/*")

	// 4. Now FindClass should succeed.
	result := runJnicliAuth(t, certDir, "jni", "class", "find", "--name", "java/lang/String")
	resp := parseJSON(t, result)
	handle := getInt64Field(t, resp, "classHandle")
	if handle <= 0 {
		t.Errorf("expected classHandle > 0, got %d", handle)
	}
	t.Logf("FindClass succeeded after grant: classHandle=%d", handle)
}

func TestE2E_Auth_ListPermissions(t *testing.T) {
	skipIfNoMTLS(t)

	dbPath := os.Getenv("JNISERVICE_DB")
	if dbPath == "" {
		dbPath = "/data/local/tmp/jniservice.db"
	}
	adminBin := os.Getenv("JNISERVICEADMIN_BIN")
	if adminBin == "" {
		t.Skip("JNISERVICEADMIN_BIN not set")
	}

	certDir := t.TempDir()

	// Register.
	addr := os.Getenv("JNICTL_E2E_ADDR")
	cmd := jnicliCommand(
		"--addr", addr, "--insecure",
		"auth", "register",
		"--cn", "test-list-client",
		"--key-out", filepath.Join(certDir, "client.key"),
		"--cert-out", filepath.Join(certDir, "client.crt"),
		"--ca-out", filepath.Join(certDir, "ca.crt"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("register: %v\n%s", err, out)
	}

	// Grant some methods.
	runAdmin(t, dbPath, "grants", "approve", "test-list-client", "/jni_raw.JNIService/FindClass")

	// List permissions via jnicli.
	result := runJnicliAuth(t, certDir, "auth", "list-permissions")
	if !strings.Contains(result, "FindClass") {
		t.Errorf("expected FindClass in permissions list, got: %s", result)
	}
	t.Logf("Permissions: %s", strings.TrimSpace(result))
}
