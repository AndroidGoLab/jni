//go:build android

// Command biometric_auth demonstrates the BiometricManager API by making
// real calls: CanAuthenticate with multiple authenticator types, GetStrings,
// GetLastAuthenticationTime, and ToString.
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

	// 1. Obtain BiometricManager.
	mgr, err := biometric.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("biometric.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "BiometricManager: obtained OK")

	// 2. ToString.
	str, err := mgr.ToString()
	if err != nil {
		fmt.Fprintf(output, "  ToString: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  ToString: %s\n", str)
	}

	// 3. CanAuthenticate0 (no-arg, API 28+).
	fmt.Fprintln(output)
	fmt.Fprintln(output, "CanAuthenticate (default):")
	result, err := mgr.CanAuthenticate0()
	if err != nil {
		fmt.Fprintf(output, "  error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  result: %s\n", canAuthResultName(result))
	}

	// 4. CanAuthenticate1_1 with various authenticator types (API 30+).
	fmt.Fprintln(output)
	fmt.Fprintln(output, "CanAuthenticate (typed):")
	type authQuery struct {
		name string
		flag int32
	}
	queries := []authQuery{
		{"BIOMETRIC_STRONG", int32(biometric.BiometricStrong)},
		{"BIOMETRIC_WEAK", int32(biometric.BiometricWeak)},
		{"DEVICE_CREDENTIAL", int32(biometric.DeviceCredential)},
		{"STRONG|CREDENTIAL", int32(biometric.BiometricStrong) | int32(biometric.DeviceCredential)},
		{"WEAK|CREDENTIAL", int32(biometric.BiometricWeak) | int32(biometric.DeviceCredential)},
	}
	for _, q := range queries {
		r, err := mgr.CanAuthenticate1_1(q.flag)
		if err != nil {
			fmt.Fprintf(output, "  %s: error: %v\n", q.name, err)
		} else {
			fmt.Fprintf(output, "  %s: %s\n", q.name, canAuthResultName(r))
		}
	}

	// 5. GetStrings for each authenticator type (API 31+).
	fmt.Fprintln(output)
	fmt.Fprintln(output, "GetStrings:")
	for _, q := range queries {
		stringsObj, err := mgr.GetStrings(q.flag)
		if err != nil {
			fmt.Fprintf(output, "  %s: error: %v\n", q.name, err)
		} else if stringsObj == nil || stringsObj.Ref() == 0 {
			fmt.Fprintf(output, "  %s: (null)\n", q.name)
		} else {
			// Wrap the returned object as ManagerStrings and call its methods.
			ms := biometric.ManagerStrings{VM: vm, Obj: stringsObj}
			msStr, _ := ms.ToString()
			fmt.Fprintf(output, "  %s: %s\n", q.name, msStr)
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(stringsObj)
				return nil
			})
		}
	}

	// 6. GetLastAuthenticationTime for each authenticator type (API 34+).
	fmt.Fprintln(output)
	fmt.Fprintln(output, "GetLastAuthenticationTime:")
	for _, q := range queries {
		ts, err := mgr.GetLastAuthenticationTime(q.flag)
		if err != nil {
			fmt.Fprintf(output, "  %s: error: %v\n", q.name, err)
		} else {
			if ts == biometric.BiometricNoAuthentication {
				fmt.Fprintf(output, "  %s: NEVER\n", q.name)
			} else {
				fmt.Fprintf(output, "  %s: %d ms since epoch\n", q.name, ts)
			}
		}
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Biometric auth example complete.")
	return nil
}
