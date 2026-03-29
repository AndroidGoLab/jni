//go:build android

// Command bt_mesh_relay demonstrates BLE advertising and scanning capabilities
// relevant to mesh relay operation using the bluetooth and bluetooth/le typed
// wrapper packages. It obtains the adapter, checks advertising capabilities,
// and retrieves both the LE scanner and advertiser.
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

	fmt.Fprintln(output, "=== Mesh Relay Demo ===")
	ui.RenderOutput()

	// --- BluetoothManager ---
	mgr, err := bluetooth.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("bluetooth.NewManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "BluetoothManager: obtained OK")

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
	fmt.Fprintln(output, "Adapter obtained via Manager: OK")
	ui.RenderOutput()

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

	state, err := adapter.GetState()
	if err == nil {
		fmt.Fprintf(output, "Adapter state: %d\n", state)
	}
	ui.RenderOutput()

	// --- Mesh relay requires both scanning and advertising ---
	fmt.Fprintln(output, "\n=== Advertising capabilities ===")

	multiAdv, err := adapter.IsMultipleAdvertisementSupported()
	if err != nil {
		fmt.Fprintf(output, "IsMultipleAdvertisementSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Multiple advertisement supported: %v\n", multiAdv)
	}

	leExtAdv, err := adapter.IsLeExtendedAdvertisingSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeExtendedAdvertisingSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "LE extended advertising: %v\n", leExtAdv)
	}

	lePeriodicAdv, err := adapter.IsLePeriodicAdvertisingSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLePeriodicAdvertisingSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "LE periodic advertising: %v\n", lePeriodicAdv)
	}

	maxAdvLen, err := adapter.GetLeMaximumAdvertisingDataLength()
	if err != nil {
		fmt.Fprintf(output, "GetLeMaximumAdvertisingDataLength: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Max advertising data length: %d bytes\n", maxAdvLen)
	}
	ui.RenderOutput()

	// --- PHY capabilities (important for mesh range/throughput) ---
	fmt.Fprintln(output, "\n=== PHY capabilities ===")

	le2m, err := adapter.IsLe2MPhySupported()
	if err == nil {
		fmt.Fprintf(output, "LE 2M PHY supported: %v\n", le2m)
	}

	leCoded, err := adapter.IsLeCodedPhySupported()
	if err == nil {
		fmt.Fprintf(output, "LE Coded PHY supported: %v\n", leCoded)
	}

	leAudio, err := adapter.IsLeAudioSupported()
	if err == nil {
		fmt.Fprintf(output, "LE audio supported: %d\n", leAudio)
	}

	offloadFilter, err := adapter.IsOffloadedFilteringSupported()
	if err == nil {
		fmt.Fprintf(output, "Offloaded filtering: %v\n", offloadFilter)
	}

	offloadBatch, err := adapter.IsOffloadedScanBatchingSupported()
	if err == nil {
		fmt.Fprintf(output, "Offloaded scan batching: %v\n", offloadBatch)
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

	// --- LE Advertiser ---
	fmt.Fprintln(output, "\n=== LE Advertiser ===")
	advertiserObj, err := adapter.GetBluetoothLeAdvertiser()
	if err != nil {
		return fmt.Errorf("GetBluetoothLeAdvertiser: %w", err)
	}
	if advertiserObj == nil || advertiserObj.Ref() == 0 {
		fmt.Fprintln(output, "BLE advertiser not available")
		return nil
	}
	advertiser := &le.BluetoothLeAdvertiser{VM: vm, Obj: advertiserObj}
	advStr, err := advertiser.ToString()
	if err == nil {
		fmt.Fprintf(output, "BLE advertiser: %s\n", advStr)
	} else {
		fmt.Fprintln(output, "BLE advertiser: obtained OK")
	}
	ui.RenderOutput()

	// --- Bonded devices (mesh nodes may be bonded) ---
	fmt.Fprintln(output, "\n=== Bonded devices ===")
	bondedObj, err := adapter.GetBondedDevices()
	if err != nil {
		fmt.Fprintf(output, "GetBondedDevices: error (%v)\n", err)
	} else if bondedObj == nil || bondedObj.Ref() == 0 {
		fmt.Fprintln(output, "Bonded devices: null")
	} else {
		fmt.Fprintln(output, "Bonded devices set: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(bondedObj); return nil })
	}

	// --- GATT connected devices via Manager ---
	connDevs, err := mgr.GetConnectedDevices(int32(bluetooth.GattConst))
	if err != nil {
		fmt.Fprintf(output, "GetConnectedDevices(GATT): %v\n", err)
	} else if connDevs == nil || connDevs.Ref() == 0 {
		fmt.Fprintln(output, "GATT connected devices: null")
	} else {
		fmt.Fprintln(output, "GATT connected devices: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(connDevs); return nil })
	}

	// --- Mesh readiness summary ---
	meshReady := multiAdv && (scannerObj != nil) && (advertiserObj != nil)
	fmt.Fprintf(output, "\nMesh relay capable: %v\n", meshReady)

	fmt.Fprintln(output, "\nMesh relay demo completed successfully.")
	return nil
}
