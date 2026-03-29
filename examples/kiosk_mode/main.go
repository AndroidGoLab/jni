//go:build android

// Command kiosk_mode is a kiosk mode concept: checks keyguard state,
// sets wake lock to keep screen on, checks power save mode, and gets
// display info. Shows lockdown configuration.
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
	"github.com/AndroidGoLab/jni/os/keyguard"
	"github.com/AndroidGoLab/jni/os/power"
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

	fmt.Fprintln(output, "=== Kiosk Mode ===")

	// --- Keyguard ---
	fmt.Fprintln(output, "\n[Keyguard]")
	kgMgr, err := keyguard.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else {
		defer kgMgr.Close()

		locked, err := kgMgr.IsKeyguardLocked()
		if err != nil {
			fmt.Fprintf(output, "  Locked: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Locked: %v\n", locked)
		}

		secure, err := kgMgr.IsKeyguardSecure()
		if err != nil {
			fmt.Fprintf(output, "  Secure: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Secure: %v\n", secure)
		}

		deviceLocked, err := kgMgr.IsDeviceLocked()
		if err != nil {
			fmt.Fprintf(output, "  DeviceLocked: %v\n", err)
		} else {
			fmt.Fprintf(output, "  DeviceLocked: %v\n", deviceLocked)
		}

		deviceSecure, err := kgMgr.IsDeviceSecure()
		if err != nil {
			fmt.Fprintf(output, "  DeviceSecure: %v\n", err)
		} else {
			fmt.Fprintf(output, "  DeviceSecure: %v\n", deviceSecure)
		}
	}

	// --- Power / Wake Lock ---
	fmt.Fprintln(output, "\n[Power]")
	pwrMgr, err := power.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else {
		defer pwrMgr.Close()

		interactive, err := pwrMgr.IsInteractive()
		if err != nil {
			fmt.Fprintf(output, "  Interactive: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Interactive: %v\n", interactive)
		}

		screenOn, err := pwrMgr.IsScreenOn()
		if err != nil {
			fmt.Fprintf(output, "  ScreenOn: %v\n", err)
		} else {
			fmt.Fprintf(output, "  ScreenOn: %v\n", screenOn)
		}

		powerSave, err := pwrMgr.IsPowerSaveMode()
		if err != nil {
			fmt.Fprintf(output, "  PowerSave: %v\n", err)
		} else {
			fmt.Fprintf(output, "  PowerSave: %v\n", powerSave)
		}

		idleMode, err := pwrMgr.IsDeviceIdleMode()
		if err != nil {
			fmt.Fprintf(output, "  IdleMode: %v\n", err)
		} else {
			fmt.Fprintf(output, "  IdleMode: %v\n", idleMode)
		}

		// Create a wake lock (PARTIAL_WAKE_LOCK keeps CPU on).
		wakeLockObj, err := pwrMgr.NewWakeLock(int32(power.PartialWakeLock), "kiosk_mode:keep_alive")
		if err != nil {
			fmt.Fprintf(output, "  WakeLock: %v\n", err)
		} else if wakeLockObj != nil && wakeLockObj.Ref() != 0 {
			fmt.Fprintln(output, "  WakeLock: created OK")
			// Release it immediately (we're demonstrating the API).
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(wakeLockObj)
				return nil
			})
		}

		// Wake lock level constants.
		fmt.Fprintln(output, "\n  Wake Lock Levels:")
		fmt.Fprintf(output, "    PARTIAL: %d\n", power.PartialWakeLock)
		fmt.Fprintf(output, "    SCREEN_DIM: %d\n", power.ScreenDimWakeLock)
		fmt.Fprintf(output, "    SCREEN_BRIGHT: %d\n", power.ScreenBrightWakeLock)
		fmt.Fprintf(output, "    FULL: %d\n", power.FullWakeLock)
	}

	// --- Display ---
	fmt.Fprintln(output, "\n[Display]")
	wm, err := display.NewWindowManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else {
		defer wm.Close()

		dispObj, err := wm.GetDefaultDisplay()
		if err != nil {
			fmt.Fprintf(output, "  Display: %v\n", err)
		} else if dispObj != nil && dispObj.Ref() != 0 {
			disp := display.Display{VM: vm, Obj: dispObj}

			name, _ := disp.GetName()
			fmt.Fprintf(output, "  Name: %s\n", name)

			w, _ := disp.GetWidth()
			h, _ := disp.GetHeight()
			fmt.Fprintf(output, "  Size: %dx%d\n", w, h)

			refresh, _ := disp.GetRefreshRate()
			fmt.Fprintf(output, "  Refresh: %.1f Hz\n", refresh)

			rotation, _ := disp.GetRotation()
			fmt.Fprintf(output, "  Rotation: %d\n", rotation)

			state, _ := disp.GetState()
			fmt.Fprintf(output, "  State: %d\n", state)
		}
	}

	fmt.Fprintln(output, "\nKiosk mode complete.")
	return nil
}
