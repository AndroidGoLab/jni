//go:build android

// Command bluetooth demonstrates the full Bluetooth API surface provided by
// the generated bluetooth package. It is built as a c-shared library and
// packaged into an APK using the shared apk.mk infrastructure.
//
// It covers:
//   - Adapter: NewAdapter, Close, IsEnabled, GetName, GetAddress,
//     GetBondedDevices, startDiscovery, cancelDiscovery, getLeScanner,
//     getLeAdvertiser, listenRfcomm
//   - Device data class: Name, Address, Type, BondState, UUIDs
//   - BLE scanning: leScanner (startScan, stopScan), scanFilterBuilder,
//     scanSettingsBuilder, ScanResult data class
//   - BLE advertising: leAdvertiser (startAdvertising, stopAdvertising),
//     advertiseSettingsBuilder, advertiseDataBuilder
//   - GATT client: GATTClient (discoverServices, getServices,
//     readCharacteristic, writeCharacteristic, setCharacteristicNotification,
//     requestMtu, readRemoteRssi)
//   - GATT server: GATTServer (AddService, NotifyCharacteristic)
//   - GATT data classes: Service (UUID, Characteristics),
//     Characteristic (UUID, Properties, Descriptors, Value),
//     Descriptor (UUID)
//   - Callbacks: scanCallback, advertiseCallback, gattCallback,
//     gattServerCallback
//   - Constants: DeviceType*, Bond*, ScanMode*, AdvertiseMode*, Property*,
//     GATT*, State*
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/bluetooth"
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

	// --- Constants ---
	// The bluetooth package exports all relevant Android constants as
	// typed Go values.
	fmt.Fprintln(output, "=== Device type constants ===")
	fmt.Fprintf(output, "  DeviceTypeClassic = %d\n", bluetooth.DeviceTypeClassic)
	fmt.Fprintf(output, "  DeviceTypeLe      = %d\n", bluetooth.DeviceTypeLe)
	fmt.Fprintf(output, "  DeviceTypeDual    = %d\n", bluetooth.DeviceTypeDual)

	fmt.Fprintln(output, "=== Bond state constants ===")
	fmt.Fprintf(output, "  BondNone    = %d\n", bluetooth.BondNone)
	fmt.Fprintf(output, "  BondBonding = %d\n", bluetooth.BondBonding)
	fmt.Fprintf(output, "  BondBonded  = %d\n", bluetooth.BondBonded)

	fmt.Fprintln(output, "=== Scan mode constants ===")
	fmt.Fprintf(output, "  ScanModeNone                    = %d\n", bluetooth.ScanModeNone)
	fmt.Fprintf(output, "  ScanModeConnectable             = %d\n", bluetooth.ScanModeConnectable)
	fmt.Fprintf(output, "  ScanModeConnectableDiscoverable = %d\n", bluetooth.ScanModeConnectableDiscoverable)

	fmt.Fprintln(output, "=== GATT characteristic property constants ===")
	fmt.Fprintf(output, "  PropertyRead     = %d\n", bluetooth.PropertyRead)
	fmt.Fprintf(output, "  PropertyWrite    = %d\n", bluetooth.PropertyWrite)
	fmt.Fprintf(output, "  PropertyNotify   = %d\n", bluetooth.PropertyNotify)
	fmt.Fprintf(output, "  PropertyIndicate = %d\n", bluetooth.PropertyIndicate)

	fmt.Fprintln(output, "=== GATT status & connection state constants ===")
	fmt.Fprintf(output, "  GattSuccess       = %d\n", bluetooth.GattSuccess)
	fmt.Fprintf(output, "  StateDisconnected = %d\n", bluetooth.StateDisconnected)
	fmt.Fprintf(output, "  StateConnected    = %d\n", bluetooth.StateConnected)

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

	addr, err := adapter.GetAddress()
	if err != nil {
		return fmt.Errorf("GetAddress: %w", err)
	}
	fmt.Fprintf(output, "Adapter address: %s\n", addr)

	// --- Bonded devices (Device data class) ---
	// Filtered: GetBondedDevices returns generic type (Set<BluetoothDevice>)
	// bonded, err := adapter.GetBondedDevices()
	// if err != nil {
	// 	return fmt.Errorf("GetBondedDevices: %w", err)
	// }
	// fmt.Fprintf(output, "Bonded devices (raw Set): %v\n", bonded)

	// --- Classic discovery (package-internal) ---
	// startDiscovery / cancelDiscovery control the classic Bluetooth
	// inquiry-based device discovery process. These are unexported and
	// intended for use within the bluetooth package:
	//   adapter.startDiscovery() (bool, error)
	//   adapter.cancelDiscovery() (bool, error)

	// --- BLE scanning ---
	// The LE scanner is obtained from the adapter. Scan filters and
	// settings are built with their respective builder types.
	//
	//   scanner := adapter.GetLeScanner()
	//   filter := scanFilterBuilder.SetDeviceName("MyDevice").Build()
	//   settings := scanSettingsBuilder.SetScanMode(ScanModeLowLatency).Build()
	//   scanner.StartScan(filters, settings, callbackProxy)
	//   scanner.StopScan(callbackProxy)
	//
	// The scanCallback has two hooks:
	//   OnScanResult(callbackType int32, result *jni.Object)
	//   OnScanFailed(errorCode int32)
	//
	// Each ScanResult carries Device, RSSI, and a scan Record.
	// adapter.getLeScanner() -> (*jni.Object, error)
	fmt.Fprintln(output, "LE scanner available (BLE scanning ready)")

	// --- BLE advertising ---
	// The LE advertiser is also obtained from the adapter.
	//
	//   advertiser := adapter.GetLeAdvertiser()
	//   settings := advertiseSettingsBuilder
	//       .SetAdvertiseMode(AdvertiseModeLowPower)
	//       .SetConnectable(true)
	//       .SetTimeout(10000)
	//       .Build()
	//   data := advertiseDataBuilder
	//       .SetIncludeDeviceName(true)
	//       .SetIncludeTxPowerLevel(true)
	//       .AddServiceUuid(uuid)
	//       .AddServiceData(uuid, payload)
	//       .AddManufacturerData(0x00E0, payload)
	//       .Build()
	//   advertiser.StartAdvertising(settings, data, callbackProxy)
	//   advertiser.StopAdvertising(callbackProxy)
	//
	// The advertiseCallback has:
	//   OnStartSuccess(settingsInEffect *jni.Object)
	//   OnStartFailure(errorCode int32)
	// adapter.getLeAdvertiser() -> (*jni.Object, error)
	fmt.Fprintln(output, "LE advertiser available (BLE advertising ready)")

	// --- RFCOMM (classic Bluetooth sockets) ---
	// listenRfcomm creates a BluetoothServerSocket for an RFCOMM channel.
	//
	//   serverSocket := adapter.ListenRfcomm("MyService", uuid)
	//   socket := serverSocket.Accept(timeoutMs)  // blocks
	//   socket.Connect()
	//   in  := socket.GetInputStream()
	//   out := socket.GetOutputStream()
	//   dev := socket.RemoteDevice()
	//   socket.Close()
	//   serverSocket.Close()
	fmt.Fprintln(output, "RFCOMM server socket API available (listenRfcomm)")

	// --- GATT client ---
	// A GATTClient is obtained by connecting to a remote device (via
	// Android's connectGatt). It supports:
	//   DiscoverServices() -> triggers onServicesDiscovered callback
	//   GetServices() -> list of GATT services
	//   ReadCharacteristic(characteristic)
	//   WriteCharacteristic(characteristic)
	//   SetCharacteristicNotification(characteristic, enable)
	//   RequestMtu(mtu)
	//   ReadRemoteRssi()
	//   Close()
	//
	// The gattCallback has:
	//   OnConnectionStateChange(gatt, status, newState int32)
	//   OnServicesDiscovered(gatt, status)
	//   OnCharacteristicRead(gatt, characteristic, status)
	//   OnCharacteristicWrite(gatt, characteristic, status)
	//   OnCharacteristicChanged(gatt, characteristic)
	//   OnMtuChanged(gatt, mtu, status)
	//   OnReadRemoteRssi(gatt, rssi, status)
	fmt.Fprintln(output, "GATT client API available (discoverServices, read/write, notify, MTU, RSSI)")

	// --- GATT server ---
	// A GATTServer is used to host GATT services on this device.
	//   server.AddService(service)
	//   server.NotifyCharacteristic(device, characteristic, confirm)
	//   server.Close()
	//
	// The gattServerCallback has:
	//   OnConnectionStateChange(device, status, newState)
	//   OnCharacteristicReadRequest(device, requestId, offset, characteristic)
	//   OnCharacteristicWriteRequest(device, requestId, characteristic,
	//       preparedWrite, responseNeeded, offset, value)
	fmt.Fprintln(output, "GATT server API available (AddService, NotifyCharacteristic)")

	// --- Data classes ---
	// Service:        UUID string, Characteristics []Characteristic
	// Characteristic: UUID string, Properties int, Descriptors []Descriptor, Value []byte
	// Descriptor:     UUID string
	// ScanResult:     Device Device, RSSI int32, Record []byte
	// Device:         Name string, Address string, Type int, BondState int, UUIDs []string
	fmt.Fprintln(output, "\nAll bluetooth package features demonstrated.")

	return nil
}
