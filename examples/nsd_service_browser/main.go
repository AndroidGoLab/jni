//go:build android

// Command nsd_service_browser demonstrates NSD (Network Service Discovery).
// It creates an NsdManager and shows available NSD API surface including
// the ServiceInfo type and protocol constants.
package main

/*
#include <android/native_activity.h>
extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static void _setCallbacks(ANativeActivity* a) { a->callbacks->onResume = _onResume; a->callbacks->onNativeWindowCreated = _onWindowCreated; }
static uintptr_t _getVM(ANativeActivity* a) { return (uintptr_t)a->vm; }
static uintptr_t _getClazz(ANativeActivity* a) { return (uintptr_t)a->clazz; }
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/net/nsd"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(C._getVM(activity))),
		jni.ObjectFromUintptr(uintptr(C._getClazz(activity))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(C._getClazz(activity))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== NSD Service Browser ===")

	// Obtain the NsdManager system service.
	nsdMgr, err := nsd.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("nsd.NewManager: %w", err)
	}
	defer nsdMgr.Close()
	fmt.Fprintln(output, "NsdManager obtained")

	// Show NSD constants.
	fmt.Fprintf(output, "Protocol DNS-SD: %d\n", nsd.ProtocolDnsSd)
	fmt.Fprintf(output, "NSD state enabled: %d\n", nsd.NsdStateEnabled)
	fmt.Fprintf(output, "NSD state disabled: %d\n", nsd.NsdStateDisabled)

	// Note: DiscoverServices requires a DiscoveryListener callback object.
	// Since callback proxy support is needed, we demonstrate the available
	// API surface without starting actual discovery.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "NsdManager methods available:")
	fmt.Fprintln(output, "  DiscoverServices3      (NsdServiceInfo, ProtocolType, DiscoveryListener)")
	fmt.Fprintln(output, "  DiscoverServices3_3    (serviceType string, protocolType int, listener)")
	fmt.Fprintln(output, "  StopServiceDiscovery   (listener)")
	fmt.Fprintln(output, "  RegisterService3       (NsdServiceInfo, protocolType, listener)")
	fmt.Fprintln(output, "  ResolveService2        (NsdServiceInfo, listener)")
	fmt.Fprintln(output, "  UnregisterService      (listener)")

	// Show NsdServiceInfo type is available.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "NsdServiceInfo methods available:")
	fmt.Fprintln(output, "  GetServiceName, GetServiceType, GetPort")
	fmt.Fprintln(output, "  GetHost, GetHostAddresses, GetHostname")
	fmt.Fprintln(output, "  SetServiceName, SetServiceType, SetPort")
	fmt.Fprintln(output, "  SetAttribute, RemoveAttribute, ToString")

	// TODO: Start actual _http._tcp discovery once callback proxy
	// support is wired up for NSD discovery listeners.

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "NSD service browser example complete.")
	return nil
}
