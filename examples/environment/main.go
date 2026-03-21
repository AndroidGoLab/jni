//go:build android

// Command environment demonstrates the Android Environment API constants.
// It is built as a c-shared library and packaged into an APK.
//
// This package provides Go bindings for android.os.Environment. The class
// methods are exposed via an unexported environment type, but the package
// exports constants for standard directory names and external storage states.
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

	"github.com/AndroidGoLab/jni/os/environment"
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
	// Exported constants for external storage state values.
	// These correspond to Environment.MEDIA_* constants in Java.
	fmt.Fprintln(output, "External storage state constants:")
	fmt.Fprintf(output, "  MediaMounted:         %s\n", environment.MediaMounted)
	fmt.Fprintf(output, "  MediaMountedReadOnly: %s\n", environment.MediaMountedReadOnly)
	fmt.Fprintf(output, "  MediaRemoved:         %s\n", environment.MediaRemoved)
	fmt.Fprintf(output, "  MediaUnmounted:       %s\n", environment.MediaUnmounted)

	// The environment class methods (getExternalStorageDirectory,
	// getExternalStoragePublicDirectory, getExternalStorageState, etc.)
	// are accessed via an unexported environment type.
	//
	// The javaFile type (also unexported) wraps java.io.File with:
	//   javaFile.getAbsolutePath() string
	//     Returns the absolute path of the file.
	return nil
}
