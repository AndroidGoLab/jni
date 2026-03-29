//go:build android

// Command display_inspector demonstrates reading comprehensive display
// information: resolution, density, refresh rate, rotation, state, and
// HDR capabilities via the WindowManager and Display APIs.
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

	fmt.Fprintln(output, "=== Display Inspector ===")
	fmt.Fprintln(output)

	// --- Basic info ---
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

	// --- Resolution ---
	w, err := disp.GetWidth()
	if err != nil {
		return fmt.Errorf("getWidth: %w", err)
	}
	h, err := disp.GetHeight()
	if err != nil {
		return fmt.Errorf("getHeight: %w", err)
	}
	fmt.Fprintf(output, "resolution: %dx%d\n", w, h)

	// --- Density via DisplayMetrics ---
	metrics, err := display.NewMetrics(vm)
	if err != nil {
		fmt.Fprintf(output, "NewMetrics: %v\n", err)
	} else {
		if err := disp.GetMetrics(metrics.Obj); err != nil {
			fmt.Fprintf(output, "getMetrics: %v\n", err)
		} else {
			metricsStr, err := metrics.ToString()
			if err != nil {
				fmt.Fprintf(output, "metrics.toString: %v\n", err)
			} else {
				fmt.Fprintf(output, "metrics: %s\n", metricsStr)
			}
		}
	}

	// --- Refresh rate ---
	refreshRate, err := disp.GetRefreshRate()
	if err != nil {
		return fmt.Errorf("getRefreshRate: %w", err)
	}
	fmt.Fprintf(output, "refresh rate: %.1f Hz\n", refreshRate)

	// --- Rotation ---
	rotation, err := disp.GetRotation()
	if err != nil {
		return fmt.Errorf("getRotation: %w", err)
	}
	rotStr := "unknown"
	switch rotation {
	case int32(display.Rotation0):
		rotStr = "0 (natural)"
	case int32(display.Rotation90):
		rotStr = "90"
	case int32(display.Rotation180):
		rotStr = "180"
	case int32(display.Rotation270):
		rotStr = "270"
	}
	fmt.Fprintf(output, "rotation: %s\n", rotStr)

	// --- State ---
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

	// --- HDR capabilities ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== HDR Capabilities ===")
	hdrObj, err := disp.GetHdrCapabilities()
	if err != nil {
		fmt.Fprintf(output, "getHdrCapabilities: %v\n", err)
	} else if hdrObj == nil || hdrObj.Ref() == 0 {
		fmt.Fprintln(output, "HDR: not available")
	} else {
		hdr := display.HdrCapabilities{VM: vm, Obj: hdrObj}

		isHdr, err := disp.IsHdr()
		if err != nil {
			fmt.Fprintf(output, "isHdr: %v\n", err)
		} else {
			fmt.Fprintf(output, "HDR supported: %v\n", isHdr)
		}

		maxLum, err := hdr.GetDesiredMaxLuminance()
		if err != nil {
			fmt.Fprintf(output, "maxLuminance: %v\n", err)
		} else {
			fmt.Fprintf(output, "max luminance: %.1f nits\n", maxLum)
		}

		avgLum, err := hdr.GetDesiredMaxAverageLuminance()
		if err != nil {
			fmt.Fprintf(output, "avgLuminance: %v\n", err)
		} else {
			fmt.Fprintf(output, "max avg luminance: %.1f nits\n", avgLum)
		}

		minLum, err := hdr.GetDesiredMinLuminance()
		if err != nil {
			fmt.Fprintf(output, "minLuminance: %v\n", err)
		} else {
			fmt.Fprintf(output, "min luminance: %.4f nits\n", minLum)
		}

		hdrStr, err := hdr.ToString()
		if err != nil {
			fmt.Fprintf(output, "hdr.toString: %v\n", err)
		} else {
			fmt.Fprintf(output, "raw: %s\n", hdrStr)
		}
	}

	// --- HDR type constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "HDR type constants:")
	fmt.Fprintf(output, "  Dolby Vision = %d\n", display.HdrTypeDolbyVision)
	fmt.Fprintf(output, "  HDR10        = %d\n", display.HdrTypeHdr10)
	fmt.Fprintf(output, "  HDR10+       = %d\n", display.HdrTypeHdr10Plus)
	fmt.Fprintf(output, "  HLG          = %d\n", display.HdrTypeHlg)

	// --- Validity ---
	isValid, err := disp.IsValid()
	if err != nil {
		return fmt.Errorf("isValid: %w", err)
	}
	fmt.Fprintf(output, "\nvalid: %v\n", isValid)

	fmt.Fprintln(output, "\ndisplay_inspector complete")
	return nil
}
