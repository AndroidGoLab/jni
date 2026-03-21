//go:build android

// Command permission demonstrates checking Android runtime permissions
// using Context.checkSelfPermission.
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

	// Ensure permission JNI bindings are initialized.
	_ "github.com/AndroidGoLab/jni/content/permission"
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
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Permission Check ===")

	// Check a set of common permissions via Context.checkSelfPermission.
	perms := []struct {
		name  string
		value string
	}{
		{"INTERNET", "android.permission.INTERNET"},
		{"CAMERA", "android.permission.CAMERA"},
		{"RECORD_AUDIO", "android.permission.RECORD_AUDIO"},
		{"ACCESS_FINE_LOCATION", "android.permission.ACCESS_FINE_LOCATION"},
		{"READ_CONTACTS", "android.permission.READ_CONTACTS"},
		{"WRITE_EXTERNAL_STORAGE", "android.permission.WRITE_EXTERNAL_STORAGE"},
		{"READ_PHONE_STATE", "android.permission.READ_PHONE_STATE"},
		{"BLUETOOTH_CONNECT", "android.permission.BLUETOOTH_CONNECT"},
	}

	for _, p := range perms {
		result, err := ctx.CheckSelfPermission(p.value)
		if err != nil {
			fmt.Fprintf(output, "  %-25s %v\n", p.name+":", err)
			continue
		}
		fmt.Fprintf(output, "  %-25s %s\n", p.name+":", permissionGrantName(result))
	}

	// Also check the package name to verify context works.
	pkgName, err := ctx.GetPackageName()
	if err != nil {
		fmt.Fprintf(output, "\nPackage: %v\n", err)
	} else {
		fmt.Fprintf(output, "\nPackage: %s\n", pkgName)
	}

	return nil
}
