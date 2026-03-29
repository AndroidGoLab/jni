//go:build android

// Command net_captive_portal checks the active network for captive portal
// detection using ConnectivityManager typed wrappers. Since NetworkCapabilities
// is not exported, this example reports what the ConnectivityManager API can
// show directly: active network presence, metered status, and link properties.
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
	"github.com/AndroidGoLab/jni/net"
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

	mgr, err := net.NewConnectivityManager(ctx)
	if err != nil {
		return fmt.Errorf("ConnectivityManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "=== Captive Portal Detection ===")

	// Get the active network.
	network, err := mgr.GetActiveNetwork()
	if err != nil {
		return fmt.Errorf("GetActiveNetwork: %w", err)
	}
	if network == nil || network.Ref() == 0 {
		fmt.Fprintln(output, "No active network -- cannot check captive portal")
		return nil
	}
	defer func() {
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(network)
			return nil
		})
	}()

	fmt.Fprintln(output, "Active network found")

	// Report metered status (metered networks are more likely behind captive portals).
	metered, err := mgr.IsActiveNetworkMetered()
	if err != nil {
		fmt.Fprintf(output, "IsActiveNetworkMetered: %v\n", err)
	} else {
		fmt.Fprintf(output, "Active network metered: %v\n", metered)
	}

	// Report default network active status.
	defaultActive, err := mgr.IsDefaultNetworkActive()
	if err != nil {
		fmt.Fprintf(output, "IsDefaultNetworkActive: %v\n", err)
	} else {
		fmt.Fprintf(output, "Default network active: %v\n", defaultActive)
	}

	// Get link properties for the active network.
	linkPropsObj, err := mgr.GetLinkProperties(network)
	if err != nil {
		fmt.Fprintf(output, "GetLinkProperties: %v\n", err)
	} else if linkPropsObj != nil && linkPropsObj.Ref() != 0 {
		lp := &net.LinkProperties{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(linkPropsObj))}

		ifName, err := lp.GetInterfaceName()
		if err == nil {
			fmt.Fprintf(output, "\nInterface: %s\n", ifName)
		}
		mtu, err := lp.GetMtu()
		if err == nil {
			fmt.Fprintf(output, "MTU: %d\n", mtu)
		}
		privateDns, err := lp.IsPrivateDnsActive()
		if err == nil {
			fmt.Fprintf(output, "Private DNS active: %v\n", privateDns)
		}
		lpStr, err := lp.ToString()
		if err == nil {
			fmt.Fprintf(output, "\nLink properties:\n  %s\n", lpStr)
		}

		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(linkPropsObj)
			return nil
		})
	}

	// Note about captive portal detection.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Note: Full captive portal detection requires")
	fmt.Fprintln(output, "NetworkCapabilities.hasCapability(NET_CAPABILITY_CAPTIVE_PORTAL).")
	fmt.Fprintln(output, "The NetworkCapabilities typed wrapper is package-private;")
	fmt.Fprintln(output, "use ConnectivityManager.NetworkCallback for runtime detection.")

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Captive portal check complete.")
	return nil
}
