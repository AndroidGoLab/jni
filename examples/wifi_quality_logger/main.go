//go:build android

// Command wifi_quality_logger retrieves the current WiFi connection info
// including SSID, BSSID, link speed, frequency, RSSI, and IP address
// using the wifi.Manager.GetConnectionInfo and wifi.Info methods.
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

// ipIntToString converts an Android IP address int (little-endian) to dotted notation.
func ipIntToString(ip int32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip&0xFF,
		(ip>>8)&0xFF,
		(ip>>16)&0xFF,
		(ip>>24)&0xFF,
	)
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

	fmt.Fprintln(output, "=== WiFi Quality Logger ===")

	// Check WiFi enabled state.
	enabled, err := mgr.IsWifiEnabled()
	if err != nil {
		fmt.Fprintf(output, "IsWifiEnabled error: %v\n", err)
	} else {
		fmt.Fprintf(output, "WiFi enabled: %v\n", enabled)
	}

	// GetConnectionInfo returns a raw JNI object (WifiInfo).
	// We wrap it as a wifi.Info to use typed methods.
	connObj, err := mgr.GetConnectionInfo()
	if err != nil {
		return fmt.Errorf("GetConnectionInfo: %w", err)
	}
	if connObj == nil || connObj.Ref() == 0 {
		fmt.Fprintln(output, "No connection info available (not connected?)")
		return nil
	}

	info := &wifi.Info{VM: vm, Obj: connObj}
	defer func() {
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(info.Obj)
			return nil
		})
	}()

	// SSID.
	ssid, err := info.GetSSID()
	if err != nil {
		fmt.Fprintf(output, "SSID: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "SSID: %s\n", ssid)
	}

	// BSSID.
	bssid, err := info.GetBSSID()
	if err != nil {
		fmt.Fprintf(output, "BSSID: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "BSSID: %s\n", bssid)
	}

	// RSSI.
	rssi, err := info.GetRssi()
	if err != nil {
		fmt.Fprintf(output, "RSSI: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "RSSI: %d dBm\n", rssi)
	}

	// Link speed.
	linkSpeed, err := info.GetLinkSpeed()
	if err != nil {
		fmt.Fprintf(output, "Link speed: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Link speed: %d Mbps\n", linkSpeed)
	}

	// TX link speed.
	txSpeed, err := info.GetTxLinkSpeedMbps()
	if err != nil {
		fmt.Fprintf(output, "TX speed: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "TX speed: %d Mbps\n", txSpeed)
	}

	// RX link speed.
	rxSpeed, err := info.GetRxLinkSpeedMbps()
	if err != nil {
		fmt.Fprintf(output, "RX speed: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "RX speed: %d Mbps\n", rxSpeed)
	}

	// Frequency.
	freq, err := info.GetFrequency()
	if err != nil {
		fmt.Fprintf(output, "Frequency: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Frequency: %d MHz\n", freq)
	}

	// IP address.
	ipAddr, err := info.GetIpAddress()
	if err != nil {
		fmt.Fprintf(output, "IP address: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "IP address: %s\n", ipIntToString(ipAddr))
	}

	// Network ID.
	netID, err := info.GetNetworkId()
	if err != nil {
		fmt.Fprintf(output, "Network ID: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Network ID: %d\n", netID)
	}

	// WiFi standard.
	standard, err := info.GetWifiStandard()
	if err != nil {
		fmt.Fprintf(output, "WiFi standard: error (%v)\n", err)
	} else {
		standardName := "unknown"
		switch standard {
		case int32(wifi.WifiStandardLegacy):
			standardName = "legacy"
		case int32(wifi.WifiStandard11n):
			standardName = "802.11n (WiFi 4)"
		case int32(wifi.WifiStandard11ac):
			standardName = "802.11ac (WiFi 5)"
		case int32(wifi.WifiStandard11ax):
			standardName = "802.11ax (WiFi 6)"
		case int32(wifi.WifiStandard11be):
			standardName = "802.11be (WiFi 7)"
		}
		fmt.Fprintf(output, "WiFi standard: %s (%d)\n", standardName, standard)
	}

	// Max supported speeds.
	maxTx, err := info.GetMaxSupportedTxLinkSpeedMbps()
	if err != nil {
		fmt.Fprintf(output, "Max TX speed: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Max TX speed: %d Mbps\n", maxTx)
	}

	maxRx, err := info.GetMaxSupportedRxLinkSpeedMbps()
	if err != nil {
		fmt.Fprintf(output, "Max RX speed: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Max RX speed: %d Mbps\n", maxRx)
	}

	// Hidden SSID.
	hidden, err := info.GetHiddenSSID()
	if err != nil {
		fmt.Fprintf(output, "Hidden SSID: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Hidden SSID: %v\n", hidden)
	}

	// Full toString() for reference.
	fullStr, err := info.ToString()
	if err != nil {
		fmt.Fprintf(output, "ToString: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "\nFull WifiInfo:\n  %s\n", fullStr)
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "WiFi quality logger complete.")
	return nil
}
