//go:build android

// Command battery_dashboard provides a comprehensive battery display:
// level, status, temperature, technology, current, voltage, energy.
// It exercises all BatteryManager property query methods.
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
	"github.com/AndroidGoLab/jni/os/battery"
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

func statusName(status int32) string {
	switch int(status) {
	case battery.BatteryStatusUnknown:
		return "unknown"
	case battery.BatteryStatusCharging:
		return "charging"
	case battery.BatteryStatusDischarging:
		return "discharging"
	case battery.BatteryStatusNotCharging:
		return "not charging"
	case battery.BatteryStatusFull:
		return "full"
	default:
		return fmt.Sprintf("unknown(%d)", status)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := battery.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("battery.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Battery Dashboard ===")
	fmt.Fprintln(output)

	// --- Charging state ---
	charging, err := mgr.IsCharging()
	if err != nil {
		fmt.Fprintf(output, "charging: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "charging: %v\n", charging)
	}

	// --- Capacity (0-100%) ---
	capacity, err := mgr.GetIntProperty(int32(battery.BatteryPropertyCapacity))
	if err != nil {
		fmt.Fprintf(output, "capacity: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "capacity: %d%%\n", capacity)
	}

	// --- Status ---
	status, err := mgr.GetIntProperty(int32(battery.BatteryPropertyStatus))
	if err != nil {
		fmt.Fprintf(output, "status: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "status: %s (%d)\n", statusName(status), status)
	}

	// --- Current now (microamperes) ---
	currentNow, err := mgr.GetIntProperty(int32(battery.BatteryPropertyCurrentNow))
	if err != nil {
		fmt.Fprintf(output, "current now: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "current now: %d uA\n", currentNow)
	}

	// --- Current average (microamperes) ---
	currentAvg, err := mgr.GetIntProperty(int32(battery.BatteryPropertyCurrentAverage))
	if err != nil {
		fmt.Fprintf(output, "current avg: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "current avg: %d uA\n", currentAvg)
	}

	// --- Charge counter (microampere-hours) ---
	chargeCounter, err := mgr.GetIntProperty(int32(battery.BatteryPropertyChargeCounter))
	if err != nil {
		fmt.Fprintf(output, "charge counter: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "charge counter: %d uAh\n", chargeCounter)
	}

	// --- Energy counter (nanowatt-hours) ---
	energy, err := mgr.GetLongProperty(int32(battery.BatteryPropertyEnergyCounter))
	if err != nil {
		fmt.Fprintf(output, "energy counter: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "energy counter: %d nWh\n", energy)
	}

	// --- String representation of each property ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "String properties:")
	for _, prop := range []struct {
		name string
		id   int
	}{
		{"capacity", battery.BatteryPropertyCapacity},
		{"charge_counter", battery.BatteryPropertyChargeCounter},
		{"current_now", battery.BatteryPropertyCurrentNow},
		{"current_avg", battery.BatteryPropertyCurrentAverage},
		{"status", battery.BatteryPropertyStatus},
		{"energy_counter", battery.BatteryPropertyEnergyCounter},
	} {
		s, err := mgr.GetStringProperty(int32(prop.id))
		if err != nil {
			fmt.Fprintf(output, "  %s: error: %v\n", prop.name, err)
		} else if s != "" {
			fmt.Fprintf(output, "  %s: %s\n", prop.name, s)
		} else {
			fmt.Fprintf(output, "  %s: (empty)\n", prop.name)
		}
	}

	// --- Charge time remaining ---
	fmt.Fprintln(output)
	chargeTime, err := mgr.ComputeChargeTimeRemaining()
	if err != nil {
		fmt.Fprintf(output, "charge time remaining: error: %v\n", err)
	} else if chargeTime < 0 {
		fmt.Fprintln(output, "charge time remaining: N/A (not charging)")
	} else {
		minutes := chargeTime / 60000
		fmt.Fprintf(output, "charge time remaining: ~%d min (%d ms)\n", minutes, chargeTime)
	}

	// --- Property constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Property constants:")
	fmt.Fprintf(output, "  CAPACITY       = %d\n", battery.BatteryPropertyCapacity)
	fmt.Fprintf(output, "  CHARGE_COUNTER = %d\n", battery.BatteryPropertyChargeCounter)
	fmt.Fprintf(output, "  CURRENT_NOW    = %d\n", battery.BatteryPropertyCurrentNow)
	fmt.Fprintf(output, "  CURRENT_AVG    = %d\n", battery.BatteryPropertyCurrentAverage)
	fmt.Fprintf(output, "  ENERGY_COUNTER = %d\n", battery.BatteryPropertyEnergyCounter)
	fmt.Fprintf(output, "  STATUS         = %d\n", battery.BatteryPropertyStatus)

	// --- Status constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Status constants:")
	fmt.Fprintf(output, "  UNKNOWN      = %d\n", battery.BatteryStatusUnknown)
	fmt.Fprintf(output, "  CHARGING     = %d\n", battery.BatteryStatusCharging)
	fmt.Fprintf(output, "  DISCHARGING  = %d\n", battery.BatteryStatusDischarging)
	fmt.Fprintf(output, "  NOT_CHARGING = %d\n", battery.BatteryStatusNotCharging)
	fmt.Fprintf(output, "  FULL         = %d\n", battery.BatteryStatusFull)

	return nil
}
