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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/net"
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

	mgr, err := net.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("net.NewManager: %v", err)
	}
	defer mgr.Close()

	// Print all transport type constants.
	// These identify the transport mechanism of a network.
	fmt.Fprintln(&output, "Transport types:")
	fmt.Fprintf(&output, "  TransportCellular  = %d\n", net.TransportCellular)
	fmt.Fprintf(&output, "  TransportWiFi      = %d\n", net.TransportWiFi)
	fmt.Fprintf(&output, "  TransportBluetooth = %d\n", net.TransportBluetooth)
	fmt.Fprintf(&output, "  TransportVPN       = %d\n", net.TransportVPN)

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

	fmt.Fprintln(&output, "ConnectivityManager ready")
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
