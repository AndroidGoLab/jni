//go:build !android

package e2e_grpc_test

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	testCertFile string
	testKeyFile  string
	testCAFile   string
)

func TestMain(m *testing.M) {
	addr := os.Getenv("JNICTL_E2E_ADDR")
	if addr == "" {
		// No address set: individual tests will skip via skipIfNoEmulator.
		os.Exit(m.Run())
	}

	certDir, err := os.MkdirTemp("", "e2e-certs-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "E2E setup: creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(certDir)

	testCertFile = filepath.Join(certDir, "client.crt")
	testKeyFile = filepath.Join(certDir, "client.key")
	testCAFile = filepath.Join(certDir, "ca.crt")

	// Use a randomized CN to avoid UNIQUE constraint collisions with
	// previous test runs that may not have cleaned up.
	cn := fmt.Sprintf("e2e-test-%d", rand.Int63())

	// Register a client certificate. The Register RPC is exempted from
	// mTLS auth, so --insecure (TLS with skip-verify, no client cert) works.
	regCmd := jnicliCommand(
		"--addr", addr, "--insecure",
		"auth", "register",
		"--cn", cn,
		"--cert-out", testCertFile,
		"--key-out", testKeyFile,
		"--ca-out", testCAFile,
	)
	out, err := regCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "E2E setup: register failed: %v\n%s\n", err, out)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "E2E setup: registered client %q\n", cn)

	// Grant the test client full access via the admin tool.
	if err := grantTestPermissions(cn); err != nil {
		fmt.Fprintf(os.Stderr, "E2E setup: granting permissions: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// grantTestPermissions grants "/*" to the test client using
// jniserviceadmin. Supports two modes:
//
//   - Host mode: JNISERVICEADMIN_BIN is set to a host binary and
//     JNISERVICE_DB is a host-accessible path. The admin tool is run
//     directly on the host.
//
//   - ADB mode: JNISERVICEADMIN_ADB_BIN is set to the device-side path
//     of jniserviceadmin (e.g. /data/local/tmp/jniserviceadmin). The
//     admin tool is invoked via "adb shell" on the device. JNISERVICE_DB
//     defaults to the device-side path.
func grantTestPermissions(cn string) error {
	// Try host mode first.
	if adminBin := os.Getenv("JNISERVICEADMIN_BIN"); adminBin != "" {
		return grantViaHostAdmin(adminBin, cn)
	}

	// Fall back to adb mode.
	if adbAdmin := os.Getenv("JNISERVICEADMIN_ADB_BIN"); adbAdmin != "" {
		return grantViaADB(adbAdmin, cn)
	}

	return fmt.Errorf("neither JNISERVICEADMIN_BIN nor JNISERVICEADMIN_ADB_BIN is set; cannot grant permissions")
}

func grantViaHostAdmin(adminBin, cn string) error {
	dbPath := os.Getenv("JNISERVICE_DB")
	if dbPath == "" {
		dbPath = "/data/local/tmp/jniservice/acl.db"
	}

	cmd := exec.Command(adminBin, "--db", dbPath, "grants", "approve", cn, "/*")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("jniserviceadmin grants approve: %v\n%s", err, out)
	}
	fmt.Fprintf(os.Stderr, "E2E setup: granted /* to %q (host mode)\n", cn)
	return nil
}

func grantViaADB(adbAdminPath, cn string) error {
	dbPath := os.Getenv("JNISERVICE_DB")
	if dbPath == "" {
		dbPath = "/data/local/tmp/jniservice/acl.db"
	}

	adb := os.Getenv("ADB")
	if adb == "" {
		adb = "adb"
	}

	// Build the shell command with proper quoting.
	shellCmd := strings.Join([]string{
		adbAdminPath,
		"--db", dbPath,
		"grants", "approve", cn, "/*",
	}, " ")

	cmd := exec.Command(adb, "shell", shellCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb shell jniserviceadmin: %v\n%s", err, out)
	}
	fmt.Fprintf(os.Stderr, "E2E setup: granted /* to %q (adb mode)\n", cn)
	return nil
}
