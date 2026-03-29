//go:build android

// Command net_connectivity_monitor queries the ConnectivityManager for active
// network information, metered status, background restrictions, and multipath
// preference using typed wrappers only.
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
	fmt.Fprintln(output, "=== Connectivity Monitor ===")

	// Check metered status.
	metered, err := mgr.IsActiveNetworkMetered()
	if err != nil {
		fmt.Fprintf(output, "IsActiveNetworkMetered: %v\n", err)
	} else {
		fmt.Fprintf(output, "Active network metered: %v\n", metered)
	}

	// Check default network active.
	active, err := mgr.IsDefaultNetworkActive()
	if err != nil {
		fmt.Fprintf(output, "IsDefaultNetworkActive: %v\n", err)
	} else {
		fmt.Fprintf(output, "Default network active: %v\n", active)
	}

	// Get restrict background status.
	rbStatus, err := mgr.GetRestrictBackgroundStatus()
	if err != nil {
		fmt.Fprintf(output, "GetRestrictBackgroundStatus: %v\n", err)
	} else {
		statusName := "unknown"
		switch int(rbStatus) {
		case net.RestrictBackgroundStatusDisabled:
			statusName = "disabled"
		case net.RestrictBackgroundStatusEnabled:
			statusName = "enabled"
		case net.RestrictBackgroundStatusWhitelisted:
			statusName = "whitelisted"
		}
		fmt.Fprintf(output, "Restrict background: %s (%d)\n", statusName, rbStatus)
	}

	// Get active network.
	network, err := mgr.GetActiveNetwork()
	if err != nil {
		fmt.Fprintf(output, "GetActiveNetwork: %v\n", err)
	} else if network == nil || network.Ref() == 0 {
		fmt.Fprintln(output, "No active network")
	} else {
		fmt.Fprintln(output, "\nActive network present")

		// Get link properties via typed wrapper.
		linkPropsObj, err := mgr.GetLinkProperties(network)
		if err != nil {
			fmt.Fprintf(output, "GetLinkProperties: %v\n", err)
		} else if linkPropsObj != nil && linkPropsObj.Ref() != 0 {
			lp := &net.LinkProperties{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(linkPropsObj))}

			ifName, err := lp.GetInterfaceName()
			if err == nil {
				fmt.Fprintf(output, "\nInterface: %s\n", ifName)
			}
			domains, err := lp.GetDomains()
			if err == nil && domains != "" {
				fmt.Fprintf(output, "Domains: %s\n", domains)
			}
			mtu, err := lp.GetMtu()
			if err == nil {
				fmt.Fprintf(output, "MTU: %d\n", mtu)
			}
			privateDns, err := lp.IsPrivateDnsActive()
			if err == nil {
				fmt.Fprintf(output, "Private DNS active: %v\n", privateDns)
			}
			privateDnsServer, err := lp.GetPrivateDnsServerName()
			if err == nil && privateDnsServer != "" {
				fmt.Fprintf(output, "Private DNS server: %s\n", privateDnsServer)
			}
			wol, err := lp.IsWakeOnLanSupported()
			if err == nil {
				fmt.Fprintf(output, "Wake-on-LAN: %v\n", wol)
			}
			lpStr, err := lp.ToString()
			if err == nil {
				fmt.Fprintf(output, "\nLink properties (full):\n  %s\n", lpStr)
			}

			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(linkPropsObj)
				return nil
			})
		}

		// Get multipath preference.
		multipathPref, err := mgr.GetMultipathPreference(network)
		if err != nil {
			fmt.Fprintf(output, "GetMultipathPreference: %v\n", err)
		} else {
			fmt.Fprintf(output, "\nMultipath preference: %d\n", multipathPref)
		}

		// Clean up.
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(network)
			return nil
		})
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Connectivity monitor complete.")
	return nil
}
