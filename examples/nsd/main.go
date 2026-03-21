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
	"github.com/AndroidGoLab/jni/net/nsd"
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
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- Constants ---
	fmt.Fprintf(output, "ProtocolDnsSd = %d\n", nsd.ProtocolDnsSd)

	// --- NewManager ---
	mgr, err := nsd.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("nsd.NewManager: %w", err)
	}
	defer mgr.Close()

	// --- NsdServiceInfo ---
	// NewnsdServiceInfo creates a Java NsdServiceInfo object used to
	// describe a network service for registration and discovery.
	// Despite the awkward name, this function is exported (capital N).
	// The nsd.ServiceInfo type wraps android.net.nsd.NsdServiceInfo.
	// ServiceInfo objects are obtained via discovery/resolve callbacks.
	var svcInfo nsd.ServiceInfo
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

	fmt.Fprintln(output, "NSD example complete.")
	return nil
}
