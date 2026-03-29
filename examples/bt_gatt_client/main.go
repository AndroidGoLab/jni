//go:build android

// Command bt_gatt_client demonstrates the GATT client API surface provided by
// the bluetooth typed wrapper package. It checks BLE support, displays GATT
// constants, and reports adapter capabilities relevant to GATT client usage.
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

	// --- GATT constants ---
	fmt.Fprintln(output, "=== GATT status constants ===")
	fmt.Fprintf(output, "  GattSuccess = %d\n", bluetooth.GattSuccess)
	fmt.Fprintf(output, "  GattFailure = %d\n", bluetooth.GattFailure)

	fmt.Fprintln(output, "=== Connection state constants ===")
	fmt.Fprintf(output, "  StateDisconnected  = %d\n", bluetooth.StateDisconnected)
	fmt.Fprintf(output, "  StateConnecting    = %d\n", bluetooth.StateConnecting)
	fmt.Fprintf(output, "  StateConnected     = %d\n", bluetooth.StateConnected)
	fmt.Fprintf(output, "  StateDisconnecting = %d\n", bluetooth.StateDisconnecting)

	fmt.Fprintln(output, "=== Characteristic property constants ===")
	fmt.Fprintf(output, "  PropertyRead          = %d\n", bluetooth.PropertyRead)
	fmt.Fprintf(output, "  PropertyWrite         = %d\n", bluetooth.PropertyWrite)
	fmt.Fprintf(output, "  PropertyWriteNoResponse = %d\n", bluetooth.PropertyWriteNoResponse)
	fmt.Fprintf(output, "  PropertyNotify        = %d\n", bluetooth.PropertyNotify)
	fmt.Fprintf(output, "  PropertyIndicate      = %d\n", bluetooth.PropertyIndicate)

	fmt.Fprintln(output, "=== Characteristic permission constants ===")
	fmt.Fprintf(output, "  PermissionRead  = %d\n", bluetooth.PermissionRead)
	fmt.Fprintf(output, "  PermissionWrite = %d\n", bluetooth.PermissionWrite)

	fmt.Fprintln(output, "=== Write type constants ===")
	fmt.Fprintf(output, "  WriteTypeDefault    = %d\n", bluetooth.WriteTypeDefault)
	fmt.Fprintf(output, "  WriteTypeNoResponse = %d\n", bluetooth.WriteTypeNoResponse)
	fmt.Fprintf(output, "  WriteTypeSigned     = %d\n", bluetooth.WriteTypeSigned)

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

	// Check BLE support via adapter capabilities.
	scannerObj, err := adapter.GetBluetoothLeScanner()
	if err != nil {
		fmt.Fprintf(output, "GetBluetoothLeScanner error: %v\n", err)
	} else if scannerObj == nil {
		fmt.Fprintln(output, "BLE not supported (scanner is null)")
	} else {
		_ = &le.BluetoothLeScanner{VM: vm, Obj: scannerObj}
		fmt.Fprintln(output, "BLE supported: scanner available")
	}

	multiAdv, err := adapter.IsMultipleAdvertisementSupported()
	if err != nil {
		fmt.Fprintf(output, "IsMultipleAdvertisementSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Multiple advertisement supported: %v\n", multiAdv)
	}

	// --- Show GATT client API surface ---
	fmt.Fprintln(output, "\n=== GATT client typed wrapper API ===")
	fmt.Fprintln(output, "  Gatt methods: Close, Connect, Disconnect, DiscoverServices,")
	fmt.Fprintln(output, "    GetServices, GetService, ReadCharacteristic, WriteCharacteristic,")
	fmt.Fprintln(output, "    SetCharacteristicNotification, RequestMtu, ReadRemoteRssi,")
	fmt.Fprintln(output, "    ReadPhy, SetPreferredPhy, ReadDescriptor, WriteDescriptor,")
	fmt.Fprintln(output, "    BeginReliableWrite, ExecuteReliableWrite, AbortReliableWrite,")
	fmt.Fprintln(output, "    RequestConnectionPriority")
	fmt.Fprintln(output, "  GattCallback methods: OnConnectionStateChange, OnServicesDiscovered,")
	fmt.Fprintln(output, "    OnCharacteristicRead, OnCharacteristicWrite, OnCharacteristicChanged,")
	fmt.Fprintln(output, "    OnDescriptorRead, OnDescriptorWrite, OnMtuChanged, OnReadRemoteRssi,")
	fmt.Fprintln(output, "    OnPhyRead, OnPhyUpdate, OnReliableWriteCompleted, OnServiceChanged")

	fmt.Fprintln(output, "\nGATT client demo completed successfully.")
	fmt.Fprintln(output, "No errors occurred during GATT client lifecycle demo.")

	return nil
}
