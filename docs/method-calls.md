# Method Calls via JNI

This guide covers calling Java instance methods, static methods, and constructors from Go.

## Instance Methods

Look up a method by name and JNI signature, then call it:

```go
vm.Do(func(env *jni.Env) error {
    cls, err := env.FindClass("android/bluetooth/BluetoothAdapter")
    if err != nil {
        return err
    }

    // Look up the method ID
    mid, err := env.GetMethodID(cls, "getName", "()Ljava/lang/String;")
    if err != nil {
        return err
    }

    // Call it on an instance
    resultObj, err := env.CallObjectMethod(adapterObj, mid)
    if err != nil {
        return err
    }

    // Convert the returned String to Go
    name := env.GoString((*jni.String)(unsafe.Pointer(resultObj)))
    fmt.Println("Name:", name)
    return nil
})
```

### Typed Call Methods

| Method | Returns | Use For |
|--------|---------|---------|
| `env.CallBooleanMethod(obj, mid, args...)` | `uint8, error` | `boolean` returns |
| `env.CallByteMethod(obj, mid, args...)` | `int8, error` | `byte` returns |
| `env.CallCharMethod(obj, mid, args...)` | `uint16, error` | `char` returns |
| `env.CallShortMethod(obj, mid, args...)` | `int16, error` | `short` returns |
| `env.CallIntMethod(obj, mid, args...)` | `int32, error` | `int` returns |
| `env.CallLongMethod(obj, mid, args...)` | `int64, error` | `long` returns |
| `env.CallFloatMethod(obj, mid, args...)` | `float32, error` | `float` returns |
| `env.CallDoubleMethod(obj, mid, args...)` | `float64, error` | `double` returns |
| `env.CallObjectMethod(obj, mid, args...)` | `*Object, error` | Object returns |
| `env.CallVoidMethod(obj, mid, args...)` | `error` | `void` returns |

### Boolean Conversion

JNI represents booleans as `uint8`. Convert to Go `bool`:

```go
resultRaw, err := env.CallBooleanMethod(obj, mid)
if err != nil {
    return err
}
result := resultRaw != 0 // Convert to Go bool
```

## Static Methods

Use `GetStaticMethodID` and `CallStatic*Method`:

```go
vm.Do(func(env *jni.Env) error {
    cls, err := env.FindClass("android/os/Build")
    if err != nil {
        return err
    }

    // Static method returning String
    mid, err := env.GetStaticMethodID(cls, "getRadioVersion", "()Ljava/lang/String;")
    if err != nil {
        return err
    }
    resultObj, err := env.CallStaticObjectMethod(cls, mid)
    if err != nil {
        return err
    }
    version := env.GoString((*jni.String)(unsafe.Pointer(resultObj)))

    // Static method returning int
    mid2, err := env.GetStaticMethodID(cls, "getMajorSdkVersion", "(I)I")
    if err != nil {
        return err
    }
    major, err := env.CallStaticIntMethod(cls, mid2, jni.IntValue(36))

    return nil
})
```

## Constructors

Call constructors with `NewObject`:

```go
vm.Do(func(env *jni.Env) error {
    // Create a HandlerThread with a name argument
    cls, err := env.FindClass("android/os/HandlerThread")
    if err != nil {
        return err
    }
    initMid, err := env.GetMethodID(cls, "<init>", "(Ljava/lang/String;)V")
    if err != nil {
        return err
    }
    name, err := env.NewStringUTF("MyThread")
    if err != nil {
        return err
    }
    obj, err := env.NewObject(cls, initMid, jni.ObjectValue(&name.Object))
    if err != nil {
        return err
    }
    // obj is a local reference to the new HandlerThread
    return nil
})
```

## Passing Arguments

Wrap arguments using typed value constructors:

```go
// requestLocationUpdates(String provider, long minTime, float minDistance,
//                        LocationListener listener, Looper looper)
err := env.CallVoidMethod(mgrObj, reqMid,
    jni.ObjectValue(&providerStr.Object),  // String
    jni.LongValue(0),                      // long
    jni.FloatValue(0),                     // float
    jni.ObjectValue(listenerObj),           // LocationListener
    jni.ObjectValue(looperObj),             // Looper
)
```

## Using Generated Wrappers

The 53 generated packages eliminate manual JNI lookups. Instead of raw JNI:

```go
// Raw JNI (manual)
vm.Do(func(env *jni.Env) error {
    cls, _ := env.FindClass("android/telephony/TelephonyManager")
    mid, _ := env.GetMethodID(cls, "getNetworkOperatorName", "()Ljava/lang/String;")
    obj, _ := env.CallObjectMethod(mgrObj, mid)
    name := env.GoString((*jni.String)(unsafe.Pointer(obj)))
    return nil
})
```

Use the typed wrapper:

```go
// Generated wrapper (type-safe)
import "github.com/AndroidGoLab/jni/telephony"

mgr, _ := telephony.NewManager(ctx)
defer mgr.Close()
name, _ := mgr.GetNetworkOperatorName()
```

## Method ID Caching

Generated packages cache method IDs on first use via `ensureInit()`. This avoids repeated lookups:

```go
// Inside generated code (e.g., telephony/manager.go):
var midManagerGetPhoneType jni.MethodID

func ensureInit(env *jni.Env) error {
    // Called once per process, caches all method/field IDs
    // ...
}

func (m *Manager) GetPhoneType() (int32, error) {
    var result int32
    callErr := m.VM.Do(func(env *jni.Env) error {
        if err := ensureInit(env); err != nil {
            return err
        }
        if midManagerGetPhoneType == nil {
            return fmt.Errorf("method not available on this device")
        }
        result, callErr = env.CallIntMethod(m.Obj, midManagerGetPhoneType)
        return callErr
    })
    return result, callErr
}
```

## Context.GetSystemService

The `app.Context.GetSystemService` method is how all managers obtain their underlying Java object:

```go
import "github.com/AndroidGoLab/jni/app"

ctx, _ := app.ContextFromObject(vm, activityRef)
defer ctx.Close()

// GetSystemService calls Context.getSystemService(String) via JNI,
// converts the returned local reference to a global reference,
// and returns it.
svcObj, err := ctx.GetSystemService("battery")
```

Each `NewManager(ctx)` call does this internally with the appropriate service name.
