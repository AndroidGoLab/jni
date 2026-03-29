//go:build android

// Command display_multi shows the default display information obtained via
// the WindowManager system service. It reports display ID, name, state,
// size, refresh rate, and validity using the typed Display wrapper.
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

func stateString(state int32) string {
	switch int(state) {
	case display.StateOff:
		return "OFF"
	case display.StateOn:
		return "ON"
	case display.StateDoze:
		return "DOZE"
	case display.StateDozeSuspend:
		return "DOZE_SUSPEND"
	case display.StateOnSuspend:
		return "ON_SUSPEND"
	case display.StateVr:
		return "VR"
	case display.StateUnknown:
		return "UNKNOWN"
	default:
		return fmt.Sprintf("? (%d)", state)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Multi-Display Info ===")

	// --- State constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Display state constants:")
	fmt.Fprintf(output, "  STATE_UNKNOWN      = %d\n", display.StateUnknown)
	fmt.Fprintf(output, "  STATE_OFF          = %d\n", display.StateOff)
	fmt.Fprintf(output, "  STATE_ON           = %d\n", display.StateOn)
	fmt.Fprintf(output, "  STATE_DOZE         = %d\n", display.StateDoze)
	fmt.Fprintf(output, "  STATE_DOZE_SUSPEND = %d\n", display.StateDozeSuspend)
	fmt.Fprintf(output, "  STATE_ON_SUSPEND   = %d\n", display.StateOnSuspend)
	fmt.Fprintf(output, "  STATE_VR           = %d\n", display.StateVr)

	// --- Default display via WindowManager ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Default Display (WindowManager) ===")
	wm, err := display.NewWindowManager(ctx)
	if err != nil {
		return fmt.Errorf("NewWindowManager: %w", err)
	}
	defer wm.Close()

	defObj, err := wm.GetDefaultDisplay()
	if err != nil {
		return fmt.Errorf("getDefaultDisplay: %w", err)
	}
	defDisp := display.Display{VM: vm, Obj: defObj}

	defID, _ := defDisp.GetDisplayId()
	defName, _ := defDisp.GetName()
	defState, _ := defDisp.GetState()
	w, _ := defDisp.GetWidth()
	h, _ := defDisp.GetHeight()
	refreshRate, _ := defDisp.GetRefreshRate()
	isValid, _ := defDisp.IsValid()
	rotation, _ := defDisp.GetRotation()

	fmt.Fprintf(output, "  ID:       %d\n", defID)
	fmt.Fprintf(output, "  name:     %s\n", defName)
	fmt.Fprintf(output, "  state:    %s\n", stateString(defState))
	fmt.Fprintf(output, "  size:     %dx%d\n", w, h)
	fmt.Fprintf(output, "  refresh:  %.1f Hz\n", refreshRate)
	fmt.Fprintf(output, "  rotation: %d\n", rotation)
	fmt.Fprintf(output, "  valid:    %v\n", isValid)

	isHdr, err := defDisp.IsHdr()
	if err != nil {
		fmt.Fprintf(output, "  HDR:      %v\n", err)
	} else {
		fmt.Fprintf(output, "  HDR:      %v\n", isHdr)
	}

	isWideGamut, err := defDisp.IsWideColorGamut()
	if err != nil {
		fmt.Fprintf(output, "  WideGamut: %v\n", err)
	} else {
		fmt.Fprintf(output, "  WideGamut: %v\n", isWideGamut)
	}

	fmt.Fprintln(output, "\ndisplay_multi complete")
	return nil
}
