//go:build android

// Command permission_request checks multiple dangerous permissions at
// runtime, reporting which are granted and which are denied.
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
	"github.com/AndroidGoLab/jni/content/permission"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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

// permissionGrantName converts the int result to a human-readable string.
// 0 = PERMISSION_GRANTED, -1 = PERMISSION_DENIED.
func permissionGrantName(result int32) string {
	switch result {
	case 0:
		return "GRANTED"
	case -1:
		return "DENIED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", result)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Permission Request ===")
	fmt.Fprintln(output, "Checking multiple dangerous permissions:")
	fmt.Fprintln(output, "")

	// Dangerous permissions to check.
	perms := []struct {
		name  string
		value string
	}{
		{"ACCESS_FINE_LOCATION", permission.AccessFineLocation},
		{"ACCESS_COARSE_LOCATION", permission.AccessCoarseLocation},
		{"CAMERA", permission.Camera},
		{"READ_CONTACTS", permission.ReadContacts},
		{"WRITE_CONTACTS", permission.WriteContacts},
		{"RECORD_AUDIO", permission.RecordAudio},
		{"READ_PHONE_STATE", permission.ReadPhoneState},
		{"READ_CALENDAR", permission.ReadCalendar},
		{"WRITE_CALENDAR", permission.WriteCalendar},
		{"BODY_SENSORS", permission.BodySensors},
	}

	grantedCount := 0
	deniedCount := 0

	for _, p := range perms {
		result, err := ctx.CheckSelfPermission(p.value)
		if err != nil {
			fmt.Fprintf(output, "  %-25s ERR: %v\n", p.name+":", err)
			continue
		}
		status := permissionGrantName(result)
		fmt.Fprintf(output, "  %-25s %s\n", p.name+":", status)
		if result == 0 {
			grantedCount++
		} else {
			deniedCount++
		}
	}

	fmt.Fprintln(output, "")
	fmt.Fprintf(output, "Summary: %d granted, %d denied\n", grantedCount, deniedCount)

	// Also check some normal (auto-granted) permissions.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Normal permissions (auto-granted):")

	normalPerms := []struct {
		name  string
		value string
	}{
		{"INTERNET", permission.Internet},
		{"ACCESS_NETWORK_STATE", permission.AccessNetworkState},
		{"ACCESS_WIFI_STATE", permission.AccessWifiState},
		{"VIBRATE", permission.Vibrate},
		{"WAKE_LOCK", permission.WakeLock},
	}

	for _, p := range normalPerms {
		result, err := ctx.CheckSelfPermission(p.value)
		if err != nil {
			fmt.Fprintf(output, "  %-25s ERR: %v\n", p.name+":", err)
			continue
		}
		fmt.Fprintf(output, "  %-25s %s\n", p.name+":", permissionGrantName(result))
	}

	// Show our package name.
	pkgName, err := ctx.GetPackageName()
	if err != nil {
		fmt.Fprintf(output, "\nPackage: %v\n", err)
	} else {
		fmt.Fprintf(output, "\nPackage: %s\n", pkgName)
	}

	fmt.Fprintln(output, "\nPermission request complete.")
	return nil
}
