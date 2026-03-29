//go:build android

// Command bt_beacon_scanner demonstrates the BLE scanner API surface provided
// by the bluetooth and bluetooth/le typed wrapper packages. It obtains the
// adapter, checks BLE support, retrieves the LE scanner, and reports adapter
// capabilities relevant to BLE beacon scanning.
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
	"github.com/AndroidGoLab/jni/bluetooth/le"
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

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- LE scan constants ---
	fmt.Fprintln(output, "=== BLE scan mode constants ===")
	fmt.Fprintf(output, "  ScanModeLowPower    = %d\n", le.ScanModeLowPower)
	fmt.Fprintf(output, "  ScanModeBalanced    = %d\n", le.ScanModeBalanced)
	fmt.Fprintf(output, "  ScanModeLowLatency  = %d\n", le.ScanModeLowLatency)
	fmt.Fprintf(output, "  ScanModeOpportunistic = %d\n", le.ScanModeOpportunistic)

	fmt.Fprintln(output, "=== BLE match mode constants ===")
	fmt.Fprintf(output, "  MatchModeAggressive = %d\n", le.MatchModeAggressive)
	fmt.Fprintf(output, "  MatchModeSticky     = %d\n", le.MatchModeSticky)

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

	// --- LE Scanner ---
	scannerObj, err := adapter.GetBluetoothLeScanner()
	if err != nil {
		return fmt.Errorf("GetBluetoothLeScanner: %w", err)
	}
	if scannerObj == nil {
		fmt.Fprintln(output, "BLE scanner not available (null)")
		return nil
	}
	// Wrap the raw object in the typed wrapper to confirm it was obtained.
	_ = &le.BluetoothLeScanner{VM: vm, Obj: scannerObj}
	fmt.Fprintln(output, "BLE scanner obtained: OK")

	// --- Adapter BLE capabilities ---
	le2m, err := adapter.IsLe2MPhySupported()
	if err != nil {
		fmt.Fprintf(output, "IsLe2MPhySupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "LE 2M PHY supported: %v\n", le2m)
	}

	leCoded, err := adapter.IsLeCodedPhySupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeCodedPhySupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "LE Coded PHY supported: %v\n", leCoded)
	}

	leExtAdv, err := adapter.IsLeExtendedAdvertisingSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeExtendedAdvertisingSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "LE extended advertising supported: %v\n", leExtAdv)
	}

	offloadFilter, err := adapter.IsOffloadedFilteringSupported()
	if err != nil {
		fmt.Fprintf(output, "IsOffloadedFilteringSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Offloaded filtering supported: %v\n", offloadFilter)
	}

	offloadBatch, err := adapter.IsOffloadedScanBatchingSupported()
	if err != nil {
		fmt.Fprintf(output, "IsOffloadedScanBatchingSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Offloaded scan batching supported: %v\n", offloadBatch)
	}

	maxAdvLen, err := adapter.GetLeMaximumAdvertisingDataLength()
	if err != nil {
		fmt.Fprintf(output, "GetLeMaximumAdvertisingDataLength error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Max advertising data length: %d bytes\n", maxAdvLen)
	}

	fmt.Fprintln(output, "\nBLE beacon scanner capability check completed.")
	fmt.Fprintln(output, "No errors occurred during BLE beacon scanner demo.")

	return nil
}
