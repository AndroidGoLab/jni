//go:build android

// Command device_health_dashboard aggregates multiple system services:
// battery level, WiFi signal, cellular state, storage free space,
// and display info into one comprehensive status page.
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
	"github.com/AndroidGoLab/jni/net/wifi"
	"github.com/AndroidGoLab/jni/os/battery"
	"github.com/AndroidGoLab/jni/os/storage"
	"github.com/AndroidGoLab/jni/telephony"
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

	fmt.Fprintln(output, "=== Device Health Dashboard ===")

	// --- Battery ---
	fmt.Fprintln(output, "\n[Battery]")
	batMgr, err := battery.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else {
		defer batMgr.Close()

		level, err := batMgr.GetIntProperty(int32(battery.BatteryPropertyCapacity))
		if err != nil {
			fmt.Fprintf(output, "  Level: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Level: %d%%\n", level)
		}

		charging, err := batMgr.IsCharging()
		if err != nil {
			fmt.Fprintf(output, "  Charging: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Charging: %v\n", charging)
		}

		status, err := batMgr.GetIntProperty(int32(battery.BatteryPropertyStatus))
		if err != nil {
			fmt.Fprintf(output, "  Status: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Status: %d\n", status)
		}
	}

	// --- WiFi ---
	fmt.Fprintln(output, "\n[WiFi]")
	wifiMgr, err := wifi.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else {
		defer wifiMgr.Close()

		enabled, err := wifiMgr.IsWifiEnabled()
		if err != nil {
			fmt.Fprintf(output, "  Enabled: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Enabled: %v\n", enabled)
		}

		state, err := wifiMgr.GetWifiState()
		if err != nil {
			fmt.Fprintf(output, "  State: %v\n", err)
		} else {
			fmt.Fprintf(output, "  State: %d\n", state)
		}

		is5g, err := wifiMgr.Is5GHzBandSupported()
		if err != nil {
			fmt.Fprintf(output, "  5GHz: %v\n", err)
		} else {
			fmt.Fprintf(output, "  5GHz supported: %v\n", is5g)
		}
	}

	// --- Telephony ---
	fmt.Fprintln(output, "\n[Telephony]")
	telMgr, err := telephony.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else {
		defer telMgr.Close()

		opName, err := telMgr.GetNetworkOperatorName()
		if err != nil {
			fmt.Fprintf(output, "  Operator: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Operator: %s\n", opName)
		}

		simOp, err := telMgr.GetSimOperatorName()
		if err != nil {
			fmt.Fprintf(output, "  SIM: %v\n", err)
		} else {
			fmt.Fprintf(output, "  SIM: %s\n", simOp)
		}

		phoneType, err := telMgr.GetPhoneType()
		if err != nil {
			fmt.Fprintf(output, "  PhoneType: %v\n", err)
		} else {
			fmt.Fprintf(output, "  PhoneType: %d\n", phoneType)
		}

		dataState, err := telMgr.GetDataState()
		if err != nil {
			fmt.Fprintf(output, "  DataState: %v\n", err)
		} else {
			fmt.Fprintf(output, "  DataState: %d\n", dataState)
		}
	}

	// --- Storage ---
	fmt.Fprintln(output, "\n[Storage]")
	stoMgr, err := storage.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else {
		defer stoMgr.Close()
		fmt.Fprintln(output, "  StorageManager: obtained")

		primaryVol, err := stoMgr.GetPrimaryStorageVolume()
		if err != nil {
			fmt.Fprintf(output, "  Primary: %v\n", err)
		} else if primaryVol != nil && primaryVol.Ref() != 0 {
			fmt.Fprintln(output, "  PrimaryVolume: available")
		}
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
			fmt.Fprintf(output, "  DefaultDisplay: %v\n", err)
		} else if dispObj != nil && dispObj.Ref() != 0 {
			disp := display.Display{VM: vm, Obj: dispObj}

			name, err := disp.GetName()
			if err == nil {
				fmt.Fprintf(output, "  Name: %s\n", name)
			}

			w, _ := disp.GetWidth()
			h, _ := disp.GetHeight()
			fmt.Fprintf(output, "  Size: %dx%d\n", w, h)

			refresh, _ := disp.GetRefreshRate()
			fmt.Fprintf(output, "  Refresh: %.1f Hz\n", refresh)

			rotation, _ := disp.GetRotation()
			fmt.Fprintf(output, "  Rotation: %d\n", rotation)
		}
	}

	fmt.Fprintln(output, "\nDashboard complete.")
	return nil
}
