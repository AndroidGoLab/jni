# Location API

The `location` package wraps `android.location.LocationManager` for GPS and network location.

## Query Providers and Cached Location

```go
import "github.com/AndroidGoLab/jni/location"

mgr, err := location.NewManager(ctx)
if err != nil {
    return fmt.Errorf("location.NewManager: %w", err)
}
defer mgr.Close()

// Check which providers are available
gpsEnabled, _ := mgr.IsProviderEnabled(location.GpsProvider)      // "gps"
netEnabled, _ := mgr.IsProviderEnabled(location.NetworkProvider)   // "network"

// Read the last cached location (may be nil if no recent fix)
locObj, err := mgr.GetLastKnownLocation(location.GpsProvider)
if locObj != nil {
    // Extract lat/lon from the Java Location object
    var loc *location.ExtractedLocation
    vm.Do(func(env *jni.Env) error {
        loc, err = location.ExtractLocation(env, locObj)
        return err
    })
    fmt.Printf("lat=%.6f lon=%.6f\n", loc.Latitude, loc.Longitude)
}
```

## Provider Constants

```go
location.GpsProvider     // "gps"
location.NetworkProvider // "network"
location.PassiveProvider // "passive"
```

## Requesting Fresh Location Updates

When no cached location exists, request a fresh GPS fix. This requires creating a `LocationListener` proxy via JNI and a `HandlerThread` for callback delivery:

```go
import (
    "sync"
    "time"

    "github.com/AndroidGoLab/jni"
    "github.com/AndroidGoLab/jni/location"
)

func requestFreshLocation(vm *jni.VM, mgr *location.Manager) (*location.ExtractedLocation, error) {
    var mu sync.Mutex
    var result *location.ExtractedLocation
    done := make(chan struct{})

    var listenerGlobal *jni.Object
    var handlerThread *jni.Object
    var cleanup func()

    err := vm.Do(func(env *jni.Env) error {
        // 1. Create a HandlerThread with its own Looper for callbacks
        htClass, _ := env.FindClass("android/os/HandlerThread")
        htInit, _ := env.GetMethodID(htClass, "<init>", "(Ljava/lang/String;)V")
        threadName, _ := env.NewStringUTF("LocationHelper")
        ht, _ := env.NewObject(htClass, htInit, jni.ObjectValue(&threadName.Object))
        handlerThread = env.NewGlobalRef(ht)

        startMid, _ := env.GetMethodID(htClass, "start", "()V")
        env.CallVoidMethod(handlerThread, startMid)

        getLooperMid, _ := env.GetMethodID(htClass, "getLooper", "()Landroid/os/Looper;")
        looper, _ := env.CallObjectMethod(handlerThread, getLooperMid)

        // 2. Create a JNI proxy implementing LocationListener
        listenerClass, _ := env.FindClass("android/location/LocationListener")
        proxy, proxyCleanup, err := env.NewProxy(
            []*jni.Class{listenerClass},
            func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
                if methodName == "onLocationChanged" && len(args) > 0 {
                    loc, err := location.ExtractLocation(env, args[0])
                    if err == nil && loc != nil {
                        mu.Lock()
                        if result == nil {
                            result = loc
                            close(done)
                        }
                        mu.Unlock()
                    }
                }
                return nil, nil
            },
        )
        if err != nil {
            return err
        }
        cleanup = proxyCleanup
        listenerGlobal = env.NewGlobalRef(proxy)

        // 3. Call requestLocationUpdates(String, long, float, LocationListener, Looper)
        lmClass, _ := env.FindClass("android/location/LocationManager")
        reqMid, _ := env.GetMethodID(lmClass, "requestLocationUpdates",
            "(Ljava/lang/String;JFLandroid/location/LocationListener;Landroid/os/Looper;)V")
        providerStr, _ := env.NewStringUTF(location.GpsProvider)
        return env.CallVoidMethod(mgr.Obj, reqMid,
            jni.ObjectValue(&providerStr.Object),
            jni.LongValue(0),
            jni.FloatValue(0),
            jni.ObjectValue(listenerGlobal),
            jni.ObjectValue(looper))
    })
    if err != nil {
        if cleanup != nil { cleanup() }
        return nil, err
    }

    // 4. Wait for a location or timeout
    select {
    case <-done:
    case <-time.After(30 * time.Second):
    }

    // 5. Remove updates and clean up
    vm.Do(func(env *jni.Env) error {
        lmClass, _ := env.FindClass("android/location/LocationManager")
        removeMid, _ := env.GetMethodID(lmClass, "removeUpdates",
            "(Landroid/location/LocationListener;)V")
        env.CallVoidMethod(mgr.Obj, removeMid, jni.ObjectValue(listenerGlobal))
        env.DeleteGlobalRef(listenerGlobal)

        htClass, _ := env.FindClass("android/os/HandlerThread")
        quitMid, _ := env.GetMethodID(htClass, "quit", "()Z")
        env.CallBooleanMethod(handlerThread, quitMid)
        env.DeleteGlobalRef(handlerThread)
        return nil
    })
    if cleanup != nil { cleanup() }

    mu.Lock()
    defer mu.Unlock()
    return result, nil
}
```

## JNI Proxy Pattern (env.NewProxy)

`env.NewProxy` creates a `java.lang.reflect.Proxy` that implements one or more Java interfaces and dispatches method calls to a Go callback:

```go
proxy, cleanup, err := env.NewProxy(
    []*jni.Class{interfaceClass},
    func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
        switch methodName {
        case "onLocationChanged":
            // Handle the callback
        case "onProviderEnabled":
            // Handle provider enabled
        }
        return nil, nil
    },
)
defer cleanup() // Release the Go callback prevent leaks
```

This is how the library implements Java callbacks (listeners, BroadcastReceivers) entirely from Go.

## Required Permissions

```xml
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
<uses-permission android:name="android.permission.ACCESS_COARSE_LOCATION" />
```
