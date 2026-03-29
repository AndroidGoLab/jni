//go:build android

// Command bt_gatt_server demonstrates the GATT server API using the bluetooth
// typed wrapper package. It obtains a BluetoothManager, gets the adapter via
// the manager, queries connected GATT server devices, checks adapter
// capabilities relevant to GATT server operation, and reports BLE features.
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

	fmt.Fprintln(output, "=== GATT Server Demo ===")
	ui.RenderOutput()

	// --- BluetoothManager ---
	mgr, err := bluetooth.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("bluetooth.NewManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "BluetoothManager: obtained OK")

	mgrStr, err := mgr.ToString()
	if err == nil {
		fmt.Fprintf(output, "Manager.ToString: %s\n", mgrStr)
	}
	ui.RenderOutput()

	// --- Adapter via Manager ---
	adapterObj, err := mgr.GetAdapter()
	if err != nil {
		return fmt.Errorf("Manager.GetAdapter: %w", err)
	}
	if adapterObj == nil {
		fmt.Fprintln(output, "BluetoothAdapter is null")
		return nil
	}
	adapter := &bluetooth.Adapter{VM: vm, Obj: adapterObj}
	defer adapter.Close()
	fmt.Fprintln(output, "Adapter obtained via Manager: OK")

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

	// --- GATT server connected devices ---
	fmt.Fprintln(output, "\n=== GATT server connected devices (profile=8) ===")
	connDevs, err := mgr.GetConnectedDevices(int32(bluetooth.GattServerConst))
	if err != nil {
		fmt.Fprintf(output, "GetConnectedDevices(GATT_SERVER): %v\n", err)
	} else if connDevs == nil {
		fmt.Fprintln(output, "Connected devices list: null")
	} else {
		fmt.Fprintln(output, "Connected devices list: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(connDevs); return nil })
	}

	// Also check GATT client connected devices.
	connDevsGatt, err := mgr.GetConnectedDevices(int32(bluetooth.GattConst))
	if err != nil {
		fmt.Fprintf(output, "GetConnectedDevices(GATT): %v\n", err)
	} else if connDevsGatt == nil {
		fmt.Fprintln(output, "GATT client connected devices: null")
	} else {
		fmt.Fprintln(output, "GATT client connected devices: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(connDevsGatt); return nil })
	}
	ui.RenderOutput()

	// --- Advertising capabilities (needed for GATT server to be discoverable) ---
	fmt.Fprintln(output, "\n=== Advertising capabilities ===")

	multiAdv, err := adapter.IsMultipleAdvertisementSupported()
	if err == nil {
		fmt.Fprintf(output, "Multiple advertisement supported: %v\n", multiAdv)
	}

	leExtAdv, err := adapter.IsLeExtendedAdvertisingSupported()
	if err == nil {
		fmt.Fprintf(output, "LE extended advertising: %v\n", leExtAdv)
	}

	lePeriodicAdv, err := adapter.IsLePeriodicAdvertisingSupported()
	if err == nil {
		fmt.Fprintf(output, "LE periodic advertising: %v\n", lePeriodicAdv)
	}

	maxAdvLen, err := adapter.GetLeMaximumAdvertisingDataLength()
	if err == nil {
		fmt.Fprintf(output, "Max advertising data length: %d bytes\n", maxAdvLen)
	}

	maxAudioDev, err := adapter.GetMaxConnectedAudioDevices()
	if err == nil {
		fmt.Fprintf(output, "Max connected audio devices: %d\n", maxAudioDev)
	}
	ui.RenderOutput()

	// --- PHY capabilities ---
	fmt.Fprintln(output, "\n=== PHY capabilities ===")

	le2m, err := adapter.IsLe2MPhySupported()
	if err == nil {
		fmt.Fprintf(output, "LE 2M PHY: %v\n", le2m)
	}

	leCoded, err := adapter.IsLeCodedPhySupported()
	if err == nil {
		fmt.Fprintf(output, "LE Coded PHY: %v\n", leCoded)
	}

	leAudio, err := adapter.IsLeAudioSupported()
	if err == nil {
		fmt.Fprintf(output, "LE audio: %d\n", leAudio)
	}

	leAudioBroadcast, err := adapter.IsLeAudioBroadcastSourceSupported()
	if err == nil {
		fmt.Fprintf(output, "LE audio broadcast source: %d\n", leAudioBroadcast)
	}

	leAudioAssist, err := adapter.IsLeAudioBroadcastAssistantSupported()
	if err == nil {
		fmt.Fprintf(output, "LE audio broadcast assistant: %d\n", leAudioAssist)
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

	// --- LE Scanner and Advertiser ---
	fmt.Fprintln(output, "\n=== LE Scanner & Advertiser ===")
	scannerObj, err := adapter.GetBluetoothLeScanner()
	if err != nil {
		fmt.Fprintf(output, "GetBluetoothLeScanner: error (%v)\n", err)
	} else if scannerObj == nil {
		fmt.Fprintln(output, "BLE scanner: not available")
	} else {
		scanner := &le.BluetoothLeScanner{VM: vm, Obj: scannerObj}
		scannerStr, err := scanner.ToString()
		if err == nil {
			fmt.Fprintf(output, "BLE scanner: %s\n", scannerStr)
		} else {
			fmt.Fprintln(output, "BLE scanner: obtained OK")
		}
	}

	advObj, err := adapter.GetBluetoothLeAdvertiser()
	if err != nil {
		fmt.Fprintf(output, "GetBluetoothLeAdvertiser: error (%v)\n", err)
	} else if advObj == nil {
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
	if err == nil && bondedObj != nil {
		fmt.Fprintln(output, "\nBonded devices set: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(bondedObj); return nil })
	}

	discovering, err := adapter.IsDiscovering()
	if err == nil {
		fmt.Fprintf(output, "Is discovering: %v\n", discovering)
	}

	// --- Profile connection states ---
	gattState, err := adapter.GetProfileConnectionState(int32(bluetooth.GattConst))
	if err == nil {
		fmt.Fprintf(output, "GATT profile state: %d\n", gattState)
	}

	gattServerState, err := adapter.GetProfileConnectionState(int32(bluetooth.GattServerConst))
	if err == nil {
		fmt.Fprintf(output, "GATT server profile state: %d\n", gattServerState)
	}

	discoverableTimeout, err := adapter.GetDiscoverableTimeout()
	if err == nil && discoverableTimeout != nil {
		fmt.Fprintln(output, "Discoverable timeout: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(discoverableTimeout); return nil })
	}

	fmt.Fprintln(output, "\nGATT server demo completed successfully.")
	return nil
}
