//go:build android

// Command bt_mesh_relay demonstrates the BLE advertising and scanning API
// surfaces using the bluetooth and bluetooth/le typed wrapper packages. It
// obtains both the LE scanner and advertiser, reports advertising capabilities,
// and displays the constants relevant to mesh relay operation.
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

	// --- Advertise mode constants ---
	fmt.Fprintln(output, "=== Advertise mode constants ===")
	fmt.Fprintf(output, "  AdvertiseModeLowPower   = %d\n", le.AdvertiseModeLowPower)
	fmt.Fprintf(output, "  AdvertiseModeBalanced   = %d\n", le.AdvertiseModeBalanced)
	fmt.Fprintf(output, "  AdvertiseModeLowLatency = %d\n", le.AdvertiseModeLowLatency)

	fmt.Fprintln(output, "=== Advertise TX power constants ===")
	fmt.Fprintf(output, "  AdvertiseTxPowerUltraLow = %d\n", le.AdvertiseTxPowerUltraLow)
	fmt.Fprintf(output, "  AdvertiseTxPowerLow      = %d\n", le.AdvertiseTxPowerLow)
	fmt.Fprintf(output, "  AdvertiseTxPowerMedium   = %d\n", le.AdvertiseTxPowerMedium)
	fmt.Fprintf(output, "  AdvertiseTxPowerHigh     = %d\n", le.AdvertiseTxPowerHigh)

	fmt.Fprintln(output, "=== Scan mode constants ===")
	fmt.Fprintf(output, "  ScanModeLowPower   = %d\n", le.ScanModeLowPower)
	fmt.Fprintf(output, "  ScanModeBalanced   = %d\n", le.ScanModeBalanced)
	fmt.Fprintf(output, "  ScanModeLowLatency = %d\n", le.ScanModeLowLatency)

	fmt.Fprintln(output, "=== BLE mesh data type constants ===")
	fmt.Fprintf(output, "  DataTypeMeshMessage = %d\n", le.DataTypeMeshMessage)
	fmt.Fprintf(output, "  DataTypeMeshBeacon  = %d\n", le.DataTypeMeshBeacon)
	fmt.Fprintf(output, "  DataTypePbAdv       = %d\n", le.DataTypePbAdv)

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

	// --- Advertising capabilities ---
	multiAdv, err := adapter.IsMultipleAdvertisementSupported()
	if err != nil {
		fmt.Fprintf(output, "IsMultipleAdvertisementSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Multiple advertisement supported: %v\n", multiAdv)
	}

	leExtAdv, err := adapter.IsLeExtendedAdvertisingSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeExtendedAdvertisingSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "LE extended advertising supported: %v\n", leExtAdv)
	}

	maxAdvLen, err := adapter.GetLeMaximumAdvertisingDataLength()
	if err != nil {
		fmt.Fprintf(output, "GetLeMaximumAdvertisingDataLength error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Max advertising data length: %d bytes\n", maxAdvLen)
	}

	// --- Get LE Scanner ---
	scannerObj, err := adapter.GetBluetoothLeScanner()
	if err != nil {
		return fmt.Errorf("GetBluetoothLeScanner: %w", err)
	}
	if scannerObj == nil {
		fmt.Fprintln(output, "BLE scanner not available")
		return nil
	}
	_ = &le.BluetoothLeScanner{VM: vm, Obj: scannerObj}
	fmt.Fprintln(output, "BLE scanner obtained: OK")

	// --- Get LE Advertiser ---
	advertiserObj, err := adapter.GetBluetoothLeAdvertiser()
	if err != nil {
		return fmt.Errorf("GetBluetoothLeAdvertiser: %w", err)
	}
	if advertiserObj == nil {
		fmt.Fprintln(output, "BLE advertiser not available")
		return nil
	}
	_ = &le.BluetoothLeAdvertiser{VM: vm, Obj: advertiserObj}
	fmt.Fprintln(output, "BLE advertiser obtained: OK")

	// --- Show mesh relay API surface ---
	fmt.Fprintln(output, "\n=== Mesh relay typed wrapper API ===")
	fmt.Fprintln(output, "  BluetoothLeScanner: StartScan, StopScan, FlushPendingScanResults")
	fmt.Fprintln(output, "  BluetoothLeAdvertiser: StartAdvertising, StopAdvertising,")
	fmt.Fprintln(output, "    StartAdvertisingSet, StopAdvertisingSet")
	fmt.Fprintln(output, "  ScanSettingsBuilder: SetScanMode, SetCallbackType, SetMatchMode,")
	fmt.Fprintln(output, "    SetNumOfMatches, SetReportDelay, SetPhy, SetLegacy, Build")
	fmt.Fprintln(output, "  AdvertiseSettingsBuilder: SetAdvertiseMode, SetTxPowerLevel,")
	fmt.Fprintln(output, "    SetConnectable, SetDiscoverable, SetTimeout, Build")
	fmt.Fprintln(output, "  AdvertiseDataBuilder: SetIncludeDeviceName, SetIncludeTxPowerLevel,")
	fmt.Fprintln(output, "    AddServiceUuid, AddServiceData, AddManufacturerData, Build")

	fmt.Fprintln(output, "\nMesh relay capability check completed.")
	fmt.Fprintln(output, "No errors occurred during mesh relay demo.")

	return nil
}
