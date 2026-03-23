# General Patterns: Using Any Android API from Go

The `go-jni` library generates idiomatic, strongly typed Go wrappers for 53+
Android SDK packages. Every package follows the same patterns described below.
The typed wrappers handle JNI details internally -- you work with Go types,
Go errors, and named constants.

---

## 1. The Typed Wrapper is the Default

For any Android API, the first step is to find the generated wrapper package
and use it directly. No raw JNI needed in the common case.

### System service managers

Most Android APIs are accessed through a manager obtained from an `app.Context`:

```go
import (
    "github.com/AndroidGoLab/jni/app"
    "github.com/AndroidGoLab/jni/telephony"
)

func example(ctx *app.Context) error {
    mgr, err := telephony.NewManager(ctx)
    if err != nil {
        return err
    }
    defer mgr.Close()

    phoneType, err := mgr.GetPhoneType()
    if err != nil {
        return err
    }

    // Use named constants -- never magic numbers.
    // Constants are typed int; cast the int32 return to match.
    switch int(phoneType) {
    case telephony.PhoneTypeGsm:
        fmt.Println("GSM")
    case telephony.PhoneTypeCdma:
        fmt.Println("CDMA")
    }

    operatorName, err := mgr.GetNetworkOperatorName()
    if err != nil {
        return err
    }
    fmt.Println("Operator:", operatorName)

    return nil
}
```

Every manager follows this shape:

```go
mgr, err := somepkg.NewManager(ctx) // or NewAdapter, NewDeviceManager, etc.
defer mgr.Close()
result, err := mgr.SomeMethod(args...)
```

Available managers include `bluetooth.NewAdapter`, `bluetooth.NewManager`,
`location.NewManager`, `telephony.NewManager`, `net/wifi.NewManager`,
`net/wifi/p2p.NewWifiP2pManager`, `nfc.NewManager`, `os/battery.NewManager`,
`os/power.NewManager`, `print.NewManager`, `telecom.NewManager`, and many
more -- one for each Android system service.

### Data classes

Types like `bluetooth.Device`, `p2p.WifiP2pDevice`, `le.ScanResult` are
returned from manager methods. They have the same `{VM, Obj}` shape and
provide typed accessor methods:

```go
import "github.com/AndroidGoLab/jni/bluetooth"

// device is returned from discovery or GATT operations
name, err := device.GetName()
addr, err := device.GetAddress()
devType, err := device.GetType()

// Compare with named constants (cast int32 to int to match constant type)
if int(devType) == bluetooth.DeviceTypeLe {
    fmt.Println("BLE device")
}
```

---

## 2. Named Constants

Every generated package re-exports all Android SDK constants as named Go
values. They live in two locations:

```go
import "github.com/AndroidGoLab/jni/bluetooth"

// Use directly from the parent package:
bluetooth.DeviceTypeClassic          // = 1
bluetooth.BondBonded                 // = 12
bluetooth.GattSuccess                // = 0
bluetooth.StateConnected             // = 2
bluetooth.ConnectionPriorityHigh     // = 1
bluetooth.ScanModeConnectableDiscoverable // = 23
```

Or from the `consts` sub-package when you need to avoid name collisions:

```go
import btconsts "github.com/AndroidGoLab/jni/bluetooth/consts"

btconsts.DeviceTypeClassic
btconsts.GattSuccess
```

**Never use magic numbers.** Every Android constant has a named Go equivalent.
Use it in comparisons, switch statements, and as arguments:

```go
// GOOD: named constant
if int(bondState) == bluetooth.BondBonded { ... }

// BAD: magic number
if bondState == 12 { ... }
```

---

## 3. Method Naming Conventions

Java method names map to Go method names:

| Java method | Go method |
|---|---|
| `getFoo()` | `GetFoo()` |
| `setFoo(x)` | `SetFoo(x)` |
| `isFoo()` | `IsFoo()` |

