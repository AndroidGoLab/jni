//go:build android

// Command projection demonstrates using the MediaProjection API.
// It obtains the MediaProjectionManager system service and creates
// a screen capture intent to verify the API is functional.
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
	"github.com/AndroidGoLab/jni/media/projection"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
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

	fmt.Fprintln(output, "=== MediaProjection ===")

	// --- Obtain MediaProjectionManager ---
	mgr, err := projection.NewMediaProjectionManager(ctx)
	if err != nil {
		return fmt.Errorf("NewMediaProjectionManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "Manager: obtained OK")

	// --- Create a screen capture intent ---
	// This exercises the primary API entry point. The returned Intent
	// would normally be passed to startActivityForResult to get user
	// consent before a projection can begin.
	captureIntent, err := mgr.CreateScreenCaptureIntent0()
	if err != nil {
		fmt.Fprintf(output, "CreateScreenCaptureIntent: %v\n", err)
	} else if captureIntent == nil || captureIntent.Ref() == 0 {
		fmt.Fprintln(output, "CaptureIntent: null")
	} else {
		fmt.Fprintln(output, "CaptureIntent: created OK")

		// Inspect the intent via toString()
		vm.Do(func(env *jni.Env) error {
			intentCls := env.GetObjectClass(captureIntent)
			toStrMid, err := env.GetMethodID(intentCls, "toString", "()Ljava/lang/String;")
			if err != nil {
				return nil
			}
			strObj, err := env.CallObjectMethod(captureIntent, toStrMid)
			if err != nil {
				return nil
			}
			s := env.GoString((*jni.String)(unsafe.Pointer(strObj)))
			fmt.Fprintf(output, "  Intent: %s\n", s)

			// Get the intent action
			getActionMid, err := env.GetMethodID(intentCls, "getAction", "()Ljava/lang/String;")
			if err != nil {
				return nil
			}
			actionObj, err := env.CallObjectMethod(captureIntent, getActionMid)
			if err != nil {
				return nil
			}
			if actionObj != nil && actionObj.Ref() != 0 {
				action := env.GoString((*jni.String)(unsafe.Pointer(actionObj)))
				fmt.Fprintf(output, "  Action: %s\n", action)
			}

			return nil
		})
	}

	// --- Describe available projection types ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Available types:")

	// VirtualDisplay methods (once a projection is obtained):
	fmt.Fprintln(output, "  VirtualDisplay:")
	fmt.Fprintln(output, "    GetDisplay()")
	fmt.Fprintln(output, "    GetSurface()")
	fmt.Fprintln(output, "    Release()")
	fmt.Fprintln(output, "    Resize(w, h, dpi)")
	fmt.Fprintln(output, "    SetRotation(rot)")
	fmt.Fprintln(output, "    SetSurface(s)")

	// MediaProjection methods:
	fmt.Fprintln(output, "  MediaProjection:")
	fmt.Fprintln(output, "    Stop()")
	fmt.Fprintln(output, "    UnregisterCallback(cb)")

	// --- Workflow summary ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Screen capture workflow:")
	fmt.Fprintln(output, "1. CreateScreenCaptureIntent")
	fmt.Fprintln(output, "2. startActivityForResult")
	fmt.Fprintln(output, "   (user consent required)")
	fmt.Fprintln(output, "3. GetMediaProjection")
	fmt.Fprintln(output, "4. createVirtualDisplay")
	fmt.Fprintln(output, "5. Read from Surface")
	fmt.Fprintln(output, "6. Stop / Release")

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Projection example complete.")
	return nil
}
