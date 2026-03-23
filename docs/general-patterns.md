# General Patterns: Calling Any Android API from Go

This guide covers how to approach Android APIs that don't have a dedicated example.
The `go-jni` library auto-generates typed Go wrappers for 53+ Android SDK packages.
Every generated package follows the same structural patterns described below.

---

## Decision Tree

```
I want to call Android API X. What do I do?

1. Is there a generated wrapper package for it?
   ├─ YES → Use the typed wrapper (Section 1 or 2)
   └─ NO  → Use raw JNI calls (Section 3 or 4)

2. Does the API require a callback?
   ├─ Java interface → Use env.NewProxy() (Section 5)
   └─ Java abstract class → Use GoAbstractDispatch adapter (Section 6)

3. Does the callback need a Looper thread?
   └─ YES → Create a HandlerThread first (Section 8)
```

---

## 1. System Service Pattern

Most Android "manager" classes are obtained via system service. Every generated
manager wrapper follows the same shape:

```go
import "github.com/AndroidGoLab/jni/net/wifi/p2p"

mgr, err := p2p.NewWifiP2pManager(ctx) // ctx is *app.Context
if err != nil {
    return err
}
defer mgr.Close()

// Call typed methods directly:
// mgr.DiscoverPeers(channel, listener)
```

Under the hood, `NewWifiP2pManager` calls `ctx.GetSystemService("wifip2p")`,
which returns a `*jni.GlobalRef`. The manager struct wraps it:

```go
type WifiP2pManager struct {
    VM  *jni.VM
    Ctx *app.Context
    Obj *jni.GlobalRef
}
```

`Close()` calls `env.DeleteGlobalRef(m.Obj)`. Always defer it.

---

## 2. Data Class Pattern

Types like `WifiP2pDevice`, `BluetoothDevice` -- you never construct these directly.
They are returned from other API calls. The struct shape is:

```go
type WifiP2pDevice struct {
    VM  *jni.VM
    Obj *jni.GlobalRef
}
```

Methods call `env.Call*Method` on `Obj`:

```go
name, err := device.GetDeviceName()
addr, err := device.GetDeviceAddress()
```

When you receive a raw `*jni.Object` from a method and need to wrap it into a
typed struct, see Section 10.

---

## 3. Constructor Pattern (Raw JNI)

For classes not exposed as system services, construct them with `env.NewObject()`:

```go
vm.Do(func(env *jni.Env) error {
    cls, err := env.FindClass("android/content/IntentFilter")
    if err != nil {
        return err
    }
    initMid, err := env.GetMethodID(cls, "<init>", "(Ljava/lang/String;)V")
    if err != nil {
        return err
    }
    actionStr, err := env.NewStringUTF("android.net.wifi.p2p.PEERS_CHANGED")
    if err != nil {
        return err
    }
    obj, err := env.NewObject(cls, initMid, jni.ObjectValue(&actionStr.Object))
    if err != nil {
        return err
    }
    // obj is a local ref -- valid only inside this vm.Do() block.
    // Convert to global ref if you need it outside:
    globalObj := env.NewGlobalRef(obj)
    // ... use globalObj ...
    // Eventually: env.DeleteGlobalRef(globalObj)
    return nil
})
```

---

## 4. Static Method Pattern (Raw JNI)

```go
import "unsafe" // required for (*jni.String)(unsafe.Pointer(...)) casts

vm.Do(func(env *jni.Env) error {
    cls, err := env.FindClass("android/os/Build")
    if err != nil {
        return err
    }
    mid, err := env.GetStaticMethodID(cls, "getRadioVersion", "()Ljava/lang/String;")
    if err != nil {
        return err
    }
    result, err := env.CallStaticObjectMethod(cls, mid)
    if err != nil {
        return err
    }
    version := env.GoString((*jni.String)(unsafe.Pointer(result)))
    fmt.Println("Radio version:", version)
    return nil
})
```

---

## 5. Interface Callback Pattern

When an Android API takes a Java interface (e.g., `LocationListener`,
`ActionListener`), implement it in Go using `env.NewProxy()`:

```go
vm.Do(func(env *jni.Env) error {
    listenerClass, err := env.FindClass("android/net/wifi/p2p/WifiP2pManager$ActionListener")
    if err != nil {
        return err
    }

    proxy, cleanup, err := env.NewProxy(
        []*jni.Class{listenerClass},
        func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
            switch methodName {
            case "onSuccess":
                fmt.Println("success")
            case "onFailure":
                fmt.Println("failure")
            }
            return nil, nil
        },
    )
    if err != nil {
        return err
    }
    defer cleanup()

    // Pass proxy to the API that expects this interface
    // ...
    return nil
})
```

