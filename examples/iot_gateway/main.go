//go:build android

// Command iot_gateway is an IoT gateway concept: checks Bluetooth, WiFi,
// and NFC availability, reports which radios are available and their
// states, showing multi-radio coordination.
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
	"github.com/AndroidGoLab/jni/bluetooth"
	"github.com/AndroidGoLab/jni/content/pm"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/net/wifi"
	"github.com/AndroidGoLab/jni/nfc"
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

	fmt.Fprintln(output, "=== IoT Gateway ===")
	fmt.Fprintln(output, "Radio Availability Check:")

	// Check hardware features via PackageManager.
	pmObj, err := ctx.GetPackageManager()
	if err != nil {
		return fmt.Errorf("GetPackageManager: %w", err)
	}
	mgr := pm.PackageManager{VM: vm, Obj: pmObj}

	hasBT, _ := mgr.HasSystemFeature1(pm.FeatureBluetooth)
	hasBLE, _ := mgr.HasSystemFeature1(pm.FeatureBluetoothLe)
	hasWifi, _ := mgr.HasSystemFeature1(pm.FeatureWifi)
	hasNFC, _ := mgr.HasSystemFeature1(pm.FeatureNfc)

	fmt.Fprintln(output, "\n[Hardware Features]")
	fmt.Fprintf(output, "  Bluetooth: %v\n", hasBT)
	fmt.Fprintf(output, "  BLE: %v\n", hasBLE)
	fmt.Fprintf(output, "  WiFi: %v\n", hasWifi)
	fmt.Fprintf(output, "  NFC: %v\n", hasNFC)

	// --- Bluetooth ---
	fmt.Fprintln(output, "\n[Bluetooth]")
	if hasBT {
		btAdapter, err := bluetooth.NewAdapter(ctx)
		if err != nil {
			fmt.Fprintf(output, "  Adapter: %v\n", err)
		} else {
			defer btAdapter.Close()

			enabled, err := btAdapter.IsEnabled()
			if err != nil {
				fmt.Fprintf(output, "  Enabled: %v\n", err)
			} else {
				fmt.Fprintf(output, "  Enabled: %v\n", enabled)
			}

			name, err := btAdapter.GetName()
			if err != nil {
				fmt.Fprintf(output, "  Name: %v\n", err)
			} else {
				fmt.Fprintf(output, "  Name: %s\n", name)
			}

			state, err := btAdapter.GetState()
			if err != nil {
				fmt.Fprintf(output, "  State: %v\n", err)
			} else {
				fmt.Fprintf(output, "  State: %d\n", state)
			}

			leSupported, _ := btAdapter.IsMultipleAdvertisementSupported()
			fmt.Fprintf(output, "  LE Multi-Ad: %v\n", leSupported)
		}
	} else {
		fmt.Fprintln(output, "  Not available")
	}

	// --- WiFi ---
	fmt.Fprintln(output, "\n[WiFi]")
	if hasWifi {
		wifiMgr, err := wifi.NewManager(ctx)
		if err != nil {
			fmt.Fprintf(output, "  Manager: %v\n", err)
		} else {
			defer wifiMgr.Close()

			enabled, err := wifiMgr.IsWifiEnabled()
			if err != nil {
				fmt.Fprintf(output, "  Enabled: %v\n", err)
			} else {
				fmt.Fprintf(output, "  Enabled: %v\n", enabled)
			}

			state, err := wifiMgr.GetWifiState()
			if err != nil {
				fmt.Fprintf(output, "  State: %v\n", err)
			} else {
				fmt.Fprintf(output, "  State: %d\n", state)
			}

			is5g, _ := wifiMgr.Is5GHzBandSupported()
			is6g, _ := wifiMgr.Is6GHzBandSupported()
			fmt.Fprintf(output, "  5GHz: %v\n", is5g)
			fmt.Fprintf(output, "  6GHz: %v\n", is6g)
		}
	} else {
		fmt.Fprintln(output, "  Not available")
	}

	// --- NFC ---
	fmt.Fprintln(output, "\n[NFC]")
	if hasNFC {
		nfcMgr, err := nfc.NewManager(ctx)
		if err != nil {
			fmt.Fprintf(output, "  Manager: %v\n", err)
		} else {
			defer nfcMgr.Close()
			fmt.Fprintln(output, "  NfcManager: obtained")

			adapterObj, err := nfcMgr.GetDefaultAdapter()
			if err != nil {
				fmt.Fprintf(output, "  Adapter: %v\n", err)
			} else if adapterObj != nil && adapterObj.Ref() != 0 {
				adapter := nfc.Adapter{VM: vm, Obj: adapterObj}

				enabled, err := adapter.IsEnabled()
				if err != nil {
					fmt.Fprintf(output, "  Enabled: %v\n", err)
				} else {
					fmt.Fprintf(output, "  Enabled: %v\n", enabled)
				}

				secureNfc, err := adapter.IsSecureNfcSupported()
				if err != nil {
					fmt.Fprintf(output, "  SecureNFC: %v\n", err)
				} else {
					fmt.Fprintf(output, "  SecureNFC: %v\n", secureNfc)
				}
			}
		}
	} else {
		fmt.Fprintln(output, "  Not available")
	}

	// Summary.
	radioCount := 0
	if hasBT {
		radioCount++
	}
	if hasWifi {
		radioCount++
	}
	if hasNFC {
		radioCount++
	}
	fmt.Fprintf(output, "\nRadios available: %d/3\n", radioCount)

	fmt.Fprintln(output, "\nIoT gateway complete.")
	return nil
}