When Java has overloaded methods (same name, different parameter counts), Go
appends a numeric suffix for arity:

| Java signature | Go method |
|---|---|
| `registerReceiver(BroadcastReceiver, IntentFilter)` | `RegisterReceiver2(...)` |
| `registerReceiver(BroadcastReceiver, IntentFilter, int)` | `RegisterReceiver3_1(...)` |

The `_N` suffix disambiguates when multiple overloads share the same arity.

### Return types

Generated methods convert Java types to Go types automatically:

| Java return | Go return | Notes |
|---|---|---|
| `boolean` | `bool` | Converted from JNI `uint8` internally |
| `int` | `int32` | |
| `long` | `int64` | |
| `float` | `float32` | |
| `double` | `float64` | |
| `String` | `string` | Converted via `GoString` internally |
| Any object | `*jni.Object` | When no typed wrapper exists for the return type |

All methods return `error` as the last value.

---

## 4. Error Handling and API Level Graceful Degradation

```go
result, err := mgr.SomeMethod()
if err != nil {
    // Three common causes:
    // 1. Java exception (automatically converted to Go error)
    // 2. "...not available on this device" -- API doesn't exist at this Android version
    // 3. "...service not available" -- system service not running
}
```

Methods that don't exist on the running Android version return an error
instead of panicking. This lets you write code that works across API levels:

```go
// Try the newer API first, fall back gracefully
result, err := mgr.GetFoo()
if err != nil {
    // Might not exist on older devices -- use fallback
    result, err = mgr.GetLegacyFoo()
}
```

---

## 5. Connecting Typed Wrappers

When a generated method returns `*jni.Object` (untyped) because the return
type belongs to a different package, wrap it into the appropriate typed struct:

```go
import (
    "github.com/AndroidGoLab/jni/bluetooth"
    "github.com/AndroidGoLab/jni/bluetooth/le"
)

// GetBluetoothLeScanner returns *jni.Object (untyped)
scannerObj, err := adapter.GetBluetoothLeScanner()
if err != nil {
    return err
}

// Wrap into the typed struct from the le package
scanner := le.BluetoothLeScanner{VM: vm, Obj: scannerObj}
```

Similarly, when a method **takes** `*jni.Object`, pass the `Obj` field of
another typed wrapper:

```go
// Pass a typed wrapper's Obj to a method expecting *jni.Object
err := gatt.ReadCharacteristic(characteristic.Obj)
```

---

## 6. Callbacks

### Interface callbacks (env.NewProxy)

When an Android API takes a **Java interface** as callback, implement it
entirely in Go with `env.NewProxy()`:

```go
vm.Do(func(env *jni.Env) error {
    listenerClass, err := env.FindClass("android/location/LocationListener")
    if err != nil {
        return err
    }

    proxy, cleanup, err := env.NewProxy(
        []*jni.Class{listenerClass},
        func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
            switch methodName {
            case "onLocationChanged":
                // args[0] is the Location object
            case "onProviderEnabled":
                // ...
            }
            return nil, nil
        },
    )
    if err != nil {
        return err
    }
    defer cleanup()

    // Pass proxy to the API
    // ...
    return nil
})
```

For handlers that also need `java.lang.reflect.Method` introspection (e.g., to
detect void return type), use `env.NewProxyFull()`.

### Abstract class callbacks (GoAbstractDispatch)

When the callback is a **Java abstract class** (like `BroadcastReceiver`,
`ScanCallback`, `CameraDevice.StateCallback`), `env.NewProxy()` cannot be
used. Instead:

1. Write a small Java adapter class extending the abstract class, delegating
   to `GoAbstractDispatch.invoke(handlerID, methodName, args)`
2. From Go, register a handler with `jni.RegisterProxyHandler(handler)` to
   get a handler ID
3. Instantiate the Java adapter via raw JNI, passing the handler ID

