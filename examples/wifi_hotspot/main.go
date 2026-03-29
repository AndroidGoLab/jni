//go:build android

// Command wifi_hotspot queries the WiFi state and hotspot-related API
// surface using the wifi.Manager, including band support, concurrent
// mode capabilities, and the soft AP configuration API.
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

	fmt.Fprintln(output, "=== WiFi Hotspot ===")

	// WiFi enabled state.
	enabled, err := mgr.IsWifiEnabled()
	if err != nil {
		fmt.Fprintf(output, "IsWifiEnabled: %v\n", err)
	} else {
		fmt.Fprintf(output, "WiFi enabled: %v\n", enabled)
	}

	// WiFi state (numeric).
	state, err := mgr.GetWifiState()
	if err != nil {
		fmt.Fprintf(output, "GetWifiState: %v\n", err)
	} else {
		stateName := "unknown"
		switch state {
		case int32(wifi.WifiStateDisabled):
			stateName = "disabled"
		case int32(wifi.WifiStateDisabling):
			stateName = "disabling"
		case int32(wifi.WifiStateEnabled):
			stateName = "enabled"
		case int32(wifi.WifiStateEnabling):
			stateName = "enabling"
		case int32(wifi.WifiStateUnknown):
			stateName = "unknown"
		}
		fmt.Fprintf(output, "WiFi state: %s (%d)\n", stateName, state)
	}

	// Band support.
	fmt.Fprintln(output, "\nBand support:")
	band24, err := mgr.Is24GHzBandSupported()
	if err != nil {
		fmt.Fprintf(output, "  2.4 GHz: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  2.4 GHz: %v\n", band24)
	}

	band5, err := mgr.Is5GHzBandSupported()
	if err != nil {
		fmt.Fprintf(output, "  5 GHz: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  5 GHz: %v\n", band5)
	}

	band6, err := mgr.Is6GHzBandSupported()
	if err != nil {
		fmt.Fprintf(output, "  6 GHz: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  6 GHz: %v\n", band6)
	}

	band60, err := mgr.Is60GHzBandSupported()
	if err != nil {
		fmt.Fprintf(output, "  60 GHz: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  60 GHz: %v\n", band60)
	}

	// Hotspot / AP concurrency capabilities.
	fmt.Fprintln(output, "\nHotspot capabilities:")

	staAp, err := mgr.IsStaApConcurrencySupported()
	if err != nil {
		fmt.Fprintf(output, "  STA+AP concurrency: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  STA+AP concurrency: %v\n", staAp)
	}

	bridgedAp, err := mgr.IsBridgedApConcurrencySupported()
	if err != nil {
		fmt.Fprintf(output, "  Bridged AP: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  Bridged AP: %v\n", bridgedAp)
	}

	staBridgedAp, err := mgr.IsStaBridgedApConcurrencySupported()
	if err != nil {
		fmt.Fprintf(output, "  STA+Bridged AP: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  STA+Bridged AP: %v\n", staBridgedAp)
	}

	dualBand, err := mgr.IsDualBandSimultaneousSupported()
	if err != nil {
		fmt.Fprintf(output, "  Dual band simultaneous: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  Dual band simultaneous: %v\n", dualBand)
	}

	// Additional features relevant to hotspot.
	fmt.Fprintln(output, "\nAdditional features:")

	easyConnect, err := mgr.IsEasyConnectSupported()
	if err != nil {
		fmt.Fprintf(output, "  Easy Connect (DPP): error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  Easy Connect (DPP): %v\n", easyConnect)
	}

	enhancedOpen, err := mgr.IsEnhancedOpenSupported()
	if err != nil {
		fmt.Fprintf(output, "  Enhanced Open (OWE): error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  Enhanced Open (OWE): %v\n", enhancedOpen)
	}

	wpa3, err := mgr.IsWpa3SaeSupported()
	if err != nil {
		fmt.Fprintf(output, "  WPA3-SAE: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  WPA3-SAE: %v\n", wpa3)
	}

	maxSignal, err := mgr.GetMaxSignalLevel()
	if err != nil {
		fmt.Fprintf(output, "  Max signal level: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "  Max signal level: %d\n", maxSignal)
	}

	// Multi-internet mode.
	multiMode, err := mgr.GetStaConcurrencyForMultiInternetMode()
	if err != nil {
		fmt.Fprintf(output, "  Multi-internet mode: error (%v)\n", err)
	} else {
		modeName := "unknown"
		switch multiMode {
		case int32(wifi.WifiMultiInternetModeDisabled):
			modeName = "disabled"
		case int32(wifi.WifiMultiInternetModeDbsAp):
			modeName = "DBS AP"
		case int32(wifi.WifiMultiInternetModeMultiAp):
			modeName = "multi AP"
		}
		fmt.Fprintf(output, "  Multi-internet mode: %s (%d)\n", modeName, multiMode)
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "WiFi hotspot example complete.")
	return nil
}
