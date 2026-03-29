//go:build android

// Command bt_indoor_position demonstrates BLE scanning capabilities relevant
// to indoor positioning using the bluetooth and bluetooth/le typed wrapper
// packages. It checks adapter BLE support, reports PHY capabilities, and
// displays the scan and advertising constants used for indoor positioning.
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
	"math"
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

// estimateDistance computes distance in meters from RSSI using the
// log-distance path-loss model.
//
//	distance = 10^((txPower - rssi) / (10 * n))
//
// txPower is the expected RSSI at 1 meter (typically -59 dBm for BLE).
// n is the path-loss exponent (2.0 for free space, 2.5-4.0 indoors).
func estimateDistance(rssi int32, txPower float64, n float64) float64 {
	return math.Pow(10, (txPower-float64(rssi))/(10*n))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- Indoor positioning constants ---
	fmt.Fprintln(output, "=== BLE scan mode constants (for indoor positioning) ===")
	fmt.Fprintf(output, "  ScanModeLowPower   = %d\n", le.ScanModeLowPower)
	fmt.Fprintf(output, "  ScanModeBalanced   = %d\n", le.ScanModeBalanced)
	fmt.Fprintf(output, "  ScanModeLowLatency = %d\n", le.ScanModeLowLatency)

	fmt.Fprintln(output, "=== BLE callback type constants ===")
	fmt.Fprintf(output, "  CallbackTypeAllMatches = %d\n", le.CallbackTypeAllMatches)
	fmt.Fprintf(output, "  CallbackTypeFirstMatch = %d\n", le.CallbackTypeFirstMatch)
	fmt.Fprintf(output, "  CallbackTypeMatchLost  = %d\n", le.CallbackTypeMatchLost)

	fmt.Fprintln(output, "=== BLE match constants ===")
	fmt.Fprintf(output, "  MatchNumOneAdvertisement = %d\n", le.MatchNumOneAdvertisement)
	fmt.Fprintf(output, "  MatchNumFewAdvertisement = %d\n", le.MatchNumFewAdvertisement)
	fmt.Fprintf(output, "  MatchNumMaxAdvertisement = %d\n", le.MatchNumMaxAdvertisement)

	// --- Distance estimation demo ---
	fmt.Fprintln(output, "\n=== Distance estimation examples ===")
	const (
		txPower   = -59.0
		pathLossN = 2.5
	)
	fmt.Fprintf(output, "TX power=%.0f dBm, path-loss exponent=%.1f\n", txPower, pathLossN)
	for _, rssi := range []int32{-50, -60, -70, -80, -90} {
		dist := estimateDistance(rssi, txPower, pathLossN)
		fmt.Fprintf(output, "  RSSI=%d dBm -> estimated distance=%.2f m\n", rssi, dist)
	}

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

	// --- LE Scanner availability ---
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

	// --- PHY capabilities for indoor positioning accuracy ---
	fmt.Fprintln(output, "\n=== PHY capabilities ===")
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

	lePeriodicAdv, err := adapter.IsLePeriodicAdvertisingSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLePeriodicAdvertisingSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "LE periodic advertising supported: %v\n", lePeriodicAdv)
	}

	offloadBatch, err := adapter.IsOffloadedScanBatchingSupported()
	if err != nil {
		fmt.Fprintf(output, "IsOffloadedScanBatchingSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Offloaded scan batching supported: %v\n", offloadBatch)
	}

	fmt.Fprintln(output, "\nIndoor positioning capability check completed.")
	fmt.Fprintln(output, "No errors occurred during indoor positioning demo.")

	return nil
}
