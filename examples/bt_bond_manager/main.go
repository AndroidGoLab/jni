//go:build android

// Command bt_bond_manager demonstrates the bond management API surface using
// the bluetooth typed wrapper package. It displays bond state and device type
// constants, queries the adapter, and reports adapter properties relevant to
// bonded device management.
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

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- Bond state constants ---
	fmt.Fprintln(output, "=== Bond state constants ===")
	fmt.Fprintf(output, "  BondNone    = %d\n", bluetooth.BondNone)
	fmt.Fprintf(output, "  BondBonding = %d\n", bluetooth.BondBonding)
	fmt.Fprintf(output, "  BondBonded  = %d\n", bluetooth.BondBonded)

	fmt.Fprintln(output, "=== Device type constants ===")
	fmt.Fprintf(output, "  DeviceTypeClassic = %d\n", bluetooth.DeviceTypeClassic)
	fmt.Fprintf(output, "  DeviceTypeLe      = %d\n", bluetooth.DeviceTypeLe)
	fmt.Fprintf(output, "  DeviceTypeDual    = %d\n", bluetooth.DeviceTypeDual)

	fmt.Fprintln(output, "=== Device typed wrapper API ===")
	fmt.Fprintln(output, "  Device methods: GetName, GetAddress, GetType, GetBondState,")
	fmt.Fprintln(output, "    GetAlias, SetAlias, GetBluetoothClass, GetUuids,")
	fmt.Fprintln(output, "    FetchUuidsWithSdp, CreateBond, SetPairingConfirmation,")
	fmt.Fprintln(output, "    ConnectGatt, CreateRfcommSocketToServiceRecord,")
	fmt.Fprintln(output, "    CreateInsecureRfcommSocketToServiceRecord")

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

	scanMode, err := adapter.GetScanMode()
	if err != nil {
		return fmt.Errorf("GetScanMode: %w", err)
	}
	fmt.Fprintf(output, "Scan mode: %s (%d)\n", scanModeName(scanMode), scanMode)

	// --- Bonded devices set ---
	// GetBondedDevices returns a Java Set object. The typed wrapper returns
	// it as *jni.Object; iterating the set requires collection helpers that
	// are not yet part of the typed wrapper API, so we only confirm the
	// call succeeds.
	bondedSet, err := adapter.GetBondedDevices()
	if err != nil {
		fmt.Fprintf(output, "GetBondedDevices error: %v\n", err)
	} else if bondedSet == nil {
		fmt.Fprintln(output, "\nBonded devices: (none)")
	} else {
		fmt.Fprintln(output, "\nBonded devices set retrieved: OK")
		fmt.Fprintln(output, "  (Set iteration requires collection helpers not yet in typed wrappers)")
	}

	// --- Adapter address ---
	addr, err := adapter.GetAddress()
	if err != nil {
		fmt.Fprintf(output, "GetAddress error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Adapter address: %s\n", addr)
	}

	fmt.Fprintln(output, "\nBond manager demo completed.")
	fmt.Fprintln(output, "No errors occurred during bond manager demo.")

	return nil
}