`NewProxy` creates a `java.lang.reflect.Proxy` backed by a
`GoInvocationHandler` that dispatches to your Go function.

For handlers that also need the `java.lang.reflect.Method` object (e.g., to
inspect return type), use `env.NewProxyFull()` instead.

---

## 6. Abstract Class Callback Pattern

Java abstract classes cannot be proxied with `java.lang.reflect.Proxy`.
Instead, use a Java adapter class that extends the abstract class and delegates
to `GoAbstractDispatch.invoke(handlerID, methodName, args)`.

From Go:

```go
// Register a handler -- returns a unique ID
handlerID := jni.RegisterProxyHandler(
    func(env *jni.Env, method string, args []*jni.Object) (*jni.Object, error) {
        switch method {
        case "onOpened":
            fmt.Println("camera opened")
        case "onError":
            fmt.Println("camera error")
        }
        return nil, nil
    },
)
defer jni.UnregisterProxyHandler(handlerID)

// Then construct the Java adapter via JNI, passing handlerID as a jlong:
// env.NewObject(adapterClass, adapterInit, jni.LongValue(handlerID))
```

The adapter's Java code calls `GoAbstractDispatch.invoke(handlerID, ...)`,
which dispatches to the registered Go handler.

---

## 7. Reference Management

**Local refs** are valid only within the `vm.Do()` block where they were created.
The JVM may garbage-collect the underlying Java object once the local ref is gone.

**Global refs** survive across `vm.Do()` calls. You must free them explicitly.

Rules:

1. Generated wrapper methods handle refs automatically. You don't need to
   manage refs when using typed wrappers.
2. When using raw JNI inside `vm.Do()`, local refs are auto-valid for the
   block's duration.
3. To keep an object beyond `vm.Do()`, convert it:
   ```go
   globalRef := env.NewGlobalRef(localObj)
   ```
4. Always free global refs when done:
   ```go
   env.DeleteGlobalRef(globalRef)
   ```
5. Manager wrappers' `Close()` method handles this -- always `defer mgr.Close()`.

---

## 8. HandlerThread for Callbacks

Many callback-based Android APIs require a `Looper` thread. Create one:

```go
vm.Do(func(env *jni.Env) error {
    htClass, err := env.FindClass("android/os/HandlerThread")
    if err != nil {
        return err
    }
    htInit, err := env.GetMethodID(htClass, "<init>", "(Ljava/lang/String;)V")
    if err != nil {
        return err
    }
    name, err := env.NewStringUTF("MyCallbackThread")
    if err != nil {
        return err
    }
    ht, err := env.NewObject(htClass, htInit, jni.ObjectValue(&name.Object))
    if err != nil {
        return err
    }
    handlerThread := env.NewGlobalRef(ht)

    // Start the thread
    startMid, err := env.GetMethodID(htClass, "start", "()V")
    if err != nil {
        return err
    }
    err = env.CallVoidMethod(handlerThread, startMid)
    if err != nil {
        return err
    }

    // Get its Looper
    getLooperMid, err := env.GetMethodID(htClass, "getLooper", "()Landroid/os/Looper;")
    if err != nil {
        return err
    }
    looper, err := env.CallObjectMethod(handlerThread, getLooperMid)
    if err != nil {
        return err
    }
    // Use looper when constructing Handler or registering callbacks
    _ = looper
    return nil
})
```

---

## 9. Discovering Available Methods in a Package

Each generated package exports typed Go methods. To see what's available:

**Check the Go source files** in the package directory. Method names follow
these conventions:

| Java method          | Go method                |
|---------------------|--------------------------|
| `getFoo()`          | `GetFoo()`               |
| `setFoo(x)`        | `SetFoo(x)`              |
| `isFoo()`          | `IsFoo()`                |
| `foo()` (overloaded, 2 args) | `Foo2()`        |
| `foo()` (overloaded, 3 args, variant 1) | `Foo3_1()` |

Overloaded methods get a numeric suffix for arity, and `_N` for disambiguation
when multiple overloads have the same arity.

---

## 10. Mixing Generated Wrappers with Raw JNI

When a generated method returns `*jni.Object` (untyped), wrap it into the
appropriate typed struct:

```go
// Generated method returns raw object
obj, err := bluetoothAdapter.GetBluetoothLeScanner()
if err != nil {
    return err
}

// Wrap into typed struct from the le package
scanner := le.BluetoothLeScanner{VM: vm, Obj: obj}
defer scanner.Close() // if the type has Close()
```

