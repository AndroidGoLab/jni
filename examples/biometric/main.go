//go:build android

// Command biometric demonstrates Android biometric authentication
// using BiometricManager. Requires API 28+. It is built as a c-shared
// library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// The biometric package wraps android.hardware.biometrics.BiometricManager
// and related types (BiometricPrompt, BiometricPrompt.Builder,
// CancellationSignal). Most prompt-related types and methods are
// unexported and intended to be wrapped by higher-level helpers.
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
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/hardware/biometric"
)

func main() {}

func init() { exampleui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	exampleui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	exampleui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	exampleui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := biometric.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("biometric.NewManager: %w", err)
	}
	defer mgr.Close()

	// Error code constants from BiometricPrompt.
	fmt.Fprintf(output, "errors: hw_unavailable=%d, canceled=%d, lockout=%d, lockout_permanent=%d, user_canceled=%d, no_biometrics=%d\n",
		biometric.BiometricErrorHwUnavailable, biometric.BiometricErrorCanceled,
		biometric.BiometricErrorLockout, biometric.BiometricErrorLockoutPermanent,
		biometric.BiometricErrorUserCanceled, biometric.BiometricErrorNoBiometrics)

	// Availability check result constants from BiometricManager.canAuthenticate.
	fmt.Fprintf(output, "availability: success=%d, no_hardware=%d, hw_unavailable=%d, none_enrolled=%d\n",
		biometric.BiometricSuccess, biometric.BiometricErrorNoHardware,
		biometric.BiometricErrorHwUnavailable, biometric.BiometricErrorNoneEnrolled)

	// The authentication flow uses unexported types:
	//   1. biometricPromptBuilder - set title, subtitle, description,
	//      negative button, and allowed authenticators, then build()
	//   2. biometricPrompt - call authenticate() with a cancellation signal,
	//      executor, and callback
	//   3. cancellationSignal - cancel() to abort authentication
	//
	// The canAuthenticateRaw(authenticators) method on Manager checks
	// whether biometric authentication is available on the device.
	fmt.Fprintln(output, "BiometricManager created successfully")

	return nil
}
