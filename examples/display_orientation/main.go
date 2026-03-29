//go:build android

// Command display_orientation demonstrates reading display rotation and
// orientation. It shows how to detect portrait vs landscape and the
// current rotation angle using the Display API.
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
	"github.com/AndroidGoLab/jni/view/display"
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

	wm, err := display.NewWindowManager(ctx)
	if err != nil {
		return fmt.Errorf("NewWindowManager: %w", err)
	}
	defer wm.Close()

	dispObj, err := wm.GetDefaultDisplay()
	if err != nil {
		return fmt.Errorf("getDefaultDisplay: %w", err)
	}
	disp := display.Display{VM: vm, Obj: dispObj}

	fmt.Fprintln(output, "=== Display Orientation ===")
	fmt.Fprintln(output)

	// --- Rotation constants ---
	fmt.Fprintln(output, "Rotation constants:")
	fmt.Fprintf(output, "  ROTATION_0   = %d\n", display.Rotation0)
	fmt.Fprintf(output, "  ROTATION_90  = %d\n", display.Rotation90)
	fmt.Fprintf(output, "  ROTATION_180 = %d\n", display.Rotation180)
	fmt.Fprintf(output, "  ROTATION_270 = %d\n", display.Rotation270)

	// --- Current rotation ---
	rotation, err := disp.GetRotation()
	if err != nil {
		return fmt.Errorf("getRotation: %w", err)
	}
	degrees := 0
	switch rotation {
	case int32(display.Rotation0):
		degrees = 0
	case int32(display.Rotation90):
		degrees = 90
	case int32(display.Rotation180):
		degrees = 180
	case int32(display.Rotation270):
		degrees = 270
	}
	fmt.Fprintf(output, "\ncurrent rotation: %d degrees\n", degrees)

	// --- Orientation (deprecated but useful for demonstration) ---
	orientation, err := disp.GetOrientation()
	if err != nil {
		fmt.Fprintf(output, "getOrientation: %v\n", err)
	} else {
		fmt.Fprintf(output, "orientation (deprecated): %d\n", orientation)
	}

	// --- Detect portrait vs landscape via dimensions ---
	w, err := disp.GetWidth()
	if err != nil {
		return fmt.Errorf("getWidth: %w", err)
	}
	h, err := disp.GetHeight()
	if err != nil {
		return fmt.Errorf("getHeight: %w", err)
	}
	fmt.Fprintf(output, "dimensions: %dx%d\n", w, h)

	orientationStr := "portrait"
	if w > h {
		orientationStr = "landscape"
	} else if w == h {
		orientationStr = "square"
	}
	fmt.Fprintf(output, "detected: %s\n", orientationStr)

	// --- Detailed rotation analysis ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Rotation Analysis ===")
	switch rotation {
	case int32(display.Rotation0):
		fmt.Fprintln(output, "device is in natural orientation")
	case int32(display.Rotation90):
		fmt.Fprintln(output, "device rotated 90 CW from natural")
	case int32(display.Rotation180):
		fmt.Fprintln(output, "device rotated 180 from natural")
	case int32(display.Rotation270):
		fmt.Fprintln(output, "device rotated 270 CW from natural")
	default:
		fmt.Fprintf(output, "unknown rotation value: %d\n", rotation)
	}

	// For phones, ROTATION_0 is typically portrait.
	// For tablets, ROTATION_0 is typically landscape.
	isNaturalPortrait := false
	if rotation == int32(display.Rotation0) && h > w {
		isNaturalPortrait = true
	} else if rotation == int32(display.Rotation90) && w > h {
		isNaturalPortrait = true
	}
	if isNaturalPortrait {
		fmt.Fprintln(output, "natural orientation: portrait (phone)")
	} else {
		fmt.Fprintln(output, "natural orientation: landscape (tablet)")
	}

	// --- Display validity ---
	isValid, err := disp.IsValid()
	if err != nil {
		return fmt.Errorf("isValid: %w", err)
	}
	fmt.Fprintf(output, "\nvalid: %v\n", isValid)

	fmt.Fprintln(output, "\ndisplay_orientation complete")
	return nil
}
