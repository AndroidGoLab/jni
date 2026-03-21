//go:build android

// Command health demonstrates the Health Connect API provided by the
// generated health/connect package. It is built as a c-shared library
// and packaged into an APK.
//
// The connect.Manager wraps the HealthConnectClient and provides raw
// methods for inserting, reading, aggregating, and deleting records.
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
	"unsafe"
	"bytes"
	"fmt"

	_ "github.com/AndroidGoLab/jni/health/connect"
	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
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
	// The health/connect package provides a Manager type wrapping
	// HealthConnectClient with raw methods:
	//   getOrCreateRaw(ctx) - obtain a HealthConnectClient
	//   insertRecordsRaw(records) - insert health records
	//   readRecordsRaw(request) - read health records
	//   aggregateRaw(request) - aggregate health data
	//   deleteRecordsRaw(recordType, timeRange) - delete records
	//
	// These are unexported and intended for use by higher-level wrappers.
	fmt.Fprintln(output, "Health Connect Manager type available")
	fmt.Fprintln(output, "Raw methods: getOrCreateRaw, insertRecordsRaw, readRecordsRaw, aggregateRaw, deleteRecordsRaw")
	return nil
}
