//go:build android

// Command nsd demonstrates using the Network Service Discovery (NSD) API. It
// is built as a c-shared library and packaged into an APK using the shared
// apk.mk infrastructure.
//
// This example obtains the NsdManager system service and shows the
// ProtocolDNSSD constant. The Manager provides methods for discovering,
// resolving, registering, and unregistering network services. Three
// callback types handle discovery, resolution, and registration events.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/app"
	"github.com/xaionaro-go/jni/net/nsd"
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

	// --- Constants ---
	fmt.Fprintf(&output, "ProtocolDNSSD = %d\n", nsd.ProtocolDNSSD)

	// --- NewManager ---
	mgr, err := nsd.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("nsd.NewManager: %v", err)
	}
	defer mgr.Close()

	// --- NsdServiceInfo ---
	// NewnsdServiceInfo creates a Java NsdServiceInfo object used to
	// describe a network service for registration and discovery.
	// Despite the awkward name, this function is exported (capital N).
	svcInfo, err := nsd.NewnsdServiceInfo(mgr.VM)
	if err != nil {
		return fmt.Errorf("nsd.NewnsdServiceInfo: %v", err)
	}
	_ = svcInfo

	// The nsdServiceInfo type provides these unexported methods
	// (called from within the nsd package or higher-level wrappers):
	//
	//   .setServiceName(name string)
	//   .setServiceType(serviceType string)
	//   .setPort(port int32)
	//   .setAttribute(key, value string)
	//   .getServiceName() string
	//   .getServiceType() string
	//   .getHost() *jni.Object
	//   .getPort() int32

	// --- Manager methods (all unexported, called via wrappers) ---
	//
	// Discovery:
	//   mgr.discoverServicesRaw(serviceType, protocolType, listener)
	//   mgr.stopServiceDiscoveryRaw(listener)
	//
	// Resolution:
	//   mgr.resolveServiceRaw(serviceInfo, listener)
	//
	// Registration:
	//   mgr.registerServiceRaw(serviceInfo, protocolType, listener)
	//   mgr.unregisterServiceRaw(listener)

	// --- Callbacks ---
	// Three callback types handle NSD events. Each is registered via
	// a register* function that creates a Java proxy:
	//
	// discoveryListener{
	//   OnServiceFound         func(*jni.Object)
	//   OnServiceLost          func(*jni.Object)
	//   OnDiscoveryStarted     func(string)
	//   OnDiscoveryStopped     func(string)
	//   OnStartDiscoveryFailed func(string, int32)
	//   OnStopDiscoveryFailed  func(string, int32)
	// }
	//
	// resolveListener{
	//   OnServiceResolved func(*jni.Object)
	//   OnResolveFailed   func(*jni.Object, int32)
	// }
	//
	// registrationListener{
	//   OnServiceRegistered    func(*jni.Object)
	//   OnRegistrationFailed   func(*jni.Object, int32)
	//   OnServiceUnregistered  func(*jni.Object)
	//   OnUnregistrationFailed func(*jni.Object, int32)
	// }

	fmt.Fprintln(&output, "NSD example complete.")
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
