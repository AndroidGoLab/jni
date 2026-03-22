//go:build android

// Command build_info demonstrates reading Android device build
// information such as manufacturer, model, and SDK version.
//
// The build package wraps android.os.Build and android.os.Build.VERSION,
// providing GetBuildInfo() and GetVersionInfo() functions that read
// static fields into Go structs.
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
	"github.com/AndroidGoLab/jni/os/build"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(activity.vm)),
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	// GetBuildInfo reads static fields from android.os.Build:
	// Manufacturer, Model, Brand, Device, Hardware, Product, Fingerprint, ID.
	info, err := build.GetBuildInfo(vm)
	if err != nil {
		return fmt.Errorf("build.GetBuildInfo: %w", err)
	}
	fmt.Fprintf(output, "build info: %+v\n", info)

	// GetVersionInfo reads android.os.Build.VERSION fields:
	// SDKInt (API level as int32) and Release (version string like "14").
	ver, err := build.GetVersionInfo(vm)
	if err != nil {
		return fmt.Errorf("build.GetVersionInfo: %w", err)
	}
	fmt.Fprintf(output, "version info: %+v\n", ver)

	return nil
}
