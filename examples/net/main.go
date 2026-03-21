//go:build android

// Command net demonstrates the ConnectivityManager JNI bindings. It is built
// as a c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the ConnectivityManager system service and
// prints the transport type constants. The package wraps
// android.net.ConnectivityManager and provides the networkCallback
// struct for receiving connectivity change notifications.
package main

/*
#include <android/native_activity.h>
extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static void _setCallbacks(ANativeActivity* a) { a->callbacks->onResume = _onResume; a->callbacks->onNativeWindowCreated = _onWindowCreated; }
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/net"
)

func main() {}

func init() { exampleui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	exampleui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	exampleui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	exampleui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := net.NewConnectivityManager(ctx)
	if err != nil {
		return fmt.Errorf("net.NewConnectivityManager: %v", err)
	}
	defer mgr.Close()

	// Print all transport type constants.
	// These identify the transport mechanism of a network.
	fmt.Fprintln(output, "Transport types:")
	fmt.Fprintf(output, "  TransportCellular  = %d\n", net.TransportCellular)
	fmt.Fprintf(output, "  TransportWifi      = %d\n", net.TransportWifi)
	fmt.Fprintf(output, "  TransportBluetooth = %d\n", net.TransportBluetooth)
	fmt.Fprintf(output, "  TransportVpn       = %d\n", net.TransportVpn)

	// Package-internal Manager methods:
	//   getActiveNetworkRaw()                    - get the active Network object
	//   getNetworkCapabilitiesRaw(network)       - get capabilities of a network
	//   registerDefaultNetworkCallbackRaw(cb)    - register for network changes
	//   unregisterNetworkCallbackRaw(cb)         - unregister callback
	//
	// The networkCapabilities wrapper provides:
	//   hasTransport(transport int32) bool - check if transport type is present
	//   getLinkDown() int32               - downstream bandwidth in Kbps
	//   getLinkUp() int32                 - upstream bandwidth in Kbps
	//
	// The networkCallback struct enables Go code to receive connectivity
	// change notifications via a Java proxy:
	//   OnAvailable(network *jni.Object)
	//     Called when a network becomes available.
	//   OnLost(network *jni.Object)
	//     Called when a network is lost.
	//   OnCapabilitiesChanged(network *jni.Object, caps *jni.Object)
	//     Called when network capabilities change.

	fmt.Fprintln(output, "ConnectivityManager ready")
	return nil
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
