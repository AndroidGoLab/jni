# Bluetooth API

The `bluetooth` package wraps `android.bluetooth.*` -- adapter, device, GATT client/server, and manager. BLE scanning and advertising live in `bluetooth/le`.

Import path: `github.com/AndroidGoLab/jni/bluetooth`

## 1. Adapter: Obtain and Query State

`bluetooth.NewAdapter` obtains the `BluetoothAdapter` via the `BluetoothManager` system service. Internally it calls `ctx.GetSystemService("bluetooth")` to get a `BluetoothManager`, then calls `BluetoothManager.getAdapter()` (JNI signature `"()Landroid/bluetooth/BluetoothAdapter;"`) on it.

```go
import (
    "fmt"

    "github.com/AndroidGoLab/jni/app"
    "github.com/AndroidGoLab/jni/bluetooth"
)

func queryAdapter(ctx *app.Context) error {
    adapter, err := bluetooth.NewAdapter(ctx)
    if err != nil {
        return fmt.Errorf("bluetooth.NewAdapter: %w", err)
    }
    defer adapter.Close() // releases the JNI global reference

    // Check if Bluetooth hardware is enabled.
    // Internally: env.CallBooleanMethod(obj, mid) -> uint8, then resultRaw != 0
    enabled, err := adapter.IsEnabled()
    if err != nil {
        return err
    }
    fmt.Printf("Bluetooth enabled: %v\n", enabled)

    // Adapter name and MAC address (both return string via CallObjectMethod + GoString)
    name, err := adapter.GetName()
    if err != nil {
        return err
    }
    addr, err := adapter.GetAddress()
    if err != nil {
        return err
    }
    fmt.Printf("Adapter: %s (%s)\n", name, addr)

    // Scan mode (int32)
    scanMode, err := adapter.GetScanMode()
    if err != nil {
        return err
    }
    fmt.Printf("Scan mode: %d\n", scanMode)

    return nil
}
```

### How NewAdapter Works (JNI Details)

For reference, this is the internal JNI sequence. You do not need to write this yourself -- `bluetooth.NewAdapter` does it -- but understanding it helps with debugging.

```go
// Pseudocode of what NewAdapter does internally:
svc, _ := ctx.GetSystemService("bluetooth")       // returns BluetoothManager global ref
bmClass, _ := env.FindClass("android/bluetooth/BluetoothManager")
getAdapterMid, _ := env.GetMethodID(bmClass, "getAdapter",
    "()Landroid/bluetooth/BluetoothAdapter;")       // JNI signature: no args, returns object
adapterLocal, _ := env.CallObjectMethod(svc, getAdapterMid)
adapter.Obj = env.NewGlobalRef(adapterLocal)        // promote to global ref
env.DeleteLocalRef(adapterLocal)                    // free local ref
```

## 2. Classic Discovery (Complete End-to-End)

Classic Bluetooth discovery uses `adapter.StartDiscovery()` to begin scanning. Results arrive via the `ACTION_FOUND` broadcast. Since `BroadcastReceiver` is an **abstract class** (not an interface), `env.NewProxy` cannot create one directly. You need a small Java adapter class.

### Step 1: Java Adapter Class

`GoAbstractDispatch.java` is already included in the library's classpath. You only need to write the `BroadcastReceiver` subclass and bundle it in your APK:

```java
// GoBroadcastReceiver.java
package center.dx.jni.generated;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import center.dx.jni.internal.GoAbstractDispatch;

public class GoBroadcastReceiver extends BroadcastReceiver {
    private final long handlerID;

    public GoBroadcastReceiver(long handlerID) {
        this.handlerID = handlerID;
    }

    @Override
    public void onReceive(Context context, Intent intent) {
        GoAbstractDispatch.invoke(handlerID, "onReceive",
            new Object[]{context, intent});
    }
}
```

### Step 2: Go Code -- Full Discovery with BroadcastReceiver

