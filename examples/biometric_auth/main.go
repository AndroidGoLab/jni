//go:build android

// Command biometric_auth demonstrates the BiometricPrompt authentication
// API surface. It checks biometric hardware availability and queries
// authenticator support for strong, weak, and device credential types.
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
	"github.com/AndroidGoLab/jni/hardware/biometric"
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

func canAuthResultName(code int32) string {
	switch int(code) {
	case biometric.BiometricSuccess:
		return "SUCCESS"
	case biometric.BiometricErrorNoHardware:
		return "NO_HARDWARE"
	case biometric.BiometricErrorHwUnavailable:
		return "HW_UNAVAILABLE"
	case biometric.BiometricErrorNoneEnrolled:
		return "NONE_ENROLLED"
	case biometric.BiometricErrorSecurityUpdateRequired:
		return "SECURITY_UPDATE_REQUIRED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", code)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Biometric Authentication ===")

	// --- Check hardware availability ---
	mgr, err := biometric.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("biometric.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Hardware check (canAuthenticate):")

	// No-arg canAuthenticate (API 28+).
	result, err := mgr.CanAuthenticate0()
	if err != nil {
		fmt.Fprintf(output, "  default: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  default: %s\n", canAuthResultName(result))
	}

	// Typed canAuthenticate (API 30+).
	type authQuery struct {
		name string
		flag int32
	}
	queries := []authQuery{
		{"BIOMETRIC_STRONG", int32(biometric.BiometricStrong)},
		{"BIOMETRIC_WEAK", int32(biometric.BiometricWeak)},
		{"DEVICE_CREDENTIAL", int32(biometric.DeviceCredential)},
		{"STRONG|CREDENTIAL", int32(biometric.BiometricStrong) | int32(biometric.DeviceCredential)},
	}
	for _, q := range queries {
		r, err := mgr.CanAuthenticate1_1(q.flag)
		if err != nil {
			fmt.Fprintf(output, "  %s: error: %v\n", q.name, err)
		} else {
			fmt.Fprintf(output, "  %s: %s\n", q.name, canAuthResultName(r))
		}
	}

	// --- CryptoObject API surface ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "CryptoObject API surface:")
	fmt.Fprintln(output, "  GetCipher()")
	fmt.Fprintln(output, "  GetMac()")
	fmt.Fprintln(output, "  GetSignature()")
	fmt.Fprintln(output, "  GetIdentityCredential()")
	fmt.Fprintln(output, "  GetPresentationSession()")
	fmt.Fprintln(output, "  GetOperationHandle()")

	// --- Authentication callback types ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "AuthenticationCallback methods:")
	fmt.Fprintln(output, "  OnAuthenticationSucceeded(result)")
	fmt.Fprintln(output, "  OnAuthenticationFailed()")
	fmt.Fprintln(output, "  OnAuthenticationError(errorCode, errString)")
	fmt.Fprintln(output, "  OnAuthenticationHelp(helpCode, helpString)")

	// --- Constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Constants:")
	fmt.Fprintf(output, "  BIOMETRIC_STRONG  = 0x%04X\n", biometric.BiometricStrong)
	fmt.Fprintf(output, "  BIOMETRIC_WEAK    = 0x%04X\n", biometric.BiometricWeak)
	fmt.Fprintf(output, "  DEVICE_CREDENTIAL = 0x%04X\n", biometric.DeviceCredential)

	return nil
}
