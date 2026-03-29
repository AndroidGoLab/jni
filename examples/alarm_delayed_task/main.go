//go:build android

// Command alarm_delayed_task demonstrates the AlarmManager API.
// It shows alarm constants, checks exact-alarm capability,
// and queries the next alarm clock status.
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
	"github.com/AndroidGoLab/jni/app/alarm"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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

	mgr, err := alarm.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("alarm.NewManager: %w", err)
	}
	defer mgr.Close()

	// --- Alarm type constants ---
	fmt.Fprintln(output, "=== Alarm Type Constants ===")
	fmt.Fprintf(output, "RTC_WAKEUP          = %d\n", alarm.RtcWakeup)
	fmt.Fprintf(output, "RTC                 = %d\n", alarm.Rtc)
	fmt.Fprintf(output, "ELAPSED_REALTIME    = %d\n", alarm.ElapsedRealtime)
	fmt.Fprintf(output, "ELAPSED_RT_WAKEUP   = %d\n", alarm.ElapsedRealtimeWakeup)

	// --- Check exact alarm scheduling capability ---
	canSchedule, err := mgr.CanScheduleExactAlarms()
	if err != nil {
		fmt.Fprintf(output, "\nCanScheduleExactAlarms: %v\n", err)
	} else {
		fmt.Fprintf(output, "\ncan schedule exact alarms: %v\n", canSchedule)
	}

	// --- Check next alarm clock ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Next Alarm Clock ===")
	nextAlarm, err := mgr.GetNextAlarmClock()
	if err != nil {
		fmt.Fprintf(output, "GetNextAlarmClock: %v\n", err)
	} else if nextAlarm != nil && nextAlarm.Ref() != 0 {
		fmt.Fprintln(output, "next alarm clock: present")
	} else {
		fmt.Fprintln(output, "next alarm clock: none")
	}

	// --- Interval constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Interval Constants ===")
	fmt.Fprintf(output, "INTERVAL_FIFTEEN_MINUTES = %d ms\n", alarm.IntervalFifteenMinutes)
	fmt.Fprintf(output, "INTERVAL_HALF_HOUR       = %d ms\n", alarm.IntervalHalfHour)
	fmt.Fprintf(output, "INTERVAL_HOUR            = %d ms\n", alarm.IntervalHour)
	fmt.Fprintf(output, "INTERVAL_HALF_DAY        = %d ms\n", alarm.IntervalHalfDay)
	fmt.Fprintf(output, "INTERVAL_DAY             = %d ms\n", alarm.IntervalDay)

	fmt.Fprintln(output, "\nalarm_delayed_task complete")
	return nil
}