```go
import (
    "fmt"
    "unsafe"

    "github.com/AndroidGoLab/jni"
    "github.com/AndroidGoLab/jni/app"
    "github.com/AndroidGoLab/jni/bluetooth"
)

func runDiscovery(vm *jni.VM, ctx *app.Context) error {
    adapter, err := bluetooth.NewAdapter(ctx)
    if err != nil {
        return fmt.Errorf("NewAdapter: %w", err)
    }
    defer adapter.Close()

    enabled, err := adapter.IsEnabled()
    if err != nil {
        return err
    }
    if !enabled {
        return fmt.Errorf("bluetooth is not enabled")
    }

    // --- Register a BroadcastReceiver for ACTION_FOUND ---

    // 1. Register a Go handler that receives dispatched method calls.
    //    jni.RegisterProxyHandler returns a unique int64 ID.
    handlerID := jni.RegisterProxyHandler(
        func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
            if methodName != "onReceive" || len(args) < 2 {
                return nil, nil
            }
            // args[0] = Context, args[1] = Intent
            intentObj := args[1]

            // Extract the BluetoothDevice from the intent via raw JNI:
            //   BluetoothDevice device = intent.getParcelableExtra(BluetoothDevice.EXTRA_DEVICE);
            intentClass, err := env.FindClass("android/content/Intent")
            if err != nil {
                return nil, err
            }
            defer env.DeleteLocalRef(&intentClass.Object)

            getParcelableMid, err := env.GetMethodID(intentClass,
                "getParcelableExtra",
                "(Ljava/lang/String;)Landroid/os/Parcelable;")
            if err != nil {
                return nil, err
            }

            extraKey, err := env.NewStringUTF("android.bluetooth.device.extra.DEVICE")
            if err != nil {
                return nil, err
            }
            defer env.DeleteLocalRef(&extraKey.Object)

            deviceObj, err := env.CallObjectMethod(intentObj, getParcelableMid,
                jni.ObjectValue(&extraKey.Object))
            if err != nil || deviceObj == nil {
                return nil, err
            }
            defer env.DeleteLocalRef(deviceObj)

            // Wrap in the generated Device type to use typed accessors.
            // We need a global ref for the Device struct.
            deviceGlobal := env.NewGlobalRef(deviceObj)
            device := bluetooth.Device{VM: vm, Obj: deviceGlobal}

            name, _ := device.GetName()
            addr, _ := device.GetAddress()
            devType, _ := device.GetType()
            fmt.Printf("Found device: %s (%s) type=%d\n", name, addr, devType)

            env.DeleteGlobalRef(deviceGlobal)
            return nil, nil
        },
    )
    defer jni.UnregisterProxyHandler(handlerID)

    // 2. Instantiate GoBroadcastReceiver and IntentFilter via raw JNI.
    var receiverGlobal *jni.GlobalRef
    err = vm.Do(func(env *jni.Env) error {
        // Create the GoBroadcastReceiver(handlerID)
        recvClass, err := env.FindClass("center/dx/jni/generated/GoBroadcastReceiver")
        if err != nil {
            return fmt.Errorf("find GoBroadcastReceiver: %w", err)
        }
        defer env.DeleteLocalRef(&recvClass.Object)

        recvInit, err := env.GetMethodID(recvClass, "<init>", "(J)V")
        if err != nil {
            return fmt.Errorf("get <init>: %w", err)
        }
        recvLocal, err := env.NewObject(recvClass, recvInit, jni.LongValue(handlerID))
        if err != nil {
            return fmt.Errorf("new GoBroadcastReceiver: %w", err)
        }
        receiverGlobal = env.NewGlobalRef(recvLocal)
        env.DeleteLocalRef(recvLocal)

        // Create IntentFilter("android.bluetooth.device.action.FOUND")
        ifClass, err := env.FindClass("android/content/IntentFilter")
        if err != nil {
            return err
        }
        defer env.DeleteLocalRef(&ifClass.Object)

        ifInit, err := env.GetMethodID(ifClass, "<init>", "(Ljava/lang/String;)V")
        if err != nil {
            return err
        }
        actionStr, err := env.NewStringUTF("android.bluetooth.device.action.FOUND")
        if err != nil {
            return err
        }
        defer env.DeleteLocalRef(&actionStr.Object)

        filterLocal, err := env.NewObject(ifClass, ifInit, jni.ObjectValue(&actionStr.Object))
        if err != nil {
            return err
        }
        filterGlobal := env.NewGlobalRef(filterLocal)
        env.DeleteLocalRef(filterLocal)
        defer func() { env.DeleteGlobalRef(filterGlobal) }()

        // Register the receiver with the Context.
        // ctx.RegisterReceiver2(receiver, filter) -> Intent
        _, err = ctx.RegisterReceiver2(
            (*jni.Object)(unsafe.Pointer(receiverGlobal)),
            (*jni.Object)(unsafe.Pointer(filterGlobal)),
        )
        return err
    })
    if err != nil {
        return fmt.Errorf("register receiver: %w", err)
    }

    // 3. Start discovery.
    //    Internally calls env.CallBooleanMethod with JNI signature "()Z".
    //    Returns (bool, error): true if discovery started successfully.
    started, err := adapter.StartDiscovery()
    if err != nil {
        return fmt.Errorf("StartDiscovery: %w", err)
    }
    if !started {
        return fmt.Errorf("StartDiscovery returned false (check permissions)")
    }
    fmt.Println("Discovery started -- devices will arrive in the handler above")

    // ... wait for results, e.g. time.Sleep or channel ...

    // 4. Cleanup: cancel discovery + unregister receiver.
    _, _ = adapter.CancelDiscovery()
    _ = ctx.UnregisterReceiver((*jni.Object)(unsafe.Pointer(receiverGlobal)))
    _ = vm.Do(func(env *jni.Env) error {
        env.DeleteGlobalRef(receiverGlobal)
        return nil
    })

    return nil
}
```

