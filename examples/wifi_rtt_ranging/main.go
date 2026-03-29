//go:build android

// Command wifi_rtt_ranging demonstrates the WiFi RTT (Round-Trip Time) API:
// checks availability and queries RTT characteristics using typed wrappers.
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
	"github.com/AndroidGoLab/jni/net/wifi/rtt"
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

	fmt.Fprintln(output, "=== WiFi RTT Ranging ===")

	// Check if device supports RTT via WifiRttManager.
	rttMgr, err := rtt.NewWifiRttManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "WifiRttManager not available: %v\n", err)
		fmt.Fprintln(output, "(WiFi RTT requires Android 9+ and hardware support)")
	} else {
		defer rttMgr.Close()
		fmt.Fprintln(output, "WifiRttManager obtained")

		available, err := rttMgr.IsAvailable()
		if err != nil {
			fmt.Fprintf(output, "IsAvailable error: %v\n", err)
		} else {
			fmt.Fprintf(output, "RTT available: %v\n", available)
		}

		// Query RTT characteristics (returned as opaque object; no exported wrapper).
		chars, err := rttMgr.GetRttCharacteristics()
		if err != nil {
			fmt.Fprintf(output, "GetRttCharacteristics: %v\n", err)
		} else if chars != nil && chars.Ref() != 0 {
			fmt.Fprintln(output, "RTT characteristics: obtained")
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(chars)
				return nil
			})
		}
	}

	// Query WiFi scan results and check for RTT-capable APs.
	wifiMgr, err := wifi.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "\nWifiManager not available: %v\n", err)
	} else {
		defer wifiMgr.Close()

		enabled, err := wifiMgr.IsWifiEnabled()
		if err != nil {
			fmt.Fprintf(output, "\nWiFi enabled check: %v\n", err)
		} else {
			fmt.Fprintf(output, "\nWiFi enabled: %v\n", enabled)
		}

		fmt.Fprintln(output, "\nScanning for RTT-capable APs...")

		scanListObj, err := wifiMgr.GetScanResults()
		if err != nil {
			fmt.Fprintf(output, "GetScanResults: %v\n", err)
		} else if scanListObj != nil && scanListObj.Ref() != 0 {
			// GetScanResults returns a Java List; we need the WiFi connection info instead.
			// Since List iteration requires raw JNI, show WiFi connection info.
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(scanListObj)
				return nil
			})
			fmt.Fprintln(output, "Scan results obtained (List object)")
			fmt.Fprintln(output, "Note: Iterating Java List<ScanResult> requires raw JNI;")
			fmt.Fprintln(output, "use WifiManager.GetConnectionInfo() for current AP info.")
		}

		// Show connection info as typed wrapper.
		connInfo, err := wifiMgr.GetConnectionInfo()
		if err != nil {
			fmt.Fprintf(output, "\nGetConnectionInfo: %v\n", err)
		} else if connInfo != nil && connInfo.Ref() != 0 {
			wInfo := &wifi.Info{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(connInfo))}
			rssi, err := wInfo.GetRssi()
			if err == nil {
				fmt.Fprintf(output, "\nCurrent AP RSSI: %d dBm\n", rssi)
			}
			freq, err := wInfo.GetFrequency()
			if err == nil {
				fmt.Fprintf(output, "Current AP frequency: %d MHz\n", freq)
			}
			linkSpeed, err := wInfo.GetLinkSpeed()
			if err == nil {
				fmt.Fprintf(output, "Link speed: %d Mbps\n", linkSpeed)
			}
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(connInfo)
				return nil
			})
		}
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "WiFi RTT ranging example complete.")
	return nil
}
