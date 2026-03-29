//go:build android

// Command keyguard_lock_state demonstrates the KeyguardManager API to
// check if the device is locked, if the keyguard is secure, and if
// the device has a secure lock screen.
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
	"github.com/AndroidGoLab/jni/os/keyguard"
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

	mgr, err := keyguard.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("keyguard.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Keyguard Lock State ===")
	fmt.Fprintln(output)

	// --- Keyguard locked ---
	locked, err := mgr.IsKeyguardLocked()
	if err != nil {
		fmt.Fprintf(output, "keyguard locked: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "keyguard locked: %v\n", locked)
	}

	// --- Keyguard secure ---
	secure, err := mgr.IsKeyguardSecure()
	if err != nil {
		fmt.Fprintf(output, "keyguard secure: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "keyguard secure: %v\n", secure)
	}

	// --- Device locked ---
	deviceLocked, err := mgr.IsDeviceLocked()
	if err != nil {
		fmt.Fprintf(output, "device locked: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "device locked: %v\n", deviceLocked)
	}

	// --- Device secure ---
	deviceSecure, err := mgr.IsDeviceSecure()
	if err != nil {
		fmt.Fprintf(output, "device secure: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "device secure: %v\n", deviceSecure)
	}

	// --- Restricted input mode ---
	restrictedInput, err := mgr.InKeyguardRestrictedInputMode()
	if err != nil {
		fmt.Fprintf(output, "restricted input: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "restricted input: %v\n", restrictedInput)
	}

	// --- Confirm device credential intent ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Confirm credential intent:")
	intentObj, err := mgr.CreateConfirmDeviceCredentialIntent("Go JNI", "Please confirm your identity")
	if err != nil {
		fmt.Fprintf(output, "  create: error: %v\n", err)
	} else if intentObj == nil {
		fmt.Fprintln(output, "  create: null (no secure lock)")
	} else {
		fmt.Fprintln(output, "  create: OK (intent available)")
	}

	// --- Keyguard lock ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Keyguard lock:")
	lockObj, err := mgr.NewKeyguardLock("go-jni-demo")
	if err != nil {
		fmt.Fprintf(output, "  create: error: %v\n", err)
	} else if lockObj == nil {
		fmt.Fprintln(output, "  create: null")
	} else {
		fmt.Fprintln(output, "  create: OK (tag=go-jni-demo)")
	}

	// --- Callback API surface ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Dismiss callback methods:")
	fmt.Fprintln(output, "  OnDismissSucceeded()")
	fmt.Fprintln(output, "  OnDismissError()")
	fmt.Fprintln(output, "  OnDismissCancelled()")

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Locked state listener:")
	fmt.Fprintln(output, "  OnLockedStateChanged(isLocked)")

	// --- Summary ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Summary:")
	if secure {
		fmt.Fprintln(output, "  Device has a secure lock screen (PIN/pattern/password).")
	} else {
		fmt.Fprintln(output, "  Device does NOT have a secure lock screen.")
	}
	if locked {
		fmt.Fprintln(output, "  Keyguard is currently displayed.")
	} else {
		fmt.Fprintln(output, "  Keyguard is NOT displayed (device unlocked).")
	}

	return nil
}