### JNI Signature for startDiscovery

The generated code calls:

```go
env.CallBooleanMethod(m.Obj, midAdapterStartDiscovery)
```

where `midAdapterStartDiscovery` was obtained by:

```go
env.GetMethodID(adapterClass, "startDiscovery", "()Z")
//                                                ^^^
//  "()Z" means: no parameters, returns boolean (Z = boolean in JNI)
```

`env.CallBooleanMethod` returns `(uint8, error)`. The wrapper converts to Go bool:

```go
resultRaw, err := env.CallBooleanMethod(m.Obj, midAdapterStartDiscovery)
result = resultRaw != 0  // uint8 -> bool
```

This pattern is the same for `CancelDiscovery`, `IsEnabled`, `IsDiscovering`, `SetName`, etc.

## 3. Device Data Class

Each `BluetoothDevice` is wrapped with typed accessor methods:

```go
// Device wraps android.bluetooth.BluetoothDevice
type Device struct {
    VM  *jni.VM
    Obj *jni.GlobalRef
}

// All accessors follow the same vm.Do + ensureInit + CallXxxMethod pattern.
name, err := device.GetName()           // string (CallObjectMethod + GoString)
addr, err := device.GetAddress()        // string
devType, err := device.GetType()        // int32 (CallIntMethod)
bondState, err := device.GetBondState() // int32 (CallIntMethod)
alias, err := device.GetAlias()         // string
uuids, err := device.GetUuids()         // *jni.Object (raw ParcelUuid[] array)

// Boolean methods:
bonded, err := device.CreateBond()                  // bool
ok, err := device.SetPairingConfirmation(true)      // bool
ok, err := device.FetchUuidsWithSdp()               // bool

// GATT connection:
gattObj, err := device.ConnectGatt3(contextObj, false, callbackObj)
gattObj, err := device.ConnectGatt4_1(contextObj, false, callbackObj, transport)
```

## 4. BLE Scanning

### Get the Scanner

```go
import (
    "github.com/AndroidGoLab/jni"
    "github.com/AndroidGoLab/jni/bluetooth"
    "github.com/AndroidGoLab/jni/bluetooth/le"
)

func bleScan(vm *jni.VM, ctx *app.Context) error {
    adapter, err := bluetooth.NewAdapter(ctx)
    if err != nil {
        return err
    }
    defer adapter.Close()

    // GetBluetoothLeScanner returns a *jni.Object (global ref).
    scannerObj, err := adapter.GetBluetoothLeScanner()
    if err != nil {
        return fmt.Errorf("GetBluetoothLeScanner: %w", err)
    }
    scanner := le.BluetoothLeScanner{
        VM:  vm,
        Obj: (*jni.GlobalRef)(unsafe.Pointer(scannerObj)),
    }
```

### Create a ScanCallback via GoAbstractDispatch

`ScanCallback` is an **abstract class** (not an interface), so `env.NewProxy` cannot be used. Use the same Java adapter pattern as `GoBroadcastReceiver` above.

**Java adapter** (`GoScanCallback.java`):

```java
package center.dx.jni.generated;

import android.bluetooth.le.ScanCallback;
import android.bluetooth.le.ScanResult;
import center.dx.jni.internal.GoAbstractDispatch;
import java.util.List;

public class GoScanCallback extends ScanCallback {
    private final long handlerID;

    public GoScanCallback(long handlerID) {
        this.handlerID = handlerID;
    }

    @Override
    public void onScanResult(int callbackType, ScanResult result) {
        GoAbstractDispatch.invoke(handlerID, "onScanResult",
            new Object[]{Integer.valueOf(callbackType), result});
    }

    @Override
    public void onScanFailed(int errorCode) {
        GoAbstractDispatch.invoke(handlerID, "onScanFailed",
            new Object[]{Integer.valueOf(errorCode)});
    }
}
```

