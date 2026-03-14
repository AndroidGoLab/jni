//go:build android

// Command alarm demonstrates scheduling alarms via the Android
// AlarmManager system service, wrapped by the alarm package. It is
// built as a c-shared library and packaged into an APK using the
// shared apk.mk infrastructure.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/app/alarm"
	"github.com/xaionaro-go/jni/app"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	vm := jni.VMFromPtr(unsafe.Pointer(cvm))
	if err := run(vm); err != nil {
		fmt.Fprintf(&output, "ERROR: %v\n", err)
	}
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}

func run(vm *jni.VM) error {
	ctx, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := alarm.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("alarm.NewManager: %v", err)
	}
	defer mgr.Close()

	// Check whether the app is allowed to schedule exact alarms (API 31+).
	canSchedule, err := mgr.CanScheduleExactAlarms()
	if err != nil {
		return fmt.Errorf("CanScheduleExactAlarms: %v", err)
	}
	fmt.Fprintf(&output, "can schedule exact alarms: %v\n", canSchedule)

	// Alarm type constants correspond to android.app.AlarmManager fields:
	//   RTC_WAKEUP, RTC, ELAPSED_REALTIME_WAKEUP, ELAPSED_REALTIME.
	fmt.Fprintf(&output, "alarm types: RTC_WAKEUP=%d, RTC=%d, ELAPSED_REALTIME_WAKEUP=%d, ELAPSED_REALTIME=%d\n",
		alarm.RTCWakeup, alarm.RTC, alarm.ElapsedRealtimeWakeup, alarm.ElapsedRealtime)

	// The following AlarmManager methods require a valid PendingIntent and
	// cannot be called with nil. In a real app, obtain a PendingIntent via
	// PendingIntent.getBroadcast or similar, then pass it to these methods:
	fmt.Fprintf(&output, "AlarmManager methods that require a PendingIntent:\n")
	fmt.Fprintf(&output, "  Cancel(op)                                  - cancel a scheduled alarm\n")
	fmt.Fprintf(&output, "  set(type, triggerAtMillis, op)               - schedule an alarm\n")
	fmt.Fprintf(&output, "  setExact(type, triggerAtMillis, op)          - schedule an exact alarm\n")
	fmt.Fprintf(&output, "  setExactAndAllowWhileIdle(type, millis, op)  - exact alarm, even in idle\n")
	fmt.Fprintf(&output, "  setRepeating(type, trigger, interval, op)   - schedule a repeating alarm\n")
	fmt.Fprintf(&output, "  setWindow(type, start, length, op)          - schedule within a window\n")
	fmt.Fprintf(&output, "  setAlarmClock(info, op)                     - schedule an alarm clock\n")

	return nil
}

// getAppContext obtains an Android Context via ActivityThread.currentApplication().
func getAppContext(vm *jni.VM) (*app.Context, error) {
	var ctx app.Context
	ctx.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}

		atClass, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread: %w", err)
		}

		curAppMid, err := env.GetStaticMethodID(atClass, "currentApplication", "()Landroid/app/Application;")
		if err != nil {
			return fmt.Errorf("get currentApplication: %w", err)
		}
		appObj, err := env.CallStaticObjectMethod(atClass, curAppMid)
		if err != nil {
			return fmt.Errorf("call currentApplication: %w", err)
		}
		if appObj == nil || appObj.Ref() == 0 {
			return fmt.Errorf("currentApplication returned null")
		}

		ctx.Obj = env.NewGlobalRef(appObj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}
