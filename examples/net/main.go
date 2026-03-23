//go:build android

// Command net demonstrates the ConnectivityManager JNI bindings. It obtains
// the active network, checks whether it is metered, queries network
// capabilities, and prints live connectivity information.
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
	fmt.Fprintln(output, "=== ConnectivityManager ===")

	// Check if the active network is metered.
	metered, err := mgr.IsActiveNetworkMetered()
	if err != nil {
		fmt.Fprintf(output, "IsActiveNetworkMetered: %v\n", err)
	} else {
		fmt.Fprintf(output, "Active network metered: %v\n", metered)
	}

	// Check if the default network is active.
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

	// Get the active network object.
	network, err := mgr.GetActiveNetwork()
	if err != nil {
		fmt.Fprintf(output, "GetActiveNetwork: %v\n", err)
	} else if network == nil || network.Ref() == 0 {
		fmt.Fprintln(output, "No active network")
	} else {
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "Active network present")

		// Call toString() on the network object.
		vm.Do(func(env *jni.Env) error {
			cls := env.GetObjectClass(network)
			mid, err := env.GetMethodID(cls, "toString", "()Ljava/lang/String;")
			if err != nil {
				return nil
			}
			strObj, err := env.CallObjectMethod(network, mid)
			if err != nil {
				return nil
			}
			s := env.GoString((*jni.String)(unsafe.Pointer(strObj)))
			fmt.Fprintf(output, "  Network: %s\n", s)
			return nil
		})

		// Get link properties.
		linkProps, err := mgr.GetLinkProperties(network)
		if err != nil {
			fmt.Fprintf(output, "GetLinkProperties: %v\n", err)
		} else if linkProps != nil && linkProps.Ref() != 0 {
			vm.Do(func(env *jni.Env) error {
				cls := env.GetObjectClass(linkProps)
				mid, err := env.GetMethodID(cls, "toString", "()Ljava/lang/String;")
				if err != nil {
					return nil
				}
				strObj, err := env.CallObjectMethod(linkProps, mid)
				if err != nil {
					return nil
				}
				s := env.GoString((*jni.String)(unsafe.Pointer(strObj)))
				fmt.Fprintf(output, "Link properties:\n  %s\n", s)
				env.DeleteGlobalRef(linkProps)
				return nil
			})
		}

		// Get multipath preference.
		multipathPref, err := mgr.GetMultipathPreference(network)
		if err != nil {
			fmt.Fprintf(output, "GetMultipathPreference: %v\n", err)
		} else {
			fmt.Fprintf(output, "Multipath preference: %d\n", multipathPref)
		}

		// Get network capabilities.
		caps, err := mgr.GetNetworkCapabilities(network)
		if err != nil {
			fmt.Fprintf(output, "GetNetworkCapabilities: %v\n", err)
		} else if caps != nil && caps.Ref() != 0 {
			vm.Do(func(env *jni.Env) error {
				cls := env.GetObjectClass(caps)
				mid, err := env.GetMethodID(cls, "toString", "()Ljava/lang/String;")
				if err != nil {
					return nil
				}
				strObj, err := env.CallObjectMethod(caps, mid)
				if err != nil {
					return nil
				}
				s := env.GoString((*jni.String)(unsafe.Pointer(strObj)))
				fmt.Fprintf(output, "Capabilities:\n  %s\n", s)

				// Check individual transports.
				hasMid, err := env.GetMethodID(cls, "hasTransport", "(I)Z")
				if err != nil {
					return nil
				}
				transports := []struct {
					name string
					id   int
				}{
					{"Cellular", net.TransportCellular},
					{"WiFi", net.TransportWifi},
					{"Bluetooth", net.TransportBluetooth},
					{"Ethernet", net.TransportEthernet},
					{"VPN", net.TransportVpn},
					{"USB", net.TransportUsb},
				}
				fmt.Fprintln(output, "Transports:")
				for _, t := range transports {
					has, err := env.CallBooleanMethod(caps, hasMid, jni.IntValue(int32(t.id)))
					if err != nil {
						continue
					}
					fmt.Fprintf(output, "  %-10s: %v\n", t.name, has != 0)
				}

				// Check key capabilities.
				hasCapMid, err := env.GetMethodID(cls, "hasCapability", "(I)Z")
				if err == nil {
					capChecks := []struct {
						name string
						id   int
					}{
						{"Internet", net.NetCapabilityInternet},
						{"Validated", net.NetCapabilityValidated},
						{"NotMetered", net.NetCapabilityNotMetered},
						{"NotVPN", net.NetCapabilityNotVpn},
						{"NotRoaming", net.NetCapabilityNotRoaming},
						{"Trusted", net.NetCapabilityTrusted},
					}
					fmt.Fprintln(output, "Capabilities:")
					for _, c := range capChecks {
						has, err := env.CallBooleanMethod(caps, hasCapMid, jni.IntValue(int32(c.id)))
						if err != nil {
							continue
						}
						fmt.Fprintf(output, "  %-12s: %v\n", c.name, has != 0)
					}
				}

				// Get bandwidth info.
				downMid, err := env.GetMethodID(cls, "getLinkDownstreamBandwidthKbps", "()I")
				if err == nil {
					down, err := env.CallIntMethod(caps, downMid)
					if err == nil {
						fmt.Fprintf(output, "Downstream: %d Kbps\n", down)
					}
				}
				upMid, err := env.GetMethodID(cls, "getLinkUpstreamBandwidthKbps", "()I")
				if err == nil {
					up, err := env.CallIntMethod(caps, upMid)
					if err == nil {
						fmt.Fprintf(output, "Upstream: %d Kbps\n", up)
					}
				}

				// Get signal strength.
				sigMid, err := env.GetMethodID(cls, "getSignalStrength", "()I")
				if err == nil {
					sig, err := env.CallIntMethod(caps, sigMid)
					if err == nil {
						fmt.Fprintf(output, "Signal strength: %d\n", sig)
					}
				}

				env.DeleteGlobalRef(caps)
				return nil
			})
		}

		// Clean up network global ref.
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(network)
			return nil
		})
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Network example complete.")
	return nil
}
