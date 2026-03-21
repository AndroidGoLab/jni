//go:build android

// Command mediastore demonstrates the MediaStore JNI bindings.
// It is built as a c-shared library and packaged into an APK.
//
// This example prints all available MediaStore constants including
// intent actions, volume names, content URIs, and column names used
// for querying media files. It also describes the static methods
// available for creating media modification requests.
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

	"github.com/AndroidGoLab/jni/provider/media"
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
	// Intent action constants for capturing and picking media.
	fmt.Fprintln(output, "Intent actions:")
	fmt.Fprintf(output, "  ActionImageCapture = %q\n", media.ActionImageCapture)
	fmt.Fprintf(output, "  ActionVideoCapture = %q\n", media.ActionVideoCapture)
	fmt.Fprintf(output, "  ActionPickImages   = %q\n", media.ActionPickImages)
	fmt.Fprintf(output, "  ExtraPickImagesMax = %q\n", media.ExtraPickImagesMax)

	// Volume constants identify storage volumes.
	fmt.Fprintln(output, "Storage volumes:")
	fmt.Fprintf(output, "  VolumeInternal        = %q\n", media.VolumeInternal)
	fmt.Fprintf(output, "  VolumeExternal        = %q\n", media.VolumeExternal)
	fmt.Fprintf(output, "  VolumeExternalPrimary = %q\n", media.VolumeExternalPrimary)

	// The mediaStore wrapper provides static methods:
	//   getExternalVolumeNamesRaw, createWriteRequestRaw,
	//   createTrashRequestRaw, createDeleteRequestRaw,
	//   createFavoriteRequestRaw

	// The mediaStore wrapper provides static methods:
	//   getExternalVolumeNamesRaw(ctx)              - get external volume names
	//   createWriteRequestRaw(resolver, uris)       - create write permission request
	//   createTrashRequestRaw(resolver, uris, trash)- create trash/untrash request
	//   createDeleteRequestRaw(resolver, uris)      - create delete permission request
	//   createFavoriteRequestRaw(resolver, uris, fav)- create favorite toggle request
	fmt.Fprintln(output, "mediastore bindings available for media operations")
	return nil
}
