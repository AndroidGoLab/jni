//go:build android

// Command wifi_site_survey queries WiFi state, connection info, and scan
// results using the wifi package typed wrappers.
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
	"github.com/AndroidGoLab/jni/net/wifi"
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

	mgr, err := wifi.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("wifi.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== WiFi Site Survey ===")

	// Check WiFi state.
	enabled, err := mgr.IsWifiEnabled()
	if err != nil {
		fmt.Fprintf(output, "IsWifiEnabled: %v\n", err)
	} else {
		fmt.Fprintf(output, "WiFi enabled: %v\n", enabled)
	}

	state, err := mgr.GetWifiState()
	if err != nil {
		fmt.Fprintf(output, "GetWifiState: %v\n", err)
	} else {
		stateNames := map[int32]string{
			0: "DISABLING", 1: "DISABLED", 2: "ENABLING", 3: "ENABLED", 4: "UNKNOWN",
		}
		name := stateNames[state]
		if name == "" {
			name = fmt.Sprintf("STATE_%d", state)
		}
		fmt.Fprintf(output, "WiFi state: %s (%d)\n", name, state)
	}

	// Trigger a scan (best-effort; results may come from cache).
	started, err := mgr.StartScan()
	if err != nil {
		fmt.Fprintf(output, "StartScan error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Scan started: %v\n", started)
	}

	// Show current connection info using typed wrapper.
	connInfoObj, err := mgr.GetConnectionInfo()
	if err != nil {
		fmt.Fprintf(output, "\nGetConnectionInfo: %v\n", err)
	} else if connInfoObj != nil && connInfoObj.Ref() != 0 {
		info := &wifi.Info{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(connInfoObj))}

		fmt.Fprintln(output, "\nCurrent connection:")

		ssid, err := info.GetSSID()
		if err == nil {
			fmt.Fprintf(output, "  SSID: %s\n", ssid)
		}
		bssid, err := info.GetBSSID()
		if err == nil {
			fmt.Fprintf(output, "  BSSID: %s\n", bssid)
		}
		rssi, err := info.GetRssi()
		if err == nil {
			fmt.Fprintf(output, "  RSSI: %d dBm\n", rssi)
		}
		freq, err := info.GetFrequency()
		if err == nil {
			fmt.Fprintf(output, "  Frequency: %d MHz\n", freq)
		}
		linkSpeed, err := info.GetLinkSpeed()
		if err == nil {
			fmt.Fprintf(output, "  Link speed: %d Mbps\n", linkSpeed)
		}
		netId, err := info.GetNetworkId()
		if err == nil {
			fmt.Fprintf(output, "  Network ID: %d\n", netId)
		}
		macAddr, err := info.GetMacAddress()
		if err == nil {
			fmt.Fprintf(output, "  MAC: %s\n", macAddr)
		}
		hiddenSSID, err := info.GetHiddenSSID()
		if err == nil {
			fmt.Fprintf(output, "  Hidden SSID: %v\n", hiddenSSID)
		}
		rxSpeed, err := info.GetRxLinkSpeedMbps()
		if err == nil {
			fmt.Fprintf(output, "  Rx link speed: %d Mbps\n", rxSpeed)
		}
		txSpeed, err := info.GetTxLinkSpeedMbps()
		if err == nil {
			fmt.Fprintf(output, "  Tx link speed: %d Mbps\n", txSpeed)
		}
		wifiStd, err := info.GetWifiStandard()
		if err == nil {
			stdNames := map[int32]string{
				0: "UNKNOWN", 1: "LEGACY", 4: "11n", 5: "11ac", 6: "11ax", 7: "11ad", 8: "11be",
			}
			stdName := stdNames[wifiStd]
			if stdName == "" {
				stdName = fmt.Sprintf("STANDARD_%d", wifiStd)
			}
			fmt.Fprintf(output, "  WiFi standard: %s (%d)\n", stdName, wifiStd)
		}

		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(connInfoObj)
			return nil
		})
	}

	// Check scan-related capabilities.
	fmt.Fprintln(output, "\nCapabilities:")
	scanAlways, err := mgr.IsScanAlwaysAvailable()
	if err == nil {
		fmt.Fprintf(output, "  Scan always available: %v\n", scanAlways)
	}
	band5g, err := mgr.Is5GHzBandSupported()
	if err == nil {
		fmt.Fprintf(output, "  5 GHz supported: %v\n", band5g)
	}
	band6g, err := mgr.Is6GHzBandSupported()
	if err == nil {
		fmt.Fprintf(output, "  6 GHz supported: %v\n", band6g)
	}
	band60g, err := mgr.Is60GHzBandSupported()
	if err == nil {
		fmt.Fprintf(output, "  60 GHz supported: %v\n", band60g)
	}
	p2p, err := mgr.IsP2pSupported()
	if err == nil {
		fmt.Fprintf(output, "  P2P supported: %v\n", p2p)
	}
	tdls, err := mgr.IsTdlsSupported()
	if err == nil {
		fmt.Fprintf(output, "  TDLS supported: %v\n", tdls)
	}

	maxSig, err := mgr.GetMaxSignalLevel()
	if err == nil {
		fmt.Fprintf(output, "  Max signal level: %d\n", maxSig)
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "WiFi site survey complete.")
	return nil
}
