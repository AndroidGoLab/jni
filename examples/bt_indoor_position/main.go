//go:build android

// Command bt_indoor_position demonstrates BLE scanning capabilities relevant
// to indoor positioning using the bluetooth and bluetooth/le typed wrapper
// packages. It checks adapter BLE support, reports PHY capabilities, obtains
// both the LE scanner and advertiser, and queries all adapter properties
// relevant to indoor positioning accuracy.
//
// Required permissions (Android 12+): BLUETOOTH_SCAN, BLUETOOTH_CONNECT.
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

	fmt.Fprintln(output, "=== Indoor Positioning Demo ===")
	ui.RenderOutput()

	// --- Distance estimation demo ---
	fmt.Fprintln(output, "\n=== Distance estimation examples ===")
	const (
		txPower   = -59.0
		pathLossN = 2.5
	)
	fmt.Fprintf(output, "TX power=%.0f dBm, path-loss exponent=%.1f\n", txPower, pathLossN)
	for _, rssi := range []int32{-50, -60, -70, -80, -90} {
		dist := estimateDistance(rssi, txPower, pathLossN)
		fmt.Fprintf(output, "  RSSI=%d dBm -> distance=%.2f m\n", rssi, dist)
	}
	ui.RenderOutput()

	// --- BluetoothManager ---
	mgr, err := bluetooth.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("bluetooth.NewManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "\nBluetoothManager: obtained OK")

	mgrStr, err := mgr.ToString()
	if err == nil {
		fmt.Fprintf(output, "Manager.ToString: %s\n", mgrStr)
	}

	// --- Adapter via Manager ---
	adapterObj, err := mgr.GetAdapter()
	if err != nil {
		return fmt.Errorf("Manager.GetAdapter: %w", err)
	}
	if adapterObj == nil || adapterObj.Ref() == 0 {
		fmt.Fprintln(output, "BluetoothAdapter is null")
		return nil
	}
	adapter := &bluetooth.Adapter{VM: vm, Obj: adapterObj}
	defer adapter.Close()

	enabled, err := adapter.IsEnabled()
	if err != nil {
		return fmt.Errorf("IsEnabled: %w", err)
	}
	fmt.Fprintf(output, "Bluetooth enabled: %v\n", enabled)
	if !enabled {
		fmt.Fprintln(output, "Bluetooth is off; enable it in Settings.")
		return nil
	}

	name, err := adapter.GetName()
	if err == nil {
		fmt.Fprintf(output, "Adapter name: %s\n", name)
	}

	addr, err := adapter.GetAddress()
	if err == nil {
		fmt.Fprintf(output, "Adapter address: %s\n", addr)
	}

	state, err := adapter.GetState()
	if err == nil {
		fmt.Fprintf(output, "Adapter state: %d\n", state)
	}

	scanMode, err := adapter.GetScanMode()
	if err == nil {
		fmt.Fprintf(output, "Scan mode: %d\n", scanMode)
	}
	ui.RenderOutput()

	// --- LE Scanner ---
	fmt.Fprintln(output, "\n=== LE Scanner ===")
	scannerObj, err := adapter.GetBluetoothLeScanner()
	if err != nil {
		return fmt.Errorf("GetBluetoothLeScanner: %w", err)
	}
	if scannerObj == nil || scannerObj.Ref() == 0 {
		fmt.Fprintln(output, "BLE scanner not available")
		return nil
	}
	scanner := &le.BluetoothLeScanner{VM: vm, Obj: scannerObj}
	scannerStr, err := scanner.ToString()
	if err == nil {
		fmt.Fprintf(output, "BLE scanner: %s\n", scannerStr)
	} else {
		fmt.Fprintln(output, "BLE scanner: obtained OK")
	}
	ui.RenderOutput()

	// --- PHY capabilities for indoor positioning accuracy ---
	fmt.Fprintln(output, "\n=== PHY capabilities ===")

	le2m, err := adapter.IsLe2MPhySupported()
	if err == nil {
		fmt.Fprintf(output, "LE 2M PHY supported: %v\n", le2m)
	}

	leCoded, err := adapter.IsLeCodedPhySupported()
	if err == nil {
		fmt.Fprintf(output, "LE Coded PHY supported: %v\n", leCoded)
	}

	lePeriodicAdv, err := adapter.IsLePeriodicAdvertisingSupported()
	if err == nil {
		fmt.Fprintf(output, "LE periodic advertising supported: %v\n", lePeriodicAdv)
	}

	leExtAdv, err := adapter.IsLeExtendedAdvertisingSupported()
	if err == nil {
		fmt.Fprintf(output, "LE extended advertising supported: %v\n", leExtAdv)
	}

	offloadBatch, err := adapter.IsOffloadedScanBatchingSupported()
	if err == nil {
		fmt.Fprintf(output, "Offloaded scan batching supported: %v\n", offloadBatch)
	}

	offloadFilter, err := adapter.IsOffloadedFilteringSupported()
	if err == nil {
		fmt.Fprintf(output, "Offloaded filtering supported: %v\n", offloadFilter)
	}

	multiAdv, err := adapter.IsMultipleAdvertisementSupported()
	if err == nil {
		fmt.Fprintf(output, "Multiple advertisement supported: %v\n", multiAdv)
	}

	maxAdvLen, err := adapter.GetLeMaximumAdvertisingDataLength()
	if err == nil {
		fmt.Fprintf(output, "Max advertising data length: %d bytes\n", maxAdvLen)
	}

	maxAudioDev, err := adapter.GetMaxConnectedAudioDevices()
	if err == nil {
		fmt.Fprintf(output, "Max connected audio devices: %d\n", maxAudioDev)
	}

	leAudio, err := adapter.IsLeAudioSupported()
	if err == nil {
		fmt.Fprintf(output, "LE audio supported: %d\n", leAudio)
	}
	ui.RenderOutput()

	// --- LE Advertiser ---
	fmt.Fprintln(output, "\n=== LE Advertiser ===")
	advObj, err := adapter.GetBluetoothLeAdvertiser()
	if err != nil {
		fmt.Fprintf(output, "GetBluetoothLeAdvertiser: error (%v)\n", err)
	} else if advObj == nil || advObj.Ref() == 0 {
		fmt.Fprintln(output, "BLE advertiser: not available")
	} else {
		advertiser := &le.BluetoothLeAdvertiser{VM: vm, Obj: advObj}
		advStr, err := advertiser.ToString()
		if err == nil {
			fmt.Fprintf(output, "BLE advertiser: %s\n", advStr)
		} else {
			fmt.Fprintln(output, "BLE advertiser: obtained OK")
		}
	}
	ui.RenderOutput()

	// --- Bonded devices ---
	bondedObj, err := adapter.GetBondedDevices()
	if err != nil {
		fmt.Fprintf(output, "\nGetBondedDevices: error (%v)\n", err)
	} else if bondedObj == nil || bondedObj.Ref() == 0 {
		fmt.Fprintln(output, "\nBonded devices: null")
	} else {
		fmt.Fprintln(output, "\nBonded devices set: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(bondedObj); return nil })
	}

	// --- GATT connected devices ---
	connDevs, err := mgr.GetConnectedDevices(int32(bluetooth.GattConst))
	if err != nil {
		fmt.Fprintf(output, "GetConnectedDevices(GATT): %v\n", err)
	} else if connDevs == nil || connDevs.Ref() == 0 {
		fmt.Fprintln(output, "GATT connected devices: null")
	} else {
		fmt.Fprintln(output, "GATT connected devices: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(connDevs); return nil })
	}

	discovering, err := adapter.IsDiscovering()
	if err == nil {
		fmt.Fprintf(output, "Is discovering: %v\n", discovering)
	}

	// --- GATT profile connection state ---
	gattState, err := adapter.GetProfileConnectionState(int32(bluetooth.GattConst))
	if err == nil {
		fmt.Fprintf(output, "GATT profile connection state: %d\n", gattState)
	}

	fmt.Fprintln(output, "\nIndoor positioning demo completed successfully.")
	return nil
}