When a generated method **takes** `*jni.Object` as a parameter, you must
construct the correct Java object type via raw JNI and pass it:

```go
vm.Do(func(env *jni.Env) error {
    // Build the Java object the method expects
    cls, _ := env.FindClass("android/net/wifi/p2p/WifiP2pConfig")
    initMid, _ := env.GetMethodID(cls, "<init>", "()V")
    config, _ := env.NewObject(cls, initMid)

    // Pass it to the generated method that takes *jni.Object
    err := mgr.Connect(channel, &config.Object, listener)
    return err
})
```

---

## 11. Error Handling

All generated methods return `error` as the last value.

```go
result, err := mgr.SomeMethod()
if err != nil {
    // Three common causes:
    // 1. Java exception (automatically converted to Go error)
    // 2. API not available: "...not available on this device"
    // 3. Service not available: "...service not available"
}
```

Methods gracefully return an error when the underlying API doesn't exist on the
running Android version (checked via nil method ID).

---

## 12. JNI Type Signatures Quick Reference

Use these when calling `GetMethodID`, `GetStaticMethodID`, etc.

### Primitive types

| Java type | Signature |
|-----------|-----------|
| `boolean` | `Z`       |
| `byte`    | `B`       |
| `char`    | `C`       |
| `short`   | `S`       |
| `int`     | `I`       |
| `long`    | `J`       |
| `float`   | `F`       |
| `double`  | `D`       |
| `void`    | `V`       |

### Object types

| Java type  | Signature                  |
|------------|---------------------------|
| `String`   | `Ljava/lang/String;`      |
| `Object`   | `Ljava/lang/Object;`      |
| `int[]`    | `[I`                      |
| `String[]` | `[Ljava/lang/String;`     |

### Method signature format

`(parameter-types)return-type`

Examples:
- `void foo()` -> `()V`
- `int bar(String s)` -> `(Ljava/lang/String;)I`
- `String baz(long x, float y)` -> `(JF)Ljava/lang/String;`
- `void qux(int[] a, String b)` -> `([ILjava/lang/String;)V`

---

## 13. Value Constructors

When passing arguments to raw JNI method calls, use these typed constructors
from the `jni` package:

```go
jni.IntValue(int32(42))
jni.LongValue(int64(1000))
jni.FloatValue(float32(3.14))
jni.DoubleValue(float64(2.718))
jni.BooleanValue(uint8(1))     // JNI booleans are uint8, not Go bool
jni.ByteValue(int8(0x7F))
jni.CharValue(uint16('A'))
jni.ShortValue(int16(100))
jni.ObjectValue(&someObj.Object)
```

Note: `BooleanValue` takes `uint8`, not `bool`. Use `1` for true, `0` for false.

---

## 14. Boolean Method Results

JNI booleans are `uint8`. Generated wrappers convert to Go `bool` automatically.

In raw JNI calls, convert manually:

```go
raw, err := env.CallBooleanMethod(obj, mid) // returns uint8
if err != nil {
    return err
}
result := raw != 0 // convert to Go bool
```

---

## Common Pitfalls

1. **Using local refs outside `vm.Do()`** -- Local refs are invalidated when
   the `vm.Do()` block returns. Convert to global refs if you need them later.

2. **Forgetting to call `Close()` / `DeleteGlobalRef()`** -- Global refs prevent
   Java GC. Leaked global refs accumulate until the JVM global ref table overflows.

3. **Wrong JNI signature string** -- A typo in the signature passed to
   `GetMethodID` silently returns an error. Double-check against the Android
   SDK docs or use `javap -s` on the class.

4. **Calling callback APIs without a Looper** -- Many Android APIs
   (`WifiP2pManager`, `LocationManager`, `BluetoothAdapter`) require callbacks
   to run on a Looper thread. Create a `HandlerThread` (Section 8).

5. **Passing Go `bool` to `BooleanValue`** -- `BooleanValue` takes `uint8`.
   Passing `true`/`false` won't compile.

6. **Thread affinity** -- `*jni.Env` is thread-local. Never pass an `Env`
   across goroutines. Always obtain a fresh `Env` inside `vm.Do()`.

7. **Ignoring "not available on this device" errors** -- Generated methods
   return this when the Android API level doesn't support the method. Check
   errors and have fallback behavior for older devices.

8. **Proxy ClassLoader in APK mode** -- When using `env.NewProxy()` inside a
   packaged APK, call `jni.SetProxyClassLoader()` first with the APK's
   ClassLoader (from `Context.getClassLoader()`). Without it, the proxy
   infrastructure can't find its helper classes on native threads.
