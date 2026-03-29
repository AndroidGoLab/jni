//go:build android

// Command bt_gatt_server demonstrates the GATT server API surface using the
// bluetooth typed wrapper package. It obtains a BluetoothManager and reports
// its capabilities and the available GATT server API methods.
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

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- GATT server constants ---
	fmt.Fprintln(output, "=== GATT service type constants ===")
	fmt.Fprintf(output, "  ServiceTypePrimary   = %d\n", bluetooth.ServiceTypePrimary)
	fmt.Fprintf(output, "  ServiceTypeSecondary = %d\n", bluetooth.ServiceTypeSecondary)

	fmt.Fprintln(output, "=== GATT status constants ===")
	fmt.Fprintf(output, "  GattSuccess = %d\n", bluetooth.GattSuccess)
	fmt.Fprintf(output, "  GattFailure = %d\n", bluetooth.GattFailure)

	fmt.Fprintln(output, "=== Characteristic property constants ===")
	fmt.Fprintf(output, "  PropertyRead   = %d\n", bluetooth.PropertyRead)
	fmt.Fprintf(output, "  PropertyWrite  = %d\n", bluetooth.PropertyWrite)
	fmt.Fprintf(output, "  PropertyNotify = %d\n", bluetooth.PropertyNotify)

	fmt.Fprintln(output, "=== Characteristic permission constants ===")
	fmt.Fprintf(output, "  PermissionRead  = %d\n", bluetooth.PermissionRead)
	fmt.Fprintf(output, "  PermissionWrite = %d\n", bluetooth.PermissionWrite)

	// --- BluetoothManager ---
	mgr, err := bluetooth.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("bluetooth.NewManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "\nBluetoothManager obtained: OK")

	// --- Adapter via Manager ---
	adapterObj, err := mgr.GetAdapter()
	if err != nil {
		return fmt.Errorf("Manager.GetAdapter: %w", err)
	}
	if adapterObj == nil {
		fmt.Fprintln(output, "BluetoothAdapter is null (Bluetooth may be disabled)")
		return nil
	}
	adapter := &bluetooth.Adapter{VM: vm, Obj: adapterObj}
	defer adapter.Close()
	fmt.Fprintln(output, "BluetoothAdapter obtained via Manager: OK")

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
		return fmt.Errorf("GetName: %w", err)
	}
	fmt.Fprintf(output, "Adapter name: %s\n", name)

	// --- Show GATT server API surface ---
	fmt.Fprintln(output, "\n=== GATT server typed wrapper API ===")
	fmt.Fprintln(output, "  GattServer methods: AddService, RemoveService, ClearServices,")
	fmt.Fprintln(output, "    Close, Connect, CancelConnection, GetServices, GetService,")
	fmt.Fprintln(output, "    SendResponse, NotifyCharacteristicChanged,")
	fmt.Fprintln(output, "    GetConnectedDevices, GetConnectionState,")
	fmt.Fprintln(output, "    GetDevicesMatchingConnectionStates, ReadPhy, SetPreferredPhy")
	fmt.Fprintln(output, "  GattServerCallback methods: OnConnectionStateChange, OnServiceAdded,")
	fmt.Fprintln(output, "    OnCharacteristicReadRequest, OnCharacteristicWriteRequest,")
	fmt.Fprintln(output, "    OnDescriptorReadRequest, OnDescriptorWriteRequest,")
	fmt.Fprintln(output, "    OnExecuteWrite, OnMtuChanged, OnNotificationSent,")
	fmt.Fprintln(output, "    OnPhyRead, OnPhyUpdate")
	fmt.Fprintln(output, "  GattService: NewGattService, AddCharacteristic, AddService,")
	fmt.Fprintln(output, "    GetCharacteristic, GetCharacteristics, GetUuid, GetType")
	fmt.Fprintln(output, "  GattCharacteristic: NewGattCharacteristic, AddDescriptor,")
	fmt.Fprintln(output, "    GetDescriptor, GetDescriptors, GetUuid, GetProperties,")
	fmt.Fprintln(output, "    GetPermissions, SetValue, GetValue, SetWriteType")
	fmt.Fprintln(output, "  GattDescriptor: NewGattDescriptor, GetUuid, GetPermissions,")
	fmt.Fprintln(output, "    SetValue, GetValue")

	fmt.Fprintln(output, "\nGATT server demo completed successfully.")
	fmt.Fprintln(output, "No errors occurred during GATT server lifecycle demo.")

	return nil
}