```go
handlerID := jni.RegisterProxyHandler(
    func(env *jni.Env, method string, args []*jni.Object) (*jni.Object, error) {
        switch method {
        case "onReceive":
            // handle broadcast
        }
        return nil, nil
    },
)
defer jni.UnregisterProxyHandler(handlerID)
```

See `broadcast-receiver.md` and `bluetooth.md` for complete examples.

**How to tell:** If Android docs say `public abstract class`, you need a Java
adapter. If `public interface`, use `env.NewProxy()`.

---

## 7. Reference Management

**Generated wrappers handle references automatically.** You only need to think
about references when mixing in raw JNI.

Rules:

1. Always `defer mgr.Close()` after creating a manager -- this releases the
   underlying JNI global reference.
2. `*jni.Object` values returned by generated methods are already global refs
   that survive across `vm.Do()` scopes.
3. Inside `vm.Do()`, any raw JNI local refs you create are valid only within
   that block. Convert with `env.NewGlobalRef()` if needed outside.
4. Free global refs with `env.DeleteGlobalRef()` when done.

---

## 8. HandlerThread for Callbacks

Many callback-based Android APIs require a `Looper` thread.
`os.NewHandlerThread` creates and starts the thread in one call:

```go
import "github.com/AndroidGoLab/jni/os"

ht, err := os.NewHandlerThread(vm, "GoCallbackThread")
if err != nil {
    return err
}
defer ht.Close()

looperObj, err := ht.GetLooper()
if err != nil {
    return err
}
// Pass looperObj when registering callbacks...

```

---

## 9. When Raw JNI Is Needed

Raw JNI is the fallback for the rare cases not covered by typed wrappers:
- **Class references** for `env.NewProxy()` interface lookups

Even then, prefer typed wrappers for everything they cover and drop to raw JNI
only for the gap.

### JNI type signatures (for GetMethodID)

| Java type | Sig | Java type | Sig |
|---|---|---|---|
| `boolean` | `Z` | `String` | `Ljava/lang/String;` |
| `int` | `I` | `Object` | `Ljava/lang/Object;` |
| `long` | `J` | `int[]` | `[I` |
| `float` | `F` | `void` | `V` |

Method format: `(param-types)return-type`. E.g., `(Ljava/lang/String;I)V`
means takes a String and an int, returns void.

### Value constructors (for raw JNI call arguments)

```go
jni.IntValue(int32(42))
jni.LongValue(int64(1000))
jni.FloatValue(float32(3.14))
jni.BooleanValue(uint8(1))     // JNI booleans are uint8, not Go bool
jni.ObjectValue(&someObj.Object)
```

---

## 10. Common Pitfalls

1. **Using local refs outside `vm.Do()`** -- Local refs are invalidated when
   the block returns. Convert to global refs if you need them later.

2. **Forgetting `defer mgr.Close()`** -- Global refs prevent Java GC. Leaked
   refs accumulate until the JVM global ref table overflows.

3. **Magic numbers instead of constants** -- Every Android constant has a
   named Go equivalent in the package or its `consts` sub-package. Use it.

4. **Calling callback APIs without a Looper** -- APIs like `LocationManager`,
   `WifiP2pManager` require callbacks on a Looper thread. Create a
   `HandlerThread` (Section 8).

5. **Using `env.NewProxy()` for abstract classes** -- `NewProxy` only works
   for Java interfaces. For abstract classes, use the `GoAbstractDispatch`
   Java adapter pattern (Section 6).

6. **Thread affinity** -- `*jni.Env` is thread-local. Never pass it across
   goroutines. Always get a fresh one inside `vm.Do()`.

7. **Proxy ClassLoader in APK mode** -- When using `env.NewProxy()` inside a
   packaged APK, call `jni.SetProxyClassLoader()` first with the APK's
   ClassLoader. Without it, the proxy infrastructure can't find its helper
   classes on native threads.
