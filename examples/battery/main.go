//go:build android

// Command battery demonstrates reading live battery information from
// Android's BatteryManager system service. It calls every available
// query method: IsCharging, GetIntProperty, GetLongProperty,
// GetStringProperty, and ComputeChargeTimeRemaining.
package main

/*
#include <android/native_activity.h>
extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static void _setCallbacks(ANativeActivity* a) { a->callbacks->onResume = _onResume; a->callbacks->onNativeWindowCreated = _onWindowCreated; }
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
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
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

	fmt.Fprintln(output, "=== Battery Status ===")

	// IsCharging - boolean convenience.
	charging, err := mgr.IsCharging()
	if err != nil {
		fmt.Fprintf(output, "charging: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "charging: %v\n", charging)
	}

	// GetIntProperty - battery capacity (0-100%).
	capacity, err := mgr.GetIntProperty(int32(battery.BatteryPropertyCapacity))
	if err != nil {
		fmt.Fprintf(output, "capacity: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "capacity: %d%%\n", capacity)
	}

	// GetIntProperty - battery status.
	status, err := mgr.GetIntProperty(int32(battery.BatteryPropertyStatus))
	if err != nil {
		fmt.Fprintf(output, "status: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "status: %s (%d)\n", statusName(status), status)
	}

	// GetIntProperty - current now (microamperes).
	currentNow, err := mgr.GetIntProperty(int32(battery.BatteryPropertyCurrentNow))
	if err != nil {
		fmt.Fprintf(output, "current now: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "current now: %d uA\n", currentNow)
	}

	// GetIntProperty - current average (microamperes).
	currentAvg, err := mgr.GetIntProperty(int32(battery.BatteryPropertyCurrentAverage))
	if err != nil {
		fmt.Fprintf(output, "current avg: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "current avg: %d uA\n", currentAvg)
	}

	// GetIntProperty - charge counter (microampere-hours).
	chargeCounter, err := mgr.GetIntProperty(int32(battery.BatteryPropertyChargeCounter))
	if err != nil {
		fmt.Fprintf(output, "charge counter: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "charge counter: %d uAh\n", chargeCounter)
	}

	// GetLongProperty - energy counter (nanowatt-hours).
	energy, err := mgr.GetLongProperty(int32(battery.BatteryPropertyEnergyCounter))
	if err != nil {
		fmt.Fprintf(output, "energy counter: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "energy counter: %d nWh\n", energy)
	}

	// GetStringProperty - query string representation of each property.
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
			fmt.Fprintf(output, "str(%s): error: %v\n", prop.name, err)
		} else if s != "" {
			fmt.Fprintf(output, "str(%s): %s\n", prop.name, s)
		}
	}

	// ComputeChargeTimeRemaining (milliseconds, -1 if not charging).
	chargeTime, err := mgr.ComputeChargeTimeRemaining()
	if err != nil {
		fmt.Fprintf(output, "charge time remaining: error: %v\n", err)
	} else if chargeTime < 0 {
		fmt.Fprintln(output, "charge time remaining: N/A")
	} else {
		minutes := chargeTime / 60000
		fmt.Fprintf(output, "charge time remaining: ~%d min\n", minutes)
	}

	return nil
}
