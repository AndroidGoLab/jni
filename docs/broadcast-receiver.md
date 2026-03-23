# BroadcastReceiver and Callbacks from Go

This guide shows how to register a `BroadcastReceiver` from Go and receive
system broadcasts (WiFi state changes, Bluetooth discovery, battery updates,
etc.) using `Context.registerReceiver`.

## Why BroadcastReceiver needs a Java adapter

`BroadcastReceiver` is an **abstract class**, not an interface.
`env.NewProxy()` only works for Java interfaces. For abstract classes, this
library provides the `GoAbstractDispatch` pattern: you write a small Java
class that extends the abstract class and delegates each method to
`GoAbstractDispatch.invoke(handlerID, methodName, args)`. The Go runtime
routes the call to your registered Go handler.

## Complete BroadcastReceiver pattern

### Step 1: Java adapter class

Add this file to your APK build (e.g. alongside
`center/dx/jni/internal/GoAbstractDispatch.java`):

**`center/dx/jni/generated/GoBroadcastReceiver.java`**

```java
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
        GoAbstractDispatch.invoke(handlerID, "onReceive", new Object[]{context, intent});
    }
}
```

This follows the exact pattern used by `CameraDeviceCallback.java` in the
library's examples. The constructor takes a `long handlerID`; `onReceive`
forwards its arguments through the native dispatch bridge.

### Step 2: Go code — register, listen, clean up

```go
package main

import (
    "fmt"
    "log"
    "unsafe"

    "github.com/AndroidGoLab/jni"
    "github.com/AndroidGoLab/jni/app"
)

// listenForWifiStateChanges registers a BroadcastReceiver for
// android.net.wifi.WIFI_STATE_CHANGED_ACTION and prints state changes.
//
// activity must be a global ref to the Activity/Context.
func listenForWifiStateChanges(vm *jni.VM, activity *jni.Object) (cleanup func(), err error) {
    // 1. Register the Go handler that will receive onReceive calls.
    handlerID := jni.RegisterProxyHandler(
        func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
            if methodName != "onReceive" {
                return nil, nil
            }
            // args[0] = Context, args[1] = Intent

            // Use the typed Intent wrapper to read action and extras.
            intent := &app.Intent{VM: vm, Obj: args[1]}
            action, err := intent.GetAction()
            if err != nil {
                return nil, err
            }
            fmt.Printf("Broadcast received: action=%s\n", action)

            // Read an int extra (e.g. wifi_state).
            const defaultWifiState = -1
            state, err := intent.GetIntExtra("wifi_state", defaultWifiState)
            if err != nil {
                return nil, err
            }
            fmt.Printf("  wifi_state=%d\n", state)

            return nil, nil
        },
    )

    // 2. Create an IntentFilter for the desired action.
    filter, err := content.NewIntentFilter(vm, "android.net.wifi.WIFI_STATE_CHANGED")
    if err != nil {
        jni.UnregisterProxyHandler(handlerID)
        return nil, fmt.Errorf("new IntentFilter: %w", err)
    }

    // 3. Instantiate GoBroadcastReceiver (custom Java adapter -- raw JNI required).
    var receiverGlobal *jni.Object
    err = vm.Do(func(env *jni.Env) error {
        if err := jni.EnsureProxyInit(env); err != nil {
            return fmt.Errorf("EnsureProxyInit: %w", err)
        }

        // GoBroadcastReceiver is a user-defined Java class. In NativeActivity,
        // FindClass uses the boot ClassLoader which cannot see APK classes.
        // Use the Activity's ClassLoader instead.
        actCls := env.GetObjectClass(activity)
        getClassLoaderMid, err := env.GetMethodID(actCls, "getClassLoader",
            "()Ljava/lang/ClassLoader;")
        if err != nil {
            return fmt.Errorf("get getClassLoader: %w", err)
        }
        classLoader, err := env.CallObjectMethod(activity, getClassLoaderMid)
        if err != nil {
            return fmt.Errorf("getClassLoader: %w", err)
        }

        clCls := env.GetObjectClass(classLoader)
        loadClassMid, err := env.GetMethodID(clCls, "loadClass",
            "(Ljava/lang/String;)Ljava/lang/Class;")
        if err != nil {
            return fmt.Errorf("get loadClass: %w", err)
        }

        jName, err := env.NewStringUTF("center.dx.jni.generated.GoBroadcastReceiver")
        if err != nil {
            return err
        }
        defer env.DeleteLocalRef(&jName.Object)
        receiverCls, err := env.CallObjectMethod(classLoader, loadClassMid,
            jni.ObjectValue(&jName.Object))
        if err != nil {
            return fmt.Errorf("loadClass(GoBroadcastReceiver): %w", err)
        }

        cls := (*jni.Class)(unsafe.Pointer(receiverCls))
        initMid, err := env.GetMethodID(cls, "<init>", "(J)V")
        if err != nil {
            return fmt.Errorf("get GoBroadcastReceiver.<init>: %w", err)
        }
        receiver, err := env.NewObject(cls, initMid, jni.LongValue(handlerID))
        if err != nil {
            return fmt.Errorf("new GoBroadcastReceiver: %w", err)
        }
        receiverGlobal = env.NewGlobalRef(receiver)
        return nil
    })
    if err != nil {
        filter.Close()
        jni.UnregisterProxyHandler(handlerID)
        return nil, err
    }

    // 4. Register the receiver with the Context.
    ctx := &app.Context{VM: vm, Obj: activity}
    _, err = ctx.RegisterReceiver2(receiverGlobal, filter.Obj)
    if err != nil {
        filter.Close()
        jni.UnregisterProxyHandler(handlerID)
        return nil, fmt.Errorf("registerReceiver: %w", err)
    }

    // 5. Return a cleanup function.
    cleanup = func() {
        ctx := &app.Context{VM: vm, Obj: activity}
        ctx.UnregisterReceiver(receiverGlobal)
        vm.Do(func(env *jni.Env) error {
            env.DeleteGlobalRef(receiverGlobal)
            return nil
        })
        filter.Close()
        jni.UnregisterProxyHandler(handlerID)
    }
    return cleanup, nil
}
```

