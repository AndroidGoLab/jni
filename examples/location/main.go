//go:build android

// Command location demonstrates the Location API provided by the generated
// location package. It is built as a c-shared library and packaged into an
// APK using the shared apk.mk infrastructure.
//
// The example first checks cached locations via GetLastKnownLocation. If no
// cached location is available, it requests a fresh GPS fix using
// requestLocationUpdates with a JNI proxy LocationListener and waits up to
// 15 seconds for a result.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/location"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	vm := jni.VMFromPtr(unsafe.Pointer(cvm))
	if err := run(vm); err != nil {
		fmt.Fprintf(&output, "ERROR: %v\n", err)
	}
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}

func run(vm *jni.VM) error {
	ctx, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(&output, "=== Provider constants ===")
	fmt.Fprintf(&output, "  GPS     = %q\n", location.GPS)
	fmt.Fprintf(&output, "  Network = %q\n", location.Network)
	fmt.Fprintf(&output, "  Passive = %q\n", location.Passive)

	mgr, err := location.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("location.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(&output, "\n=== Provider status ===")
	for _, provider := range []string{location.GPS, location.Network, location.Passive} {
		enabled, err := mgr.IsProviderEnabled(provider)
		if err != nil {
			fmt.Fprintf(&output, "  IsProviderEnabled(%s): %v\n", provider, err)
			continue
		}
		fmt.Fprintf(&output, "  %q enabled: %v\n", provider, enabled)
	}

	// Check all providers including "fused" (Google Play Services).
	providers := []string{location.GPS, location.Network, location.Passive, "fused"}

	fmt.Fprintln(&output, "\n=== Last known location ===")
	gotLocation := false
	for _, provider := range providers {
		enabled, _ := mgr.IsProviderEnabled(provider)
		if !enabled {
			continue
		}
		locObj, err := mgr.GetLastKnownLocation(provider)
		if err != nil {
			fmt.Fprintf(&output, "  GetLastKnownLocation(%s): %v\n", provider, err)
			continue
		}
		if locObj == nil || locObj.Ref() == 0 {
			fmt.Fprintf(&output, "  %s: no cached location\n", provider)
			continue
		}
		printLocation(vm, provider, locObj)
		gotLocation = true
	}

	// If no cached location was found, request a fresh GPS fix.
	if !gotLocation {
		fmt.Fprintln(&output, "\n=== Requesting fresh GPS fix (up to 30s) ===")
		loc, err := requestFreshLocation(vm, mgr)
		switch {
		case err != nil:
			fmt.Fprintf(&output, "  requestFreshLocation: %v\n", err)
		case loc != nil:
			fmt.Fprintf(&output, "  %s: lat=%.6f lon=%.6f alt=%.1f acc=%.1f speed=%.1f bearing=%.1f time=%d\n",
				loc.Provider, loc.Latitude, loc.Longitude,
				loc.Altitude, loc.Accuracy, loc.Speed, loc.Bearing, loc.Time)
			gotLocation = true
		default:
			fmt.Fprintln(&output, "  No location received within timeout. Try again outdoors.")
		}
	}

	fmt.Fprintln(&output, "\nLocation example completed successfully.")
	return nil
}

func printLocation(vm *jni.VM, provider string, locObj *jni.Object) {
	var loc *location.Location
	var err error
	vm.Do(func(env *jni.Env) error {
		loc, err = location.ExtractLocation(env, locObj)
		return err
	})
	if err != nil {
		fmt.Fprintf(&output, "  ExtractLocation(%s): %v\n", provider, err)
		return
	}
	fmt.Fprintf(&output, "  %s: lat=%.6f lon=%.6f alt=%.1f acc=%.1f speed=%.1f bearing=%.1f time=%d\n",
		loc.Provider, loc.Latitude, loc.Longitude,
		loc.Altitude, loc.Accuracy, loc.Speed, loc.Bearing, loc.Time)
}

// requestFreshLocation uses requestLocationUpdates with a JNI proxy
// LocationListener to obtain a fresh GPS fix. It creates a HandlerThread
// with its own Looper so callbacks can be delivered while the main thread
// waits. Waits up to 15 seconds.
func requestFreshLocation(vm *jni.VM, mgr *location.Manager) (*location.Location, error) {
	var mu sync.Mutex
	var result *location.Location
	done := make(chan struct{})

	var listenerGlobal *jni.Object
	var handlerThread *jni.Object
	var cleanup func()

	err := vm.Do(func(env *jni.Env) error {
		// Create a HandlerThread with its own Looper for callbacks.
		htClass, err := env.FindClass("android/os/HandlerThread")
		if err != nil {
			return fmt.Errorf("find HandlerThread: %w", err)
		}
		htInit, err := env.GetMethodID(htClass, "<init>", "(Ljava/lang/String;)V")
		if err != nil {
			return fmt.Errorf("get HandlerThread init: %w", err)
		}
		threadName, err := env.NewStringUTF("LocationHelper")
		if err != nil {
			return fmt.Errorf("new string: %w", err)
		}
		ht, err := env.NewObject(htClass, htInit, jni.ObjectValue(&threadName.Object))
		if err != nil {
			return fmt.Errorf("create HandlerThread: %w", err)
		}
		handlerThread = env.NewGlobalRef(ht)

		startMid, err := env.GetMethodID(htClass, "start", "()V")
		if err != nil {
			return fmt.Errorf("get start: %w", err)
		}
		if err := env.CallVoidMethod(handlerThread, startMid); err != nil {
			return fmt.Errorf("start HandlerThread: %w", err)
		}

		getLooperMid, err := env.GetMethodID(htClass, "getLooper", "()Landroid/os/Looper;")
		if err != nil {
			return fmt.Errorf("get getLooper: %w", err)
		}
		looper, err := env.CallObjectMethod(handlerThread, getLooperMid)
		if err != nil {
			return fmt.Errorf("get looper: %w", err)
		}

		// Find the LocationListener interface.
		listenerClass, err := env.FindClass("android/location/LocationListener")
		if err != nil {
			return fmt.Errorf("find LocationListener: %w", err)
		}

		// Find classes needed for type-checking in the proxy handler.
		listClass, err := env.FindClass("java/util/List")
		if err != nil {
			return fmt.Errorf("find List class: %w", err)
		}
		listGetMid, err := env.GetMethodID(listClass, "get", "(I)Ljava/lang/Object;")
		if err != nil {
			return fmt.Errorf("get List.get: %w", err)
		}

		// Create a proxy that implements LocationListener.
		proxy, proxyCleanup, err := env.NewProxy(
			[]*jni.Class{listenerClass},
			func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
				if methodName == "onLocationChanged" && len(args) > 0 {
					locObj := args[0]

					// Android API 31+ may deliver batched updates by calling
					// onLocationChanged(List<Location>) instead of
					// onLocationChanged(Location). Since java.lang.reflect.Proxy
					// dispatches both overloads through the same InvocationHandler,
					// we must check the actual runtime type of args[0].
					if env.IsInstanceOf(locObj, listClass) {
						elem, err := env.CallObjectMethod(locObj, listGetMid, jni.IntValue(0))
						if err != nil || elem == nil {
							return nil, nil
						}
						locObj = elem
					}

					loc, err := location.ExtractLocation(env, locObj)
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
			return fmt.Errorf("create LocationListener proxy: %w", err)
		}
		cleanup = proxyCleanup
		listenerGlobal = env.NewGlobalRef(proxy)

		// requestLocationUpdates(String, long, float, LocationListener, Looper)
		lmClass, err := env.FindClass("android/location/LocationManager")
		if err != nil {
			return fmt.Errorf("find LocationManager: %w", err)
		}
		reqMid, err := env.GetMethodID(lmClass, "requestLocationUpdates",
			"(Ljava/lang/String;JFLandroid/location/LocationListener;Landroid/os/Looper;)V")
		if err != nil {
			return fmt.Errorf("get requestLocationUpdates: %w", err)
		}
		providerStr, err := env.NewStringUTF(location.GPS)
		if err != nil {
			return fmt.Errorf("new string: %w", err)
		}
		return env.CallVoidMethod(mgr.Obj, reqMid,
			jni.ObjectValue(&providerStr.Object),
			jni.LongValue(0),
			jni.FloatValue(0),
			jni.ObjectValue(listenerGlobal),
			jni.ObjectValue(looper))
	})
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, err
	}

	// Wait for a location or timeout.
	select {
	case <-done:
	case <-time.After(30 * time.Second):
	}

	// Remove updates, quit handler thread, clean up.
	vm.Do(func(env *jni.Env) error {
		lmClass, err := env.FindClass("android/location/LocationManager")
		if err != nil {
			return err
		}
		removeMid, err := env.GetMethodID(lmClass, "removeUpdates",
			"(Landroid/location/LocationListener;)V")
		if err != nil {
			return err
		}
		env.CallVoidMethod(mgr.Obj, removeMid, jni.ObjectValue(listenerGlobal))
		env.DeleteGlobalRef(listenerGlobal)

		// Quit the HandlerThread.
		htClass, err := env.FindClass("android/os/HandlerThread")
		if err != nil {
			return err
		}
		quitMid, err := env.GetMethodID(htClass, "quit", "()Z")
		if err != nil {
			return err
		}
		env.CallBooleanMethod(handlerThread, quitMid)
		env.DeleteGlobalRef(handlerThread)
		return nil
	})
	if cleanup != nil {
		cleanup()
	}

	mu.Lock()
	defer mu.Unlock()
	return result, nil
}

// getAppContext obtains an Android Context via ActivityThread.currentApplication().
func getAppContext(vm *jni.VM) (*app.Context, error) {
	var ctx app.Context
	ctx.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}

		atClass, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread: %w", err)
		}

		curAppMid, err := env.GetStaticMethodID(atClass, "currentApplication", "()Landroid/app/Application;")
		if err != nil {
			return fmt.Errorf("get currentApplication: %w", err)
		}
		appObj, err := env.CallStaticObjectMethod(atClass, curAppMid)
		if err != nil {
			return fmt.Errorf("call currentApplication: %w", err)
		}
		if appObj == nil || appObj.Ref() == 0 {
			return fmt.Errorf("currentApplication returned null")
		}

		ctx.Obj = env.NewGlobalRef(appObj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}
