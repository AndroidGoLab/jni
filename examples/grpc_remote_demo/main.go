//go:build android

// Command grpc_remote_demo demonstrates using multiple Android APIs
// together as would be exposed via gRPC: read build info, battery,
// location state, and wifi state. Shows the data that jni-proxy
// would serialize.
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
	"github.com/AndroidGoLab/jni/location"
	"github.com/AndroidGoLab/jni/net/wifi"
	"github.com/AndroidGoLab/jni/os/battery"
	"github.com/AndroidGoLab/jni/os/build"
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

	fmt.Fprintln(output, "=== gRPC Remote Demo ===")
	fmt.Fprintln(output, "Data available for serialization:")

	// --- Build Info ---
	fmt.Fprintln(output, "\n[BuildInfo]")
	buildInfo, err := build.GetBuildInfo(vm)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Device: %s\n", buildInfo.Device)
		fmt.Fprintf(output, "  Model: %s\n", buildInfo.Model)
		fmt.Fprintf(output, "  Manufacturer: %s\n", buildInfo.Manufacturer)
		fmt.Fprintf(output, "  Brand: %s\n", buildInfo.Brand)
		fmt.Fprintf(output, "  Product: %s\n", buildInfo.Product)
		fmt.Fprintf(output, "  Board: %s\n", buildInfo.Board)
		fmt.Fprintf(output, "  Hardware: %s\n", buildInfo.Hardware)
	}

	versionInfo, err := build.GetVersionInfo(vm)
	if err != nil {
		fmt.Fprintf(output, "  VersionErr: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Release: %s\n", versionInfo.Release)
		fmt.Fprintf(output, "  SDK: %d\n", versionInfo.SDKInt)
		fmt.Fprintf(output, "  Codename: %s\n", versionInfo.Codename)
	}

	// --- Battery ---
	fmt.Fprintln(output, "\n[Battery]")
	batMgr, err := battery.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else {
		defer batMgr.Close()

		level, err := batMgr.GetIntProperty(int32(battery.BatteryPropertyCapacity))
		if err != nil {
			fmt.Fprintf(output, "  Level: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Level: %d%%\n", level)
		}

		charging, err := batMgr.IsCharging()
		if err != nil {
			fmt.Fprintf(output, "  Charging: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Charging: %v\n", charging)
		}

		chargeTime, err := batMgr.ComputeChargeTimeRemaining()
		if err != nil {
			fmt.Fprintf(output, "  ChargeTime: %v\n", err)
		} else {
			fmt.Fprintf(output, "  ChargeTimeMs: %d\n", chargeTime)
		}
	}

	// --- Location ---
	fmt.Fprintln(output, "\n[Location]")
	locMgr, err := location.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else {
		defer locMgr.Close()

		locEnabled, err := locMgr.IsLocationEnabled()
		if err != nil {
			fmt.Fprintf(output, "  Enabled: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Enabled: %v\n", locEnabled)
		}

		gpsEnabled, err := locMgr.IsProviderEnabled("gps")
		if err != nil {
			fmt.Fprintf(output, "  GPS: %v\n", err)
		} else {
			fmt.Fprintf(output, "  GPS: %v\n", gpsEnabled)
		}

		networkEnabled, err := locMgr.IsProviderEnabled("network")
		if err != nil {
			fmt.Fprintf(output, "  Network: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Network: %v\n", networkEnabled)
		}

		gnssHw, err := locMgr.GetGnssHardwareModelName()
		if err != nil {
			fmt.Fprintf(output, "  GNSS HW: %v\n", err)
		} else {
			fmt.Fprintf(output, "  GNSS HW: %s\n", gnssHw)
		}

		gnssYear, err := locMgr.GetGnssYearOfHardware()
		if err != nil {
			fmt.Fprintf(output, "  GNSS Year: %v\n", err)
		} else {
			fmt.Fprintf(output, "  GNSS Year: %d\n", gnssYear)
		}
	}

	// --- WiFi ---
	fmt.Fprintln(output, "\n[WiFi]")
	wifiMgr, err := wifi.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
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
	}

	fmt.Fprintln(output, "\ngRPC remote demo complete.")
	return nil
}
