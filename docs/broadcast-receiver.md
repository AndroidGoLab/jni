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
            intent := args[1]

            // Read the intent action.
            intentCls := env.GetObjectClass(intent)
            getActionMid, err := env.GetMethodID(intentCls, "getAction",
                "()Ljava/lang/String;")
            if err != nil {
                return nil, err
            }
            actionObj, err := env.CallObjectMethod(intent, getActionMid)
            if err != nil {
                return nil, err
            }
            action := env.GoString((*jni.String)(unsafe.Pointer(actionObj)))
            fmt.Printf("Broadcast received: action=%s\n", action)

            // Read an int extra (e.g. wifi_state).
            getIntExtraMid, err := env.GetMethodID(intentCls, "getIntExtra",
                "(Ljava/lang/String;I)I")
            if err != nil {
                return nil, err
            }
            keyStr, err := env.NewStringUTF("wifi_state")
            if err != nil {
                return nil, err
            }
            defer env.DeleteLocalRef(&keyStr.Object)
            state, err := env.CallIntMethod(intent, getIntExtraMid,
                jni.ObjectValue(&keyStr.Object), jni.IntValue(-1))
            if err != nil {
                return nil, err
            }
            fmt.Printf("  wifi_state=%d\n", state)

            return nil, nil
        },
    )

    var receiverGlobal *jni.Object
    var filterGlobal *jni.Object

    err = vm.Do(func(env *jni.Env) error {
        // Ensure proxy native methods are registered.
        if err := jni.EnsureProxyInit(env); err != nil {
            return fmt.Errorf("EnsureProxyInit: %w", err)
        }

        // 2. Load the GoBroadcastReceiver class via ClassLoader.
        //    In NativeActivity, FindClass uses the boot ClassLoader which
        //    cannot see APK classes. Use the Activity's ClassLoader instead.
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

        // 3. Instantiate GoBroadcastReceiver(handlerID).
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

        // 4. Create an IntentFilter with the desired action.
        ifClass, err := env.FindClass("android/content/IntentFilter")
        if err != nil {
            return fmt.Errorf("find IntentFilter: %w", err)
        }
        ifInit, err := env.GetMethodID(ifClass, "<init>", "(Ljava/lang/String;)V")
        if err != nil {
            return fmt.Errorf("get IntentFilter.<init>: %w", err)
        }
        actionStr, err := env.NewStringUTF("android.net.wifi.WIFI_STATE_CHANGED")
        if err != nil {
            return err
        }
        defer env.DeleteLocalRef(&actionStr.Object)
        filter, err := env.NewObject(ifClass, ifInit, jni.ObjectValue(&actionStr.Object))
        if err != nil {
            return fmt.Errorf("new IntentFilter: %w", err)
        }
        filterGlobal = env.NewGlobalRef(filter)

        // 5. Call context.registerReceiver(receiver, filter).
        ctx := &app.Context{VM: vm, Obj: activity}
        _, err = ctx.RegisterReceiver2(receiverGlobal, filterGlobal)
        if err != nil {
            return fmt.Errorf("registerReceiver: %w", err)
        }

        return nil
    })
    if err != nil {
        jni.UnregisterProxyHandler(handlerID)
        return nil, err
    }

    // 6. Return a cleanup function.
    cleanup = func() {
        vm.Do(func(env *jni.Env) error {
            ctx := &app.Context{VM: vm, Obj: activity}
            ctx.UnregisterReceiver(receiverGlobal)
            env.DeleteGlobalRef(receiverGlobal)
            env.DeleteGlobalRef(filterGlobal)
            return nil
        })
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
// Using the content.IntentFilter wrapper:
vm.Do(func(env *jni.Env) error {
    ifClass, _ := env.FindClass("android/content/IntentFilter")
    ifInit, _ := env.GetMethodID(ifClass, "<init>", "()V")
    filterLocal, _ := env.NewObject(ifClass, ifInit)
    filterObj := env.NewGlobalRef(filterLocal)

    // Use the typed wrapper for addAction:
    filter := &content.IntentFilter{VM: vm, Obj: filterObj}
    filter.AddAction("android.net.wifi.WIFI_STATE_CHANGED")
    filter.AddAction("android.net.wifi.SCAN_RESULTS")
    filter.AddAction("android.bluetooth.device.action.FOUND")

    ctx := &app.Context{VM: vm, Obj: activity}
    ctx.RegisterReceiver2(receiverGlobal, filterObj)
    return nil
})
```

This requires `import "github.com/AndroidGoLab/jni/content"`.

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
                // args[0] is android.location.Location
                locCls := env.GetObjectClass(args[0])
                getLatMid, _ := env.GetMethodID(locCls, "getLatitude", "()D")
                getLonMid, _ := env.GetMethodID(locCls, "getLongitude", "()D")
                lat, _ := env.CallDoubleMethod(args[0], getLatMid)
                lon, _ := env.CallDoubleMethod(args[0], getLonMid)
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
var handlerThread *jni.Object // global ref
var handler *jni.Object       // global ref

vm.Do(func(env *jni.Env) error {
    // Create and start a HandlerThread.
    htClass, _ := env.FindClass("android/os/HandlerThread")
    htInit, _ := env.GetMethodID(htClass, "<init>", "(Ljava/lang/String;)V")
    name, _ := env.NewStringUTF("GoCallbackThread")
    defer env.DeleteLocalRef(&name.Object)
    ht, _ := env.NewObject(htClass, htInit, jni.ObjectValue(&name.Object))
    handlerThread = env.NewGlobalRef(ht)

    startMid, _ := env.GetMethodID(htClass, "start", "()V")
    env.CallVoidMethod(handlerThread, startMid)

    // Get the Looper and create a Handler.
    getLooperMid, _ := env.GetMethodID(htClass, "getLooper",
        "()Landroid/os/Looper;")
    looper, _ := env.CallObjectMethod(handlerThread, getLooperMid)

    hClass, _ := env.FindClass("android/os/Handler")
    hInit, _ := env.GetMethodID(hClass, "<init>", "(Landroid/os/Looper;)V")
    h, _ := env.NewObject(hClass, hInit, jni.ObjectValue(looper))
    handler = env.NewGlobalRef(h)

    return nil
})

// Use registerReceiver(receiver, filter, null, handler) for delivery
// on this thread instead of the main thread. The 4-arg overload is
// RegisterReceiver3_1 in the generated bindings, which takes
// (BroadcastReceiver, IntentFilter, flags int32) -- or use raw JNI
// to call the 4-arg registerReceiver.
```

To shut down the thread:

```go
vm.Do(func(env *jni.Env) error {
    htClass := env.GetObjectClass(handlerThread)
    quitMid, _ := env.GetMethodID(htClass, "quitSafely", "()Z")
    env.CallBooleanMethod(handlerThread, quitMid)
    env.DeleteGlobalRef(handlerThread)
    env.DeleteGlobalRef(handler)
    return nil
})
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
