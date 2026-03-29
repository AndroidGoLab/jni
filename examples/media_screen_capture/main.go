//go:build android

// Command media_screen_capture demonstrates the MediaProjection API using the
// typed wrapper package. It obtains the MediaProjectionManager, creates a
// screen capture intent, and calls ToString on the manager to verify it works.
//
// Getting an actual MediaProjection requires user consent via
// startActivityForResult, which is not possible from NativeActivity without
// a Java activity. This example demonstrates the available API surface.
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
	"github.com/AndroidGoLab/jni/media/projection"
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
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Screen Capture Demo ===")
	ui.RenderOutput()

	// --- Obtain MediaProjectionManager ---
	mgr, err := projection.NewMediaProjectionManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "NewMediaProjectionManager: %v\n", err)
		fmt.Fprintln(output, "\nScreen capture demo completed (manager unavailable).")
		return nil
	}
	if mgr == nil || mgr.Obj == nil || mgr.Obj.Ref() == 0 {
		fmt.Fprintln(output, "MediaProjectionManager: null")
		fmt.Fprintln(output, "\nScreen capture demo completed (manager null).")
		return nil
	}
	defer mgr.Close()
	fmt.Fprintln(output, "MediaProjectionManager: obtained OK")
	ui.RenderOutput()

	// --- ToString ---
	mgrStr, err := mgr.ToString()
	if err != nil {
		fmt.Fprintf(output, "Manager.ToString: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Manager.ToString: %s\n", mgrStr)
	}
	ui.RenderOutput()

	// --- CreateScreenCaptureIntent (no-arg) ---
	captureIntent, err := mgr.CreateScreenCaptureIntent0()
	if err != nil {
		fmt.Fprintf(output, "CreateScreenCaptureIntent0: error (%v)\n", err)
	} else if captureIntent == nil || captureIntent.Ref() == 0 {
		fmt.Fprintln(output, "CreateScreenCaptureIntent0: null")
	} else {
		fmt.Fprintln(output, "CreateScreenCaptureIntent0: created OK")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(captureIntent)
			return nil
		})
	}
	ui.RenderOutput()

	// --- CreateScreenCaptureIntent (with MediaProjectionConfig) ---
	// The 1-arg overload requires a MediaProjectionConfig (API 34+).
	// We pass nil to test availability; it will error on older APIs.
	captureIntent2, err := mgr.CreateScreenCaptureIntent1_1(nil)
	if err != nil {
		fmt.Fprintf(output, "CreateScreenCaptureIntent1_1(nil): %v (expected on API < 34)\n", err)
	} else if captureIntent2 == nil || captureIntent2.Ref() == 0 {
		fmt.Fprintln(output, "CreateScreenCaptureIntent1_1(nil): null")
	} else {
		fmt.Fprintln(output, "CreateScreenCaptureIntent1_1(nil): created OK")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(captureIntent2)
			return nil
		})
	}
	ui.RenderOutput()

	// --- GetMediaProjection ---
	// Requires a valid result code and intent data from startActivityForResult.
	// We call with invalid args (-1, nil) to demonstrate the API is available.
	projObj, err := mgr.GetMediaProjection(-1, nil)
	if err != nil {
		fmt.Fprintf(output, "GetMediaProjection(-1, nil): %v (expected)\n", err)
	} else if projObj == nil || projObj.Ref() == 0 {
		fmt.Fprintln(output, "GetMediaProjection(-1, nil): null (no consent)")
	} else {
		// Wrap in typed struct and query.
		proj := &projection.MediaProjection{VM: vm, Obj: projObj}
		projStr, err := proj.ToString()
		if err == nil {
			fmt.Fprintf(output, "MediaProjection.ToString: %s\n", projStr)
		}
		// Stop the projection.
		if err := proj.Stop(); err != nil {
			fmt.Fprintf(output, "MediaProjection.Stop: %v\n", err)
		} else {
			fmt.Fprintln(output, "MediaProjection.Stop: OK")
		}
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(projObj)
			return nil
		})
	}
	ui.RenderOutput()

	fmt.Fprintln(output, "\nScreen capture demo completed successfully.")
	return nil
}
