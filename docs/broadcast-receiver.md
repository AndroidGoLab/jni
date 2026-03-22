# BroadcastReceiver and IntentFilter

This guide shows how to register a `BroadcastReceiver` from Go to listen for system events like WiFi state changes, battery updates, and Bluetooth discovery results.

## Creating a BroadcastReceiver via JNI Proxy

Since `BroadcastReceiver` is an abstract class (not an interface), the approach uses `java.lang.reflect.Proxy` with a wrapper interface or a direct subclass. The most practical method is using `env.NewProxy` with the `InvocationHandler` pattern.

### Pattern: WiFi State Change Listener

```go
import (
    "github.com/AndroidGoLab/jni"
    "github.com/AndroidGoLab/jni/net/wifi"
)

func listenForWifiChanges(vm *jni.VM, ctx *app.Context) error {
    mgr, err := wifi.NewManager(ctx)
    if err != nil {
        return err
    }
    defer mgr.Close()

    // Check current WiFi state
    enabled, _ := mgr.IsWifiEnabled()
    fmt.Printf("WiFi enabled: %v\n", enabled)

    // For real-time state changes, register via raw JNI:
    return vm.Do(func(env *jni.Env) error {
        // 1. Create an IntentFilter
        ifClass, _ := env.FindClass("android/content/IntentFilter")
        ifInit, _ := env.GetMethodID(ifClass, "<init>", "(Ljava/lang/String;)V")
        action, _ := env.NewStringUTF("android.net.wifi.WIFI_STATE_CHANGED")
        intentFilter, _ := env.NewObject(ifClass, ifInit, jni.ObjectValue(&action.Object))

        // 2. Create a BroadcastReceiver via proxy
        //    BroadcastReceiver is abstract, so we use a dynamic subclass.
        //    Alternatively, define a minimal Java helper class.
        //    The env.NewProxy approach works for interfaces only.

        // For BroadcastReceiver (abstract class), use registerCallback-style APIs
        // when available (Android 13+ has registerNetworkCallback, etc.)

        // See the "Modern Callback APIs" section below for the recommended approach.
        _ = intentFilter
        return nil
    })
}
```

## Modern Callback APIs (Recommended)

Android provides callback-based APIs that work better with JNI proxy:

### Network Connectivity Callback

```go
import "github.com/AndroidGoLab/jni/net"

mgr, _ := net.NewConnectivityManager(ctx)
defer mgr.Close()

// ConnectivityManager provides callback registration
// that doesn't require BroadcastReceiver
activeNet, _ := mgr.GetActiveNetwork()
caps, _ := mgr.GetNetworkCapabilities(activeNet)
```

### WiFi Scan Results (Post-Query Pattern)

```go
import "github.com/AndroidGoLab/jni/net/wifi"

mgr, _ := wifi.NewManager(ctx)
defer mgr.Close()

// Query current state directly instead of waiting for broadcasts
enabled, _ := mgr.IsWifiEnabled()
// Get scan results (populated by the OS)
scanResults, _ := mgr.GetScanResults()
```

## JNI Proxy for Interface-Based Callbacks

When the callback is a Java interface (not abstract class), `env.NewProxy` works directly:

```go
vm.Do(func(env *jni.Env) error {
    // Example: LocationListener (interface)
    listenerClass, _ := env.FindClass("android/location/LocationListener")

    proxy, cleanup, err := env.NewProxy(
        []*jni.Class{listenerClass},
        func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
            switch methodName {
            case "onLocationChanged":
                // Handle location update
            case "onProviderEnabled":
                // Handle provider enabled
            case "onProviderDisabled":
                // Handle provider disabled
            }
            return nil, nil
        },
    )
    if err != nil {
        return err
    }
    defer cleanup()

    // Register the proxy with a manager method
    // ...
    return nil
})
```

## HandlerThread for Callback Delivery

Callbacks need a `Looper` thread to be delivered. Create a `HandlerThread`:

```go
vm.Do(func(env *jni.Env) error {
    // Create HandlerThread
    htClass, _ := env.FindClass("android/os/HandlerThread")
    htInit, _ := env.GetMethodID(htClass, "<init>", "(Ljava/lang/String;)V")
    name, _ := env.NewStringUTF("CallbackThread")
    ht, _ := env.NewObject(htClass, htInit, jni.ObjectValue(&name.Object))
    handlerThread := env.NewGlobalRef(ht)

    // Start the thread
    startMid, _ := env.GetMethodID(htClass, "start", "()V")
    env.CallVoidMethod(handlerThread, startMid)

    // Get its Looper for registering callbacks
    getLooperMid, _ := env.GetMethodID(htClass, "getLooper", "()Landroid/os/Looper;")
    looper, _ := env.CallObjectMethod(handlerThread, getLooperMid)

    // Use looper when registering listeners (e.g., requestLocationUpdates)
    _ = looper

    // Later: quit the thread
    // quitMid, _ := env.GetMethodID(htClass, "quit", "()Z")
    // env.CallBooleanMethod(handlerThread, quitMid)
    // env.DeleteGlobalRef(handlerThread)

    return nil
})
```

## WiFi P2P (Wi-Fi Direct)

The `net/wifi/p2p` package wraps WiFi Direct functionality:

```go
import "github.com/AndroidGoLab/jni/net/wifi/p2p"

mgr, _ := p2p.NewWifiP2pManager(ctx)
defer mgr.Close()

// WiFi P2P uses callback interfaces that work with env.NewProxy
```

## Content Wrappers for BroadcastReceiver

The `content` package provides a generated `BroadcastReceiver` wrapper type:

```go
import "github.com/AndroidGoLab/jni/content"

// BroadcastReceiver wraps android.content.BroadcastReceiver
receiver := content.BroadcastReceiver{VM: vm, Obj: receiverObj}
receiver.AbortBroadcast()
result, _ := receiver.GetAbortBroadcast()
```

Note: instantiating a `BroadcastReceiver` from Go requires a Java subclass or dynamic proxy. The wrapper is for interacting with existing receiver instances passed via JNI.
