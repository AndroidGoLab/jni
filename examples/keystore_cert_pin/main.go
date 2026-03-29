//go:build android

// Command keystore_cert_pin demonstrates the Android KeyStore typed wrapper
// API surface. It shows the KeyStoreManager service, key protection
// parameters, and authentication requirements available through the
// typed wrappers.
package main

/*
#include <android/native_activity.h>
extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static void _setCallbacks(ANativeActivity* a) { a->callbacks->onResume = _onResume; a->callbacks->onNativeWindowCreated = _onWindowCreated; }
static uintptr_t _getVM(ANativeActivity* a) { return (uintptr_t)a->vm; }
static uintptr_t _getClazz(ANativeActivity* a) { return (uintptr_t)a->clazz; }
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/security/keystore"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(C._getVM(activity))),
		jni.ObjectFromUintptr(uintptr(C._getClazz(activity))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(C._getClazz(activity))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== KeyStore Certificate Pinning ===")
	fmt.Fprintln(output)

	// --- KeyStoreManager service ---
	fmt.Fprintln(output, "KeyStoreManager service:")
	ksMgr, err := keystore.NewKeyStoreManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  not available: %v\n", err)
	} else {
		fmt.Fprintln(output, "  obtained OK")
		ksMgr.Close()
	}

	// --- Authentication type constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Authentication types:")
	fmt.Fprintf(output, "  AUTH_BIOMETRIC_STRONG = %d\n", keystore.AuthBiometricStrong)
	fmt.Fprintf(output, "  AUTH_DEVICE_CREDENTIAL= %d\n", keystore.AuthDeviceCredential)

	// --- Key purposes ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Key purposes:")
	fmt.Fprintf(output, "  ENCRYPT   = %d\n", keystore.PurposeEncrypt)
	fmt.Fprintf(output, "  DECRYPT   = %d\n", keystore.PurposeDecrypt)
	fmt.Fprintf(output, "  SIGN      = %d\n", keystore.PurposeSign)
	fmt.Fprintf(output, "  VERIFY    = %d\n", keystore.PurposeVerify)
	fmt.Fprintf(output, "  WRAP_KEY  = %d\n", keystore.PurposeWrapKey)
	fmt.Fprintf(output, "  AGREE_KEY = %d\n", keystore.PurposeAgreeKey)

	// --- Security levels ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Security levels:")
	fmt.Fprintf(output, "  SOFTWARE            = %d\n", keystore.SecurityLevelSoftware)
	fmt.Fprintf(output, "  TRUSTED_ENVIRONMENT = %d\n", keystore.SecurityLevelTrustedEnvironment)
	fmt.Fprintf(output, "  STRONGBOX           = %d\n", keystore.SecurityLevelStrongbox)

	// --- Key algorithms ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Key algorithms:")
	fmt.Fprintf(output, "  AES  = %s\n", keystore.KeyAlgorithmAes)
	fmt.Fprintf(output, "  RSA  = %s\n", keystore.KeyAlgorithmRsa)
	fmt.Fprintf(output, "  EC   = %s\n", keystore.KeyAlgorithmEc)
	fmt.Fprintf(output, "  HMAC = %s\n", keystore.KeyAlgorithmHmacSha256)

	// --- Block modes ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Block modes:")
	fmt.Fprintf(output, "  CBC = %s\n", keystore.BlockModeCbc)
	fmt.Fprintf(output, "  GCM = %s\n", keystore.BlockModeGcm)
	fmt.Fprintf(output, "  CTR = %s\n", keystore.BlockModeCtr)
	fmt.Fprintf(output, "  ECB = %s\n", keystore.BlockModeEcb)

	return nil
}
