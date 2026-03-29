//go:build android

// Command power_save_detect checks PowerManager for power save mode,
// device idle mode, interactive state, and thermal status using typed
// wrappers.
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
	"github.com/AndroidGoLab/jni/os/power"
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

func thermalStatusName(status int32) string {
	switch int(status) {
	case power.ThermalStatusNone:
		return "NONE"
	case power.ThermalStatusLight:
		return "LIGHT"
	case power.ThermalStatusModerate:
		return "MODERATE"
	case power.ThermalStatusSevere:
		return "SEVERE"
	case power.ThermalStatusCritical:
		return "CRITICAL"
	case power.ThermalStatusEmergency:
		return "EMERGENCY"
	case power.ThermalStatusShutdown:
		return "SHUTDOWN"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", status)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := power.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("power.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Power Save Detect ===")
	fmt.Fprintln(output)

	// --- Power save mode ---
	powerSave, err := mgr.IsPowerSaveMode()
	if err != nil {
		fmt.Fprintf(output, "power save mode: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "power save mode: %v\n", powerSave)
	}

	// --- Interactive ---
	interactive, err := mgr.IsInteractive()
	if err != nil {
		fmt.Fprintf(output, "interactive: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "interactive: %v\n", interactive)
	}

	// --- Screen on ---
	screenOn, err := mgr.IsScreenOn()
	if err != nil {
		fmt.Fprintf(output, "screen on: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "screen on: %v\n", screenOn)
	}

	// --- Device idle mode (Doze) ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Doze state:")

	idleMode, err := mgr.IsDeviceIdleMode()
	if err != nil {
		fmt.Fprintf(output, "  device idle: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  device idle: %v\n", idleMode)
	}

	lightIdle, err := mgr.IsDeviceLightIdleMode()
	if err != nil {
		fmt.Fprintf(output, "  light idle: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  light idle: %v\n", lightIdle)
	}

	// --- Thermal status ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Thermal:")

	thermalStatus, err := mgr.GetCurrentThermalStatus()
	if err != nil {
		fmt.Fprintf(output, "  status: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  status: %s (%d)\n", thermalStatusName(thermalStatus), thermalStatus)
	}

	headroom, err := mgr.GetThermalHeadroom(10)
	if err != nil {
		fmt.Fprintf(output, "  headroom(10s): error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  headroom(10s): %.3f\n", headroom)
	}

	// --- Low power standby ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Low power standby:")

	standbyEnabled, err := mgr.IsLowPowerStandbyEnabled()
	if err != nil {
		fmt.Fprintf(output, "  enabled: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  enabled: %v\n", standbyEnabled)
	}

	exemptFromStandby, err := mgr.IsExemptFromLowPowerStandby()
	if err != nil {
		fmt.Fprintf(output, "  exempt: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  exempt: %v\n", exemptFromStandby)
	}

	// --- Battery discharge prediction ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Battery discharge:")

	personalized, err := mgr.IsBatteryDischargePredictionPersonalized()
	if err != nil {
		fmt.Fprintf(output, "  personalized: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  personalized: %v\n", personalized)
	}

	prediction, err := mgr.GetBatteryDischargePrediction()
	if err != nil {
		fmt.Fprintf(output, "  prediction: error: %v\n", err)
	} else if prediction == nil {
		fmt.Fprintln(output, "  prediction: null")
	} else {
		fmt.Fprintln(output, "  prediction: (Duration object returned)")
	}

	// --- Sustained performance ---
	fmt.Fprintln(output)
	sustained, err := mgr.IsSustainedPerformanceModeSupported()
	if err != nil {
		fmt.Fprintf(output, "sustained perf: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "sustained perf supported: %v\n", sustained)
	}

	// --- Thermal status constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Thermal status constants:")
	fmt.Fprintf(output, "  NONE      = %d\n", power.ThermalStatusNone)
	fmt.Fprintf(output, "  LIGHT     = %d\n", power.ThermalStatusLight)
	fmt.Fprintf(output, "  MODERATE  = %d\n", power.ThermalStatusModerate)
	fmt.Fprintf(output, "  SEVERE    = %d\n", power.ThermalStatusSevere)
	fmt.Fprintf(output, "  CRITICAL  = %d\n", power.ThermalStatusCritical)
	fmt.Fprintf(output, "  EMERGENCY = %d\n", power.ThermalStatusEmergency)
	fmt.Fprintf(output, "  SHUTDOWN  = %d\n", power.ThermalStatusShutdown)

	return nil
}
