//go:build android

// Command media_screen_capture demonstrates the MediaProjection API setup.
// It obtains the MediaProjectionManager system service, creates a screen
// capture intent, and describes the projection workflow using typed wrappers.
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

	// Obtain MediaProjectionManager.
	mgr, err := projection.NewMediaProjectionManager(ctx)
	if err != nil {
		return fmt.Errorf("NewMediaProjectionManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "Manager: obtained OK")
	ui.RenderOutput()

	// Create a screen capture intent.
	captureIntent, err := mgr.CreateScreenCaptureIntent0()
	if err != nil {
		fmt.Fprintf(output, "CreateScreenCaptureIntent: %v\n", err)
	} else if captureIntent == nil || captureIntent.Ref() == 0 {
		fmt.Fprintln(output, "CaptureIntent: null")
	} else {
		fmt.Fprintln(output, "CaptureIntent: created OK")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(captureIntent)
			return nil
		})
	}
	ui.RenderOutput()

	// Describe the projection API surface.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "API surface:")
	fmt.Fprintln(output, "  Manager methods:")
	fmt.Fprintln(output, "    CreateScreenCaptureIntent()")
	fmt.Fprintln(output, "    GetMediaProjection(code,data)")
	ui.RenderOutput()

	fmt.Fprintln(output, "  MediaProjection methods:")
	fmt.Fprintln(output, "    Stop()")
	fmt.Fprintln(output, "    UnregisterCallback(cb)")
	ui.RenderOutput()

	fmt.Fprintln(output, "  VirtualDisplay methods:")
	fmt.Fprintln(output, "    GetDisplay()")
	fmt.Fprintln(output, "    GetSurface()")
	fmt.Fprintln(output, "    Release()")
	fmt.Fprintln(output, "    Resize(w, h, dpi)")
	fmt.Fprintln(output, "    SetSurface(s)")
	ui.RenderOutput()

	// Workflow summary.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Capture workflow:")
	fmt.Fprintln(output, "1. CreateScreenCaptureIntent")
	fmt.Fprintln(output, "2. startActivityForResult")
	fmt.Fprintln(output, "   (user consent required)")
	fmt.Fprintln(output, "3. GetMediaProjection")
	fmt.Fprintln(output, "4. createVirtualDisplay")
	fmt.Fprintln(output, "5. Read from Surface")
	fmt.Fprintln(output, "6. Stop / Release")

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Screen capture example complete.")
	return nil
}
