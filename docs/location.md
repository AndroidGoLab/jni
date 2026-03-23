# Location API

The `location` package wraps `android.location.LocationManager` for GPS and network location.

## Query Providers and Cached Location

```go
import (
    "fmt"

    "github.com/AndroidGoLab/jni/location"
)

mgr, err := location.NewManager(ctx)
if err != nil {
    return fmt.Errorf("location.NewManager: %w", err)
}
defer mgr.Close()

// Check which providers are available (typed wrapper takes string)
gpsEnabled, _ := mgr.IsProviderEnabled(location.GpsProvider)      // "gps"
netEnabled, _ := mgr.IsProviderEnabled(location.NetworkProvider)   // "network"

// Read the last cached location (may be nil if no recent fix)
locObj, err := mgr.GetLastKnownLocation(location.GpsProvider)
if locObj != nil {
    // Wrap the returned JNI object in the typed Location wrapper.
    // GetLastKnownLocation already returns a GlobalRef, so use it directly.
    loc := &location.Location{VM: mgr.VM, Obj: locObj}
    lat, _ := loc.GetLatitude()
    lon, _ := loc.GetLongitude()
    fmt.Printf("lat=%.6f lon=%.6f\n", lat, lon)
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
    goos "github.com/AndroidGoLab/jni/os"
)

func requestFreshLocation(vm *jni.VM, mgr *location.Manager) (*location.Location, error) {
    var mu sync.Mutex
    var result *location.Location
    done := make(chan struct{})

    var listenerGlobal *jni.Object
    var cleanup func()

    // 1. Create and start a HandlerThread for callback delivery.
    ht, err := goos.NewHandlerThread(vm, "LocationHelper")
    if err != nil {
        return nil, fmt.Errorf("new HandlerThread: %w", err)
    }
    looper, err := ht.GetLooper()
    if err != nil {
        ht.Close()
        return nil, fmt.Errorf("get looper: %w", err)
    }

    err = vm.Do(func(env *jni.Env) error {

        // 2. Create a JNI proxy implementing LocationListener.
        //    Raw JNI: LocationListener is an interface; env.NewProxy is the
        //    only way to implement a Java interface callback from Go.
        listenerClass, _ := env.FindClass("android/location/LocationListener")
        proxy, proxyCleanup, err := env.NewProxy(
            []*jni.Class{listenerClass},
            func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
                if methodName == "onLocationChanged" && len(args) > 0 {
                    // Wrap the callback's Location argument in the typed wrapper
                    locGlobal := env.NewGlobalRef(args[0])
                    loc := &location.Location{VM: vm, Obj: locGlobal}
                    mu.Lock()
                    if result == nil {
                        result = loc
                        close(done)
                    }
                    mu.Unlock()
                }
                return nil, nil
            },
        )
        if err != nil {
            return err
        }
        cleanup = proxyCleanup
        listenerGlobal = env.NewGlobalRef(proxy)
        env.DeleteLocalRef(proxy)

        // 3. Use typed RequestLocationUpdates5_4:
        //    requestLocationUpdates(String, long, float, LocationListener, Looper)
        return mgr.RequestLocationUpdates5_4(
            location.GpsProvider, // provider
            0,                    // minTimeMs
            0,                    // minDistanceMeters
            listenerGlobal,       // listener
            looper,               // looper
        )
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

    // 5. Remove updates and clean up using typed wrappers
    mgr.RemoveUpdates1(listenerGlobal)
    vm.Do(func(env *jni.Env) error {
        env.DeleteGlobalRef(listenerGlobal)
        return nil
    })
    ht.Close()
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
