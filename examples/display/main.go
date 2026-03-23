//go:build android

// Command display demonstrates the Android Display and WindowManager API.
// It queries the default display for size, refresh rate, rotation, and state.
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

	// Get the default display object and wrap it.
	dispObj, err := wm.GetDefaultDisplay()
	if err != nil {
		return fmt.Errorf("getDefaultDisplay: %w", err)
	}
	disp := display.Display{VM: vm, Obj: dispObj}

	fmt.Fprintln(output, "=== Display Info ===")
	fmt.Fprintln(output)

	name, err := disp.GetName()
	if err != nil {
		return fmt.Errorf("getName: %w", err)
	}
	fmt.Fprintf(output, "name: %s\n", name)

	id, err := disp.GetDisplayId()
	if err != nil {
		return fmt.Errorf("getDisplayId: %w", err)
	}
	fmt.Fprintf(output, "id: %d\n", id)

	w, err := disp.GetWidth()
	if err != nil {
		return fmt.Errorf("getWidth: %w", err)
	}
	h, err := disp.GetHeight()
	if err != nil {
		return fmt.Errorf("getHeight: %w", err)
	}
	fmt.Fprintf(output, "size: %dx%d\n", w, h)

	rotation, err := disp.GetRotation()
	if err != nil {
		return fmt.Errorf("getRotation: %w", err)
	}
	fmt.Fprintf(output, "rotation: %d\n", rotation)

	refreshRate, err := disp.GetRefreshRate()
	if err != nil {
		return fmt.Errorf("getRefreshRate: %w", err)
	}
	fmt.Fprintf(output, "refresh: %.1f Hz\n", refreshRate)

	state, err := disp.GetState()
	if err != nil {
		return fmt.Errorf("getState: %w", err)
	}
	stateStr := "unknown"
	switch int(state) {
	case display.StateOff:
		stateStr = "OFF"
	case display.StateOn:
		stateStr = "ON"
	case display.StateDoze:
		stateStr = "DOZE"
	case display.StateDozeSuspend:
		stateStr = "DOZE_SUSPEND"
	case display.StateOnSuspend:
		stateStr = "ON_SUSPEND"
	case display.StateVr:
		stateStr = "VR"
	}
	fmt.Fprintf(output, "state: %s (%d)\n", stateStr, state)

	isValid, err := disp.IsValid()
	if err != nil {
		return fmt.Errorf("isValid: %w", err)
	}
	fmt.Fprintf(output, "valid: %v\n", isValid)

	// Display metrics via toString.
	metricsStr, err := disp.ToString()
	if err != nil {
		return fmt.Errorf("toString: %w", err)
	}
	fmt.Fprintln(output)
	fmt.Fprintf(output, "raw: %s\n", metricsStr)

	return nil
}
