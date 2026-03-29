//go:build android

// Command bt_classic_discovery demonstrates classic Bluetooth discovery using
// the bluetooth typed wrapper package. It obtains the adapter, starts
// discovery, checks status, and cancels discovery. It also displays adapter
// state, scan mode, and bond/device-type constants.
//
// Required permissions (Android 12+): BLUETOOTH_SCAN, BLUETOOTH_CONNECT,
// BLUETOOTH_ADVERTISE.
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
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/bluetooth"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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

// scanModeName returns a human-readable name for an adapter scan mode constant.
func scanModeName(m int32) string {
	switch int(m) {
	case bluetooth.ScanModeNone:
		return "None"
	case bluetooth.ScanModeConnectable:
		return "Connectable"
	case bluetooth.ScanModeConnectableDiscoverable:
		return "ConnectableDiscoverable"
	default:
		return fmt.Sprintf("Unknown(%d)", m)
	}
}

// adapterStateName returns a human-readable name for an adapter state constant.
func adapterStateName(s int32) string {
	switch int(s) {
	case bluetooth.StateOff:
		return "Off"
	case bluetooth.StateTurningOn:
		return "TurningOn"
	case bluetooth.StateOn:
		return "On"
	case bluetooth.StateTurningOff:
		return "TurningOff"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- Constants ---
	fmt.Fprintln(output, "=== Device type constants ===")
	fmt.Fprintf(output, "  DeviceTypeClassic = %d\n", bluetooth.DeviceTypeClassic)
	fmt.Fprintf(output, "  DeviceTypeLe      = %d\n", bluetooth.DeviceTypeLe)
	fmt.Fprintf(output, "  DeviceTypeDual    = %d\n", bluetooth.DeviceTypeDual)

	fmt.Fprintln(output, "=== Bond state constants ===")
	fmt.Fprintf(output, "  BondNone    = %d\n", bluetooth.BondNone)
	fmt.Fprintf(output, "  BondBonding = %d\n", bluetooth.BondBonding)
	fmt.Fprintf(output, "  BondBonded  = %d\n", bluetooth.BondBonded)

	fmt.Fprintln(output, "=== Adapter state constants ===")
	fmt.Fprintf(output, "  StateOff       = %d\n", bluetooth.StateOff)
	fmt.Fprintf(output, "  StateTurningOn = %d\n", bluetooth.StateTurningOn)
	fmt.Fprintf(output, "  StateOn        = %d\n", bluetooth.StateOn)
	fmt.Fprintf(output, "  StateTurningOff = %d\n", bluetooth.StateTurningOff)

	// --- Adapter ---
	adapter, err := bluetooth.NewAdapter(ctx)
	if err != nil {
		return fmt.Errorf("bluetooth.NewAdapter: %w", err)
	}
	defer adapter.Close()

	enabled, err := adapter.IsEnabled()
	if err != nil {
		return fmt.Errorf("IsEnabled: %w", err)
	}
	fmt.Fprintf(output, "\nBluetooth enabled: %v\n", enabled)
	if !enabled {
		fmt.Fprintln(output, "Bluetooth is off; enable it in Settings.")
		return nil
	}

	name, err := adapter.GetName()
	if err != nil {
		return fmt.Errorf("GetName: %w", err)
	}
	fmt.Fprintf(output, "Adapter name: %s\n", name)

	state, err := adapter.GetState()
	if err != nil {
		return fmt.Errorf("GetState: %w", err)
	}
	fmt.Fprintf(output, "Adapter state: %s (%d)\n", adapterStateName(state), state)

	scanMode, err := adapter.GetScanMode()
	if err != nil {
		return fmt.Errorf("GetScanMode: %w", err)
	}
	fmt.Fprintf(output, "Scan mode: %s (%d)\n", scanModeName(scanMode), scanMode)

	// --- Start discovery ---
	started, err := adapter.StartDiscovery()
	if err != nil {
		fmt.Fprintf(output, "StartDiscovery error (may need location permission): %v\n", err)
	} else {
		fmt.Fprintf(output, "\nStartDiscovery result: %v\n", started)
	}

	isDiscovering, err := adapter.IsDiscovering()
	if err != nil {
		fmt.Fprintf(output, "IsDiscovering error: %v\n", err)
	} else {
		fmt.Fprintf(output, "IsDiscovering: %v\n", isDiscovering)
	}

	// Wait briefly to let discovery run.
	time.Sleep(3 * time.Second)

	isDiscovering, err = adapter.IsDiscovering()
	if err != nil {
		fmt.Fprintf(output, "IsDiscovering (after wait) error: %v\n", err)
	} else {
		fmt.Fprintf(output, "IsDiscovering (after 3s): %v\n", isDiscovering)
	}

	// --- Cancel discovery ---
	cancelled, err := adapter.CancelDiscovery()
	if err != nil {
		fmt.Fprintf(output, "CancelDiscovery error: %v\n", err)
	} else {
		fmt.Fprintf(output, "\nCancelDiscovery result: %v\n", cancelled)
	}

	fmt.Fprintln(output, "No errors occurred during classic Bluetooth discovery.")

	return nil
}