**Go code:**

```go
    // Register Go handler for scan callbacks.
    scanHandlerID := jni.RegisterProxyHandler(
        func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
            switch methodName {
            case "onScanResult":
                // args[0] = Integer(callbackType), args[1] = ScanResult
                if len(args) >= 2 && args[1] != nil {
                    resultGlobal := env.NewGlobalRef(args[1])
                    sr := le.ScanResult{VM: vm, Obj: resultGlobal}
                    rssi, _ := sr.GetRssi()
                    deviceObj, _ := sr.GetDevice()
                    if deviceObj != nil {
                        dev := bluetooth.Device{VM: vm, Obj: env.NewGlobalRef(deviceObj)}
                        name, _ := dev.GetName()
                        addr, _ := dev.GetAddress()
                        fmt.Printf("BLE device: %s (%s) RSSI=%d\n", name, addr, rssi)
                        env.DeleteGlobalRef(dev.Obj)
                        env.DeleteLocalRef(deviceObj)
                    }
                    env.DeleteGlobalRef(resultGlobal)
                }
            case "onScanFailed":
                fmt.Println("BLE scan failed")
            }
            return nil, nil
        },
    )
    defer jni.UnregisterProxyHandler(scanHandlerID)

    // Instantiate GoScanCallback via raw JNI.
    var callbackObj *jni.Object
    err = vm.Do(func(env *jni.Env) error {
        cbClass, err := env.FindClass("center/dx/jni/generated/GoScanCallback")
        if err != nil {
            return fmt.Errorf("find GoScanCallback: %w", err)
        }
        defer env.DeleteLocalRef(&cbClass.Object)
        cbInit, err := env.GetMethodID(cbClass, "<init>", "(J)V")
        if err != nil {
            return err
        }
        cbLocal, err := env.NewObject(cbClass, cbInit, jni.LongValue(scanHandlerID))
        if err != nil {
            return err
        }
        callbackObj = env.NewGlobalRef(cbLocal)
        env.DeleteLocalRef(cbLocal)
        return nil
    })
    if err != nil {
        return fmt.Errorf("create GoScanCallback: %w", err)
    }
    defer vm.Do(func(env *jni.Env) error {
        env.DeleteGlobalRef(callbackObj)
        return nil
    })

    // Start scanning (no filters -- receives all advertisements).
    err = scanner.StartScan(callbackObj)
    if err != nil {
        return fmt.Errorf("StartScan: %w", err)
    }

    // ... wait for results ...

    // Stop scanning.
    err = scanner.StopScan1(callbackObj)
    return err
}

## 5. BLE Advertising

```go
import "github.com/AndroidGoLab/jni/bluetooth/le"

