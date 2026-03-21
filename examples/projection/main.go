//go:build android

// Command projection demonstrates using the MediaProjection API. It is built
// as a c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the MediaProjectionManager system service and
// describes the screen capture workflow: creating a capture intent,
// obtaining a MediaProjection from the activity result, registering a
// projectionCallback, and creating a virtual display.
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
	"github.com/AndroidGoLab/jni/media/projection"
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
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- NewManager ---
	mgr, err := projection.NewMediaProjectionManager(ctx)
	if err != nil {
		return fmt.Errorf("projection.NewMediaProjectionManager: %w", err)
	}
	defer mgr.Close()

	// --- Manager methods (unexported) ---
	// The screen capture workflow:
	//
	// 1. Create a screen capture intent:
	//    intent, err := mgr.createScreenCaptureIntent()
	//    Pass this intent to startActivityForResult.
	//
	// 2. Obtain the projection from the result:
	//    projObj, err := mgr.getMediaProjection(resultCode, resultData)

	// --- Projection (unexported methods) ---
	// The Projection type wraps android.media.projection.MediaProjection:
	//
	//   proj.stop()
	//     Stop the media projection.
	//
	//   proj.registerCallback(callback, handler *jni.Object) error
	//     Register a callback for projection lifecycle events.
	//
	//   proj.createVirtualDisplayRaw(name, width, height, dpi, flags,
	//     surface, callback, handler) (*jni.Object, error)
	//     Create a virtual display for screen capture.

	// --- projectionCallback (unexported) ---
	// Registered via registerprojectionCallback to handle stop events:
	//
	//   projectionCallback{
	//     OnStop func()
	//   }
	//   proxy, cleanup, err := registerprojectionCallback(env, cb)

	// --- VirtualDisplay (unexported methods) ---
	// The VirtualDisplay type wraps android.hardware.display.VirtualDisplay:
	//
	//   vd.release()
	//     Release the virtual display when screen capture is done.

	fmt.Fprintf(output, "MediaProjectionManager obtained from context\n")
	fmt.Fprintln(output, "Projection example complete.")
	return nil
}