**Key points:**

- `jni.RegisterProxyHandler(handler)` returns an `int64` handler ID that the
  Java adapter stores and passes to `GoAbstractDispatch.invoke(...)` on each
  call.
- `jni.UnregisterProxyHandler(id)` removes the handler when done.
- `app.Intent.GetAction()` and `app.Intent.GetIntExtra(key, default)` replace
  manual `GetMethodID`/`CallObjectMethod`/`CallIntMethod` chains for reading
  intent data.
- `app.Context.RegisterReceiver2(receiver, filter)` calls
  `android.content.Context.registerReceiver(BroadcastReceiver, IntentFilter)`.
- `app.Context.UnregisterReceiver(receiver)` calls
  `android.content.Context.unregisterReceiver(BroadcastReceiver)`.
- The receiver object **must** be a global ref because `registerReceiver`
  stores it beyond the current JNI scope.

### Multiple actions on one receiver

To listen for multiple broadcast actions, create a no-arg IntentFilter and
call `addAction` for each:

```go
// Create an empty filter and add multiple actions:
filter, _ := content.NewIntentFilter(vm, "")
filter.AddAction("android.net.wifi.WIFI_STATE_CHANGED")
filter.AddAction("android.net.wifi.SCAN_RESULTS")
filter.AddAction("android.bluetooth.device.action.FOUND")
defer filter.Close()

ctx := &app.Context{VM: vm, Obj: activity}
ctx.RegisterReceiver2(receiverGlobal, filter.Obj)
```

## Interface-based callbacks with env.NewProxy

When the callback is a **Java interface** (not an abstract class),
`env.NewProxy()` works directly -- no Java adapter needed:

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
                // Use the typed Location wrapper for latitude/longitude.
                loc := &location.Location{VM: vm, Obj: args[0]}
                lat, _ := loc.GetLatitude()
                lon, _ := loc.GetLongitude()
                fmt.Printf("Location: %.6f, %.6f\n", lat, lon)
            case "onProviderEnabled":
                fmt.Println("Provider enabled")
            case "onProviderDisabled":
                fmt.Println("Provider disabled")
            }
            return nil, nil
        },
    )
    if err != nil {
        return err
    }
    defer cleanup()

    // Use proxy with LocationManager.requestLocationUpdates, etc.
    _ = proxy
    return nil
})
```

## HandlerThread for callback delivery

Android callbacks require a `Looper` thread. Create a `HandlerThread` to
provide one:

```go
// NewHandlerThread creates and starts the thread in one call.
ht, err := os.NewHandlerThread(vm, "GoCallbackThread")
if err != nil {
    return err
}

// Get its Looper for registering callbacks.
looperObj, err := ht.GetLooper()
if err != nil {
    ht.Close()
    return err
}
// looperObj can be passed to APIs that need a Looper for callback delivery.
_ = looperObj
```

To shut down the thread:

```go
ht.Close() // quits the thread safely and releases the global ref
```

## Interfaces vs abstract classes reference

| Android type | Kind | Go approach |
|---|---|---|
| `LocationListener` | interface | `env.NewProxy()` |
| `Runnable` | interface | `env.NewProxy()` |
| `WifiP2pManager.ActionListener` | interface | `env.NewProxy()` |
| `BroadcastReceiver` | abstract class | Java adapter + `GoAbstractDispatch` |
| `CameraDevice.StateCallback` | abstract class | Java adapter + `GoAbstractDispatch` |
| `CameraCaptureSession.StateCallback` | abstract class | Java adapter + `GoAbstractDispatch` |
| `PhoneStateListener` | abstract class | Java adapter + `GoAbstractDispatch` |

The rule: if javadoc says `public abstract class`, you need a Java adapter.
If it says `public interface`, use `env.NewProxy()`.