func bleAdvertise(vm *jni.VM, ctx *app.Context) error {
    adapter, err := bluetooth.NewAdapter(ctx)
    if err != nil {
        return err
    }
    defer adapter.Close()

    advertiserObj, err := adapter.GetBluetoothLeAdvertiser()
    if err != nil {
        return fmt.Errorf("GetBluetoothLeAdvertiser: %w", err)
    }
    advertiser := le.BluetoothLeAdvertiser{
        VM:  vm,
        Obj: (*jni.GlobalRef)(unsafe.Pointer(advertiserObj)),
    }

    // Build AdvertiseSettings and AdvertiseData using the builder wrappers,
    // or construct them via raw JNI.

    // StartAdvertising3(settingsObj, dataObj, callbackProxy)
    err = advertiser.StartAdvertising3(settingsObj, dataObj, callbackProxy)
    if err != nil {
        return err
    }

    // ... later ...

    // StopAdvertising(callbackProxy)
    return advertiser.StopAdvertising(callbackProxy)
}
```

Builder types available: `le.AdvertiseSettingsBuilder`, `le.AdvertiseDataBuilder`.

## 6. GATT Client

Connect to a remote device, discover services, and read/write characteristics.

```go
func gattClient(vm *jni.VM, ctx *app.Context, device *bluetooth.Device) error {
    // Create a BluetoothGattCallback proxy (abstract class -- use Java adapter).
    // Methods: onConnectionStateChange, onServicesDiscovered, onCharacteristicRead,
    //          onCharacteristicWrite, onCharacteristicChanged, onMtuChanged, ...

    // Connect to the GATT server on the remote device.
    // ConnectGatt3(context, autoConnect, callback) -> raw BluetoothGatt object
    gattObj, err := device.ConnectGatt3(
        (*jni.Object)(unsafe.Pointer(ctx.Obj)),
        false,             // autoConnect
        callbackProxyObj,  // BluetoothGattCallback proxy
    )
    if err != nil {
        return fmt.Errorf("ConnectGatt: %w", err)
    }
    gatt := bluetooth.Gatt{
        VM:  vm,
        Obj: (*jni.GlobalRef)(unsafe.Pointer(gattObj)),
    }
    defer gatt.Close()

    // Discover services (triggers onServicesDiscovered callback).
    // Returns bool -- true if discovery started.
    ok, err := gatt.DiscoverServices()
    if err != nil || !ok {
        return fmt.Errorf("DiscoverServices: ok=%v err=%w", ok, err)
    }

    // After onServicesDiscovered fires, get a specific service by UUID:
    // serviceObj is a raw BluetoothGattService
    serviceObj, err := gatt.GetService(uuidObj)

    // Read/write characteristics (objects obtained from the service)
    ok, err = gatt.ReadCharacteristic(characteristicObj)
    ok, err = gatt.WriteCharacteristic1(characteristicObj)
    // Or with explicit value and write type (API 33+):
    // gatt.WriteCharacteristic3_1(characteristicObj, valueBytes, writeType)

    // Enable notifications for a characteristic
    ok, err = gatt.SetCharacteristicNotification(characteristicObj, true)

    // Request MTU change (triggers onMtuChanged callback)
    ok, err = gatt.RequestMtu(512)

    // Read remote RSSI (triggers onReadRemoteRssi callback)
    ok, err = gatt.ReadRemoteRssi()

    return nil
}
```

## 7. GATT Server

Host GATT services on this device using `bluetooth.Manager.OpenGattServer`.

```go
func gattServer(vm *jni.VM, ctx *app.Context) error {
    mgr, err := bluetooth.NewManager(ctx)
    if err != nil {
        return err
    }
    defer mgr.Close()

    // OpenGattServer(context, callback) -> raw BluetoothGattServer object
    // callback must implement BluetoothGattServerCallback (abstract class)
    serverObj, err := mgr.OpenGattServer(
        (*jni.Object)(unsafe.Pointer(ctx.Obj)),
        serverCallbackProxy,
    )
    if err != nil {
        return fmt.Errorf("OpenGattServer: %w", err)
    }
    server := bluetooth.GattServer{
        VM:  vm,
        Obj: (*jni.GlobalRef)(unsafe.Pointer(serverObj)),
    }
    defer server.Close()

    // Add a service (constructed via raw JNI using BluetoothGattService constructor)
    ok, err := server.AddService(serviceObj)

    // Notify connected clients of characteristic changes
    err = server.NotifyCharacteristicChanged3(deviceObj, characteristicObj, true)

    // Send response to a read/write request (in the callback)
    err = server.SendResponse(deviceObj, requestID, bluetooth.GattSuccess, offset, valueObj)

    return nil
}
```

## 8. Constants

All Android Bluetooth constants are exported as typed Go values.

```go
import "github.com/AndroidGoLab/jni/bluetooth"

// Device types (from BluetoothDevice)
bluetooth.DeviceTypeClassic  // 1
bluetooth.DeviceTypeLe       // 2
bluetooth.DeviceTypeDual     // 3
bluetooth.DeviceTypeUnknown  // 0

// Bond states (from BluetoothDevice)
bluetooth.BondNone    // 10
bluetooth.BondBonding // 11
bluetooth.BondBonded  // 12

// Scan modes (from BluetoothAdapter)
bluetooth.ScanModeNone                    // 20
bluetooth.ScanModeConnectable             // 21
bluetooth.ScanModeConnectableDiscoverable // 23

// GATT characteristic properties (from BluetoothGattCharacteristic)
bluetooth.PropertyRead           // 2
bluetooth.PropertyWrite          // 8
bluetooth.PropertyWriteNoResponse // 4
bluetooth.PropertyNotify         // 16
bluetooth.PropertyIndicate       // 32

// GATT status codes (from BluetoothGatt)
bluetooth.GattSuccess                   // 0
bluetooth.GattFailure                   // 257
bluetooth.GattReadNotPermitted          // 2
bluetooth.GattWriteNotPermitted         // 3
bluetooth.GattInsufficientAuthentication // 5
bluetooth.GattInsufficientEncryption    // 15
bluetooth.GattInvalidOffset             // 7
bluetooth.GattInvalidAttributeLength    // 13
bluetooth.GattConnectionCongested       // 143
bluetooth.GattConnectionTimeout         // 147

