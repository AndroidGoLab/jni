//go:build android

// Command power_wake_lock demonstrates the PowerManager wake lock
// lifecycle: create a wake lock, acquire it with a timeout, check
// the held state and screen state, then release it.
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
	"github.com/AndroidGoLab/jni/os/power"
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

	mgr, err := power.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("power.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Power Wake Lock ===")
	fmt.Fprintln(output)

	// --- Wake lock level support ---
	fmt.Fprintln(output, "Wake lock level support:")
	levels := []struct {
		name string
		val  int
	}{
		{"PARTIAL_WAKE_LOCK", power.PartialWakeLock},
		{"SCREEN_DIM_WAKE_LOCK", power.ScreenDimWakeLock},
		{"SCREEN_BRIGHT_WAKE_LOCK", power.ScreenBrightWakeLock},
		{"FULL_WAKE_LOCK", power.FullWakeLock},
		{"PROXIMITY_SCREEN_OFF", power.ProximityScreenOffWakeLock},
	}
	for _, l := range levels {
		supported, err := mgr.IsWakeLockLevelSupported(int32(l.val))
		if err != nil {
			fmt.Fprintf(output, "  %s: error: %v\n", l.name, err)
		} else {
			fmt.Fprintf(output, "  %s: %v\n", l.name, supported)
		}
	}

	// --- Screen state ---
	fmt.Fprintln(output)
	screenOn, err := mgr.IsScreenOn()
	if err != nil {
		fmt.Fprintf(output, "screen on: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "screen on: %v\n", screenOn)
	}

	interactive, err := mgr.IsInteractive()
	if err != nil {
		fmt.Fprintf(output, "interactive: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "interactive: %v\n", interactive)
	}

	// --- Create wake lock ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Wake lock lifecycle:")

	wlObj, err := mgr.NewWakeLock(int32(power.PartialWakeLock), "go-jni-demo")
	if err != nil {
		fmt.Fprintf(output, "  newWakeLock: %v\n", err)
		return nil
	}
	fmt.Fprintln(output, "  created: PARTIAL_WAKE_LOCK tag=go-jni-demo")

	wl := power.ManagerWakeLock{VM: vm, Obj: wlObj}

	// Check held state before acquire.
	held, err := wl.IsHeld()
	if err != nil {
		fmt.Fprintf(output, "  isHeld (before): error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  isHeld (before): %v\n", held)
	}

	// Acquire with 5 second timeout.
	err = wl.Acquire1_1(5000)
	if err != nil {
		fmt.Fprintf(output, "  acquire(5000ms): %v\n", err)
	} else {
		fmt.Fprintln(output, "  acquire(5000ms): OK")
	}

	// Check held state after acquire.
	held, err = wl.IsHeld()
	if err != nil {
		fmt.Fprintf(output, "  isHeld (after): error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  isHeld (after): %v\n", held)
	}

	// ToString.
	desc, err := wl.ToString()
	if err != nil {
		fmt.Fprintf(output, "  toString: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  toString: %s\n", desc)
	}

	// Release.
	err = wl.Release0()
	if err != nil {
		fmt.Fprintf(output, "  release: %v\n", err)
	} else {
		fmt.Fprintln(output, "  release: OK")
	}

	// Check held state after release.
	held, err = wl.IsHeld()
	if err != nil {
		fmt.Fprintf(output, "  isHeld (released): error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  isHeld (released): %v\n", held)
	}

	// --- Constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Wake lock constants:")
	fmt.Fprintf(output, "  PARTIAL          = %d\n", power.PartialWakeLock)
	fmt.Fprintf(output, "  SCREEN_DIM       = %d\n", power.ScreenDimWakeLock)
	fmt.Fprintf(output, "  SCREEN_BRIGHT    = %d\n", power.ScreenBrightWakeLock)
	fmt.Fprintf(output, "  FULL             = %d\n", power.FullWakeLock)
	fmt.Fprintf(output, "  ACQUIRE_WAKEUP   = %d\n", power.AcquireCausesWakeup)
	fmt.Fprintf(output, "  ON_AFTER_RELEASE = %d\n", power.OnAfterRelease)

	return nil
}
