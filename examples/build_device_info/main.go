//go:build android

// Command build_device_info reads Build fields (manufacturer, model,
// brand, device, product, board, hardware) and Build.VERSION fields
// (SDK version, release, codename, incremental) using typed wrappers.
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
	"github.com/AndroidGoLab/jni/os/build"
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
	fmt.Fprintln(output, "=== Build Device Info ===")
	fmt.Fprintln(output)

	// --- Build info ---
	info, err := build.GetBuildInfo(vm)
	if err != nil {
		return fmt.Errorf("build.GetBuildInfo: %w", err)
	}

	fmt.Fprintln(output, "android.os.Build:")
	fmt.Fprintf(output, "  MANUFACTURER = %s\n", info.Manufacturer)
	fmt.Fprintf(output, "  MODEL        = %s\n", info.Model)
	fmt.Fprintf(output, "  BRAND        = %s\n", info.Brand)
	fmt.Fprintf(output, "  DEVICE       = %s\n", info.Device)
	fmt.Fprintf(output, "  PRODUCT      = %s\n", info.Product)
	fmt.Fprintf(output, "  BOARD        = %s\n", info.Board)
	fmt.Fprintf(output, "  HARDWARE     = %s\n", info.Hardware)

	// --- Version info ---
	fmt.Fprintln(output)
	ver, err := build.GetVersionInfo(vm)
	if err != nil {
		return fmt.Errorf("build.GetVersionInfo: %w", err)
	}

	fmt.Fprintln(output, "android.os.Build.VERSION:")
	fmt.Fprintf(output, "  SDK_INT     = %d\n", ver.SDKInt)
	fmt.Fprintf(output, "  RELEASE     = %s\n", ver.Release)
	fmt.Fprintf(output, "  CODENAME    = %s\n", ver.Codename)
	fmt.Fprintf(output, "  INCREMENTAL = %s\n", ver.Incremental)

	// --- Build static methods ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Build static methods:")

	b, err := build.NewBuild(vm)
	if err != nil {
		fmt.Fprintf(output, "  NewBuild: %v\n", err)
	} else {
		radioVer, err := b.GetRadioVersion()
		if err != nil {
			fmt.Fprintf(output, "  radioVersion: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  radioVersion: %s\n", radioVer)
		}
	}

	return nil
}
