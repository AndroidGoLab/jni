# Bluetooth API

The `bluetooth` package wraps `android.bluetooth.BluetoothAdapter`, GATT client/server, and device data classes. BLE scanning and advertising live in `bluetooth/le`.

## Adapter: Query State and Bonded Devices

```go
import "github.com/AndroidGoLab/jni/bluetooth"

adapter, err := bluetooth.NewAdapter(ctx)
if err != nil {
    return fmt.Errorf("bluetooth.NewAdapter: %w", err)
}
defer adapter.Close()

// Check if Bluetooth is enabled
enabled, err := adapter.IsEnabled()

// Get adapter name and MAC address
name, err := adapter.GetName()
addr, err := adapter.GetAddress()

// Get bonded (paired) devices - returns raw Java Set object
bonded, err := adapter.GetBondedDevices()
```

## Device Data Class

Each `BluetoothDevice` is wrapped as a data class with typed accessor methods:

```go
// Device wraps android.bluetooth.BluetoothDevice
type Device struct {
    VM  *jni.VM
    Obj *jni.GlobalRef
}

// Access device fields via getter methods
name, err := device.GetName()       // String
addr, err := device.GetAddress()    // String
devType, err := device.GetType()    // int32 (DeviceTypeClassic, DeviceTypeLe, DeviceTypeDual)
bondState, err := device.GetBondState() // int32 (BondNone, BondBonding, BondBonded)
uuids, err := device.GetUuids()     // raw Java ParcelUuid[] object
```

## Classic Discovery

Start and stop Bluetooth device discovery:

```go
// Start classic inquiry-based discovery
started, err := adapter.StartDiscovery()

// Cancel ongoing discovery
canceled, err := adapter.CancelDiscovery()
```

Discovery results are delivered via Android's `ACTION_FOUND` broadcast.

## BLE Scanning

The LE scanner is obtained from the adapter. It returns a raw JNI object that wraps into the `bluetooth/le` package types:

```go
import (
    "github.com/AndroidGoLab/jni/bluetooth"
    "github.com/AndroidGoLab/jni/bluetooth/le"
)

// Get the BLE scanner from the adapter
scannerObj, err := adapter.GetBluetoothLeScanner()
scanner := le.BluetoothLeScanner{VM: vm, Obj: scannerObj}

// Start scanning with a callback (raw JNI object)
scanner.StartScan1(callbackProxy)

// Or with filters and settings (raw JNI objects)
scanner.StartScan3_1(filtersObj, settingsObj, callbackProxy)

// Stop scanning
scanner.StopScan1(callbackProxy)
```

## BLE Advertising

```go
import "github.com/AndroidGoLab/jni/bluetooth/le"

// Get the BLE advertiser from the adapter
advertiserObj, err := adapter.GetBluetoothLeAdvertiser()
advertiser := le.BluetoothLeAdvertiser{VM: vm, Obj: advertiserObj}

// Start advertising (settings, data, callback are raw JNI objects)
advertiser.StartAdvertising3(settingsObj, dataObj, callbackProxy)

// Stop advertising
advertiser.StopAdvertising(callbackProxy)
```

## GATT Client

Connect to a remote device and interact with its GATT services:

```go
// Gatt wraps android.bluetooth.BluetoothGatt
// Obtained by calling device.ConnectGatt(...)

// Discover services (triggers onServicesDiscovered callback)
gatt.DiscoverServices()

// Read available services
services, _ := gatt.GetServices()

// Read/write characteristics
gatt.ReadCharacteristic(characteristicObj)
gatt.WriteCharacteristic1(characteristicObj)

// Enable notifications for a characteristic
gatt.SetCharacteristicNotification(characteristic, true)

// Request MTU change
gatt.RequestMtu(512)

// Read remote RSSI
gatt.ReadRemoteRssi()

// Clean up
gatt.Close()
```

## GATT Server

Host GATT services on this device:

```go
// GattServer wraps android.bluetooth.BluetoothGattServer
server.AddService(service)
server.NotifyCharacteristicChanged3(device, characteristic, confirm)
server.Close()
```

## Constants

The bluetooth package exports all Android Bluetooth constants as typed Go values:

```go
// Device types
bluetooth.DeviceTypeClassic  // 1
bluetooth.DeviceTypeLe       // 2
bluetooth.DeviceTypeDual     // 3

// Bond states
bluetooth.BondNone    // 10
bluetooth.BondBonding // 11
bluetooth.BondBonded  // 12

// Scan modes (BluetoothAdapter)
bluetooth.ScanModeNone                    // 20
bluetooth.ScanModeConnectable             // 21
bluetooth.ScanModeConnectableDiscoverable // 23

// GATT characteristic properties
bluetooth.PropertyRead     // 2
bluetooth.PropertyWrite    // 8
bluetooth.PropertyNotify   // 16
bluetooth.PropertyIndicate // 32

// GATT status / connection state
bluetooth.GattSuccess       // 0
bluetooth.StateDisconnected // 0
bluetooth.StateConnected    // 2
```

## Required Permissions (Android 12+)

```xml
<uses-permission android:name="android.permission.BLUETOOTH_SCAN" />
<uses-permission android:name="android.permission.BLUETOOTH_CONNECT" />
<uses-permission android:name="android.permission.BLUETOOTH_ADVERTISE" />
```
