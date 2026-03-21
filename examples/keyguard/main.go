//go:build android

// Command keyguard demonstrates the KeyguardManager JNI bindings. It is built
// as a c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the KeyguardManager system service and queries
// the device lock state. It demonstrates all exported boolean query
// methods and shows how to set up a keyguardDismissCallback for
// receiving keyguard dismiss events.
package main

/*
#include <android/native_activity.h>
extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static void _setCallbacks(ANativeActivity* a) { a->callbacks->onResume = _onResume; a->callbacks->onNativeWindowCreated = _onWindowCreated; }
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
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
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

	// Check whether the keyguard (lock screen) is currently displayed.
	locked, err := mgr.IsKeyguardLocked()
	if err != nil {
		return fmt.Errorf("IsKeyguardLocked: %w", err)
	}
	fmt.Fprintf(output, "keyguard locked: %v\n", locked)

	// Check whether the keyguard is secured by a PIN, pattern, or password.
	secure, err := mgr.IsKeyguardSecure()
	if err != nil {
		return fmt.Errorf("IsKeyguardSecure: %w", err)
	}
	fmt.Fprintf(output, "keyguard secure: %v\n", secure)

	// Check whether the device is currently locked (for the current user).
	deviceLocked, err := mgr.IsDeviceLocked()
	if err != nil {
		return fmt.Errorf("IsDeviceLocked: %w", err)
	}
	fmt.Fprintf(output, "device locked: %v\n", deviceLocked)

	// Check whether the device has a secure lock screen.
	deviceSecure, err := mgr.IsDeviceSecure()
	if err != nil {
		return fmt.Errorf("IsDeviceSecure: %w", err)
	}
	fmt.Fprintf(output, "device secure: %v\n", deviceSecure)

	// The keyguardDismissCallback is used with requestDismissKeyguard
	// to receive notifications about the keyguard dismiss outcome.
	// It supports three callback functions:
	//   OnDismissSucceeded  - called when the keyguard is dismissed
	//   OnDismissError      - called when an error occurs
	//   OnDismissCancelled  - called when the user cancels the dismiss
	//
	// The keyguardLockedStateListener is used to listen for changes
	// in the keyguard locked state via:
	//   OnLockedStateChanged(isLocked bool)
	fmt.Fprintln(output, "keyguard state query complete")
	return nil
}
