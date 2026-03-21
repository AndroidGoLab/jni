//go:build android

// Command biometric demonstrates Android BiometricManager by querying
// the device's biometric authentication capabilities. It calls every
// available query method: CanAuthenticate0, CanAuthenticate1_1,
// GetLastAuthenticationTime, and GetStrings.
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/hardware/biometric"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
}

// Android BiometricManager.Authenticators constants.
const (
	authenticatorBiometricStrong  = 0x000F
	authenticatorBiometricWeak    = 0x00FF
	authenticatorDeviceCredential = 0x8000
)

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

	mgr, err := biometric.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("biometric.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Biometric Manager ===")

	// CanAuthenticate0 - no arguments, API 28+.
	result, err := mgr.CanAuthenticate0()
	if err != nil {
		fmt.Fprintf(output, "canAuthenticate(): error: %v\n", err)
	} else {
		fmt.Fprintf(output, "canAuthenticate(): %s\n", canAuthResultName(result))
	}

	// CanAuthenticate1_1 - with authenticator types, API 30+.
	type authQuery struct {
		name string
		flag int32
	}
	queries := []authQuery{
		{"BIOMETRIC_STRONG", authenticatorBiometricStrong},
		{"BIOMETRIC_WEAK", authenticatorBiometricWeak},
		{"DEVICE_CREDENTIAL", authenticatorDeviceCredential},
		{"STRONG|CREDENTIAL", authenticatorBiometricStrong | authenticatorDeviceCredential},
	}
	for _, q := range queries {
		r, err := mgr.CanAuthenticate1_1(q.flag)
		if err != nil {
			fmt.Fprintf(output, "canAuth(%s): error: %v\n", q.name, err)
		} else {
			fmt.Fprintf(output, "canAuth(%s): %s\n", q.name, canAuthResultName(r))
		}
	}

	// GetLastAuthenticationTime - API 34+.
	for _, q := range queries {
		lastAuth, err := mgr.GetLastAuthenticationTime(q.flag)
		if err != nil {
			fmt.Fprintf(output, "lastAuth(%s): error: %v\n", q.name, err)
		} else if lastAuth == int64(biometric.BiometricNoAuthentication) {
			fmt.Fprintf(output, "lastAuth(%s): none\n", q.name)
		} else {
			fmt.Fprintf(output, "lastAuth(%s): %d ms\n", q.name, lastAuth)
		}
	}

	// GetStrings - returns BiometricManager.Strings object, API 34+.
	// Call for each authenticator type to see what prompts the system provides.
	for _, q := range queries {
		stringsObj, err := mgr.GetStrings(q.flag)
		if err != nil {
			fmt.Fprintf(output, "strings(%s): error: %v\n", q.name, err)
		} else if stringsObj == nil {
			fmt.Fprintf(output, "strings(%s): nil\n", q.name)
		} else {
			// The Strings object has getButtonLabel, getPromptMessage, getSettingName.
			// Call toString() to get a human-readable summary.
			var desc string
			vm.Do(func(env *jni.Env) error {
				cls := env.GetObjectClass(stringsObj)
				mid, err := env.GetMethodID(cls, "toString", "()Ljava/lang/String;")
				if err != nil {
					return err
				}
				obj, err := env.CallObjectMethod(stringsObj, mid)
				if err != nil {
					return err
				}
				desc = env.GoString((*jni.String)(unsafe.Pointer(obj)))
				return nil
			})
			if desc != "" {
				fmt.Fprintf(output, "strings(%s): %s\n", q.name, desc)
			} else {
				fmt.Fprintf(output, "strings(%s): (object)\n", q.name)
			}
		}
	}

	return nil
}