// Connection state (from BluetoothProfile)
bluetooth.StateDisconnected  // 0
bluetooth.StateConnecting    // 1
bluetooth.StateConnected     // 2
bluetooth.StateDisconnecting // 3

// GATT connection priority
bluetooth.ConnectionPriorityBalanced // 0
bluetooth.ConnectionPriorityHigh     // 1
bluetooth.ConnectionPriorityLowPower // 2
```

BLE-specific constants live in `bluetooth/le`:

```go
import "github.com/AndroidGoLab/jni/bluetooth/le"

// Scan settings
le.ScanModeBalanced      // 1
le.ScanModeLowLatency    // 2
le.ScanModeLowPower      // 0
le.ScanModeOpportunistic // -1

// Advertise results
le.AdvertiseSuccess                // 0
le.AdvertiseFailedDataTooLarge     // 1
le.AdvertiseFailedTooManyAdvertisers // 2
le.AdvertiseFailedAlreadyStarted   // 3
le.AdvertiseFailedInternalError    // 4
le.AdvertiseFailedFeatureUnsupported // 5

// Scan failures
le.ScanFailedAlreadyStarted                 // 1
le.ScanFailedApplicationRegistrationFailed   // 2
le.ScanFailedInternalError                   // 3
le.ScanFailedFeatureUnsupported              // 4
le.ScanFailedOutOfHardwareResources          // 5
le.ScanFailedScanningTooFrequently           // 6
```

## 9. JNI Patterns Reference

Key JNI call patterns used throughout the bluetooth package.

### Boolean methods (CallBooleanMethod)

```go
// JNI signature "()Z" -- no args, returns boolean
mid, _ := env.GetMethodID(cls, "startDiscovery", "()Z")
resultRaw, err := env.CallBooleanMethod(obj, mid)  // returns (uint8, error)
result := resultRaw != 0                            // convert to Go bool
```

### String methods (CallObjectMethod + GoString)

```go
// JNI signature "()Ljava/lang/String;"
mid, _ := env.GetMethodID(cls, "getName", "()Ljava/lang/String;")
resultObj, err := env.CallObjectMethod(obj, mid)
name := env.GoString((*jni.String)(unsafe.Pointer(resultObj)))
```

### Int methods (CallIntMethod)

```go
// JNI signature "()I"
mid, _ := env.GetMethodID(cls, "getType", "()I")
result, err := env.CallIntMethod(obj, mid)  // returns (int32, error)
```

### Void methods (CallVoidMethod)

```go
// JNI signature "(Landroid/os/Parcel;I)V"
mid, _ := env.GetMethodID(cls, "writeToParcel", "(Landroid/os/Parcel;I)V")
err := env.CallVoidMethod(obj, mid, jni.ObjectValue(parcel), jni.IntValue(flags))
```

### Object methods (CallObjectMethod + GlobalRef)

```go
// JNI signature "()Landroid/bluetooth/le/BluetoothLeScanner;"
mid, _ := env.GetMethodID(cls, "getBluetoothLeScanner",
    "()Landroid/bluetooth/le/BluetoothLeScanner;")
localRef, err := env.CallObjectMethod(obj, mid)
globalRef := env.NewGlobalRef(localRef)  // must promote if keeping across vm.Do calls
env.DeleteLocalRef(localRef)
```

### Creating Java strings

```go
jStr, err := env.NewStringUTF("some string")
defer env.DeleteLocalRef(&jStr.Object)
// Pass to JNI: jni.ObjectValue(&jStr.Object)
```

## 10. Required Permissions

Add to `AndroidManifest.xml`. All are runtime permissions on Android 12+.

```xml
<!-- Classic discovery + BLE scanning -->
<uses-permission android:name="android.permission.BLUETOOTH_SCAN" />

<!-- Connecting to devices, reading names/addresses -->
<uses-permission android:name="android.permission.BLUETOOTH_CONNECT" />

<!-- BLE advertising -->
<uses-permission android:name="android.permission.BLUETOOTH_ADVERTISE" />

<!-- Required for BLE scanning on Android 11 and below -->
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
```

Request these at runtime before calling `StartDiscovery` or BLE scan/advertise methods. Without them, the calls will return `false` or fail with a security exception.
