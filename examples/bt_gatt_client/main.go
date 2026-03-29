//go:build android

// Command bt_gatt_client demonstrates the GATT client API using the bluetooth
// and bluetooth/le typed wrapper packages. It obtains a BluetoothManager, gets
// the adapter via the manager, queries connected GATT devices, checks adapter
// BLE capabilities, and retrieves the LE scanner.
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

	fmt.Fprintln(output, "=== GATT Client Demo ===")
	ui.RenderOutput()

	// --- BluetoothManager ---
	mgr, err := bluetooth.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("bluetooth.NewManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "BluetoothManager: obtained OK")
	ui.RenderOutput()

	mgrStr, err := mgr.ToString()
	if err != nil {
		fmt.Fprintf(output, "Manager.ToString: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Manager.ToString: %s\n", mgrStr)
	}

	// --- Get adapter via Manager ---
	adapterObj, err := mgr.GetAdapter()
	if err != nil {
		return fmt.Errorf("Manager.GetAdapter: %w", err)
	}
	if adapterObj == nil || adapterObj.Ref() == 0 {
		fmt.Fprintln(output, "BluetoothAdapter is null (Bluetooth may be disabled)")
		return nil
	}
	adapter := &bluetooth.Adapter{VM: vm, Obj: adapterObj}
	defer adapter.Close()
	fmt.Fprintln(output, "Adapter obtained via Manager: OK")
	ui.RenderOutput()

	// --- Adapter state ---
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
	if err != nil {
		fmt.Fprintf(output, "GetName: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Adapter name: %s\n", name)
	}

	addr, err := adapter.GetAddress()
	if err != nil {
		fmt.Fprintf(output, "GetAddress: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Adapter address: %s\n", addr)
	}

	state, err := adapter.GetState()
	if err != nil {
		fmt.Fprintf(output, "GetState: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Adapter state: %d\n", state)
	}

	scanMode, err := adapter.GetScanMode()
	if err != nil {
		fmt.Fprintf(output, "GetScanMode: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Scan mode: %d\n", scanMode)
	}
	ui.RenderOutput()

	// --- Query GATT connected devices via Manager ---
	fmt.Fprintln(output, "\n=== GATT connected devices (profile=7/GATT) ===")
	connDevs, err := mgr.GetConnectedDevices(int32(bluetooth.GattConst))
	if err != nil {
		fmt.Fprintf(output, "GetConnectedDevices(GATT): %v\n", err)
	} else if connDevs == nil || connDevs.Ref() == 0 {
		fmt.Fprintln(output, "Connected devices list: null")
	} else {
		fmt.Fprintln(output, "Connected devices list: obtained OK")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(connDevs)
			return nil
		})
	}
	ui.RenderOutput()

	// --- BLE capabilities ---
	fmt.Fprintln(output, "\n=== BLE capabilities ===")
	scannerObj, err := adapter.GetBluetoothLeScanner()
	if err != nil {
		fmt.Fprintf(output, "GetBluetoothLeScanner: error (%v)\n", err)
	} else if scannerObj == nil || scannerObj.Ref() == 0 {
		fmt.Fprintln(output, "BLE scanner: not available (null)")
	} else {
		scanner := &le.BluetoothLeScanner{VM: vm, Obj: scannerObj}
		_ = scanner
		fmt.Fprintln(output, "BLE scanner: available")
	}

	multiAdv, err := adapter.IsMultipleAdvertisementSupported()
	if err != nil {
		fmt.Fprintf(output, "IsMultipleAdvertisementSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Multiple advertisement supported: %v\n", multiAdv)
	}

	le2m, err := adapter.IsLe2MPhySupported()
	if err != nil {
		fmt.Fprintf(output, "IsLe2MPhySupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "LE 2M PHY supported: %v\n", le2m)
	}

	leCoded, err := adapter.IsLeCodedPhySupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeCodedPhySupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "LE Coded PHY supported: %v\n", leCoded)
	}

	leExtAdv, err := adapter.IsLeExtendedAdvertisingSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeExtendedAdvertisingSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "LE extended advertising supported: %v\n", leExtAdv)
	}

	lePeriodicAdv, err := adapter.IsLePeriodicAdvertisingSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLePeriodicAdvertisingSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "LE periodic advertising supported: %v\n", lePeriodicAdv)
	}

	offloadFilter, err := adapter.IsOffloadedFilteringSupported()
	if err != nil {
		fmt.Fprintf(output, "IsOffloadedFilteringSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Offloaded filtering supported: %v\n", offloadFilter)
	}

	offloadBatch, err := adapter.IsOffloadedScanBatchingSupported()
	if err != nil {
		fmt.Fprintf(output, "IsOffloadedScanBatchingSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Offloaded scan batching supported: %v\n", offloadBatch)
	}

	maxAdvLen, err := adapter.GetLeMaximumAdvertisingDataLength()
	if err != nil {
		fmt.Fprintf(output, "GetLeMaximumAdvertisingDataLength: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Max advertising data length: %d bytes\n", maxAdvLen)
	}

	maxAudioDev, err := adapter.GetMaxConnectedAudioDevices()
	if err != nil {
		fmt.Fprintf(output, "GetMaxConnectedAudioDevices: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Max connected audio devices: %d\n", maxAudioDev)
	}

	leAudio, err := adapter.IsLeAudioSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeAudioSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "LE audio supported: %d\n", leAudio)
	}

	leAudioBroadcast, err := adapter.IsLeAudioBroadcastSourceSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeAudioBroadcastSourceSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "LE audio broadcast source supported: %d\n", leAudioBroadcast)
	}

	leAudioAssist, err := adapter.IsLeAudioBroadcastAssistantSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeAudioBroadcastAssistantSupported: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "LE audio broadcast assistant supported: %d\n", leAudioAssist)
	}

	// --- Bonded devices ---
	fmt.Fprintln(output, "\n=== Bonded devices ===")
	bondedObj, err := adapter.GetBondedDevices()
	if err != nil {
		fmt.Fprintf(output, "GetBondedDevices: error (%v)\n", err)
	} else if bondedObj == nil || bondedObj.Ref() == 0 {
		fmt.Fprintln(output, "Bonded devices set: null")
	} else {
		fmt.Fprintln(output, "Bonded devices set: obtained OK")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(bondedObj)
			return nil
		})
	}

	// --- Discovery state ---
	discovering, err := adapter.IsDiscovering()
	if err != nil {
		fmt.Fprintf(output, "IsDiscovering: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Is discovering: %v\n", discovering)
	}

	// --- GATT profile connection state ---
	gattState, err := adapter.GetProfileConnectionState(int32(bluetooth.GattConst))
	if err != nil {
		fmt.Fprintf(output, "GetProfileConnectionState(GATT): error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "GATT profile connection state: %d\n", gattState)
	}

	fmt.Fprintln(output, "\nGATT client demo completed successfully.")
	return nil
}
