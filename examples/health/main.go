//go:build android

// Command health demonstrates Health Connect availability checking.
// Health Connect requires the androidx library and the Health Connect
// app to be installed. This example checks whether the required
// HealthConnectClient class is loadable and reports the status.
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
	"github.com/AndroidGoLab/jni/health/connect"
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

func run(vm *jni.VM, output *bytes.Buffer) error {
	fmt.Fprintln(output, "=== Health Connect ===")
	fmt.Fprintln(output)

	// Try to initialize the health/connect package, which resolves
	// the HealthConnectClient class via JNI. If the class is not
	// found (HealthConnect app not installed or no androidx lib),
	// Init returns an error.
	var initErr error
	vm.Do(func(env *jni.Env) error {
		initErr = connect.Init(env)
		return nil
	})

	switch {
	case initErr == nil:
		fmt.Fprintln(output, "status: AVAILABLE")
		fmt.Fprintln(output, "HealthConnectClient class loaded OK")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Available raw methods:")
		fmt.Fprintln(output, "  getOrCreate(ctx)")
		fmt.Fprintln(output, "  insertRecords(list)")
		fmt.Fprintln(output, "  readRecords(req)")
		fmt.Fprintln(output, "  aggregate(req)")
		fmt.Fprintln(output, "  deleteRecords(type,range)")
	default:
		fmt.Fprintln(output, "status: NOT AVAILABLE")
		fmt.Fprintf(output, "reason: %v\n", initErr)
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Health Connect requires the")
		fmt.Fprintln(output, "HealthConnect app or an androidx")
		fmt.Fprintln(output, "library on the classpath.")
	}

	return nil
}
