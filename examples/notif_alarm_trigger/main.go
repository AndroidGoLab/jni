//go:build android

// Command notif_alarm_trigger schedules an alarm and then posts a
// notification, demonstrating cross-package integration of
// alarm + notification packages.
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
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/app/alarm"
	"github.com/AndroidGoLab/jni/app/notification"
	"github.com/AndroidGoLab/jni/examples/common/ui"
)

const (
	// PendingIntent.FLAG_IMMUTABLE
	flagImmutable = 67108864

	// android.R.drawable.ic_dialog_info
	iconDialogInfo = 17301620

	channelID   = "alarm_notif"
	channelName = "Alarm Notification"
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

	fmt.Fprintln(output, "=== Alarm + Notification ===")
	fmt.Fprintln(output)

	// --- Alarm scheduling ---
	amgr, err := alarm.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("alarm.NewManager: %w", err)
	}
	defer amgr.Close()

	canSchedule, err := amgr.CanScheduleExactAlarms()
	if err != nil {
		return fmt.Errorf("CanScheduleExactAlarms: %w", err)
	}
	fmt.Fprintf(output, "can schedule: %v\n", canSchedule)

	if canSchedule {
		delay := 30 * time.Second
		triggerAt := time.Now().Add(delay)

		intent, err := app.NewIntent(vm, nil, nil)
		if err != nil {
			return fmt.Errorf("new intent: %w", err)
		}
		if _, err := intent.SetAction("center.dx.jni.examples.notif_alarm_trigger.FIRED"); err != nil {
			return fmt.Errorf("set action: %w", err)
		}

		pi := app.PendingIntent{VM: vm}
		pendingObj, err := pi.GetBroadcast(ctx.Obj, 0, intent.Obj, flagImmutable)
		if err != nil {
			return fmt.Errorf("getBroadcast: %w", err)
		}

		if err := amgr.Set(int32(alarm.RtcWakeup), triggerAt.UnixMilli(), pendingObj); err != nil {
			return fmt.Errorf("schedule alarm: %w", err)
		}
		fmt.Fprintf(output, "alarm in %s\n", delay)
	} else {
		fmt.Fprintln(output, "exact alarms not permitted")
	}

	// --- Notification posting ---
	nmgr, err := notification.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("notification.NewManager: %w", err)
	}
	defer nmgr.Close()

	ch, err := notification.NewChannel(
		vm, channelID, channelName,
		int32(notification.ImportanceHigh),
	)
	if err != nil {
		return fmt.Errorf("new channel: %w", err)
	}
	if err := nmgr.CreateNotificationChannel(ch.Obj); err != nil {
		return fmt.Errorf("create channel: %w", err)
	}
	fmt.Fprintln(output, "channel created")

	appCtxObj, err := ctx.GetApplicationContext()
	if err != nil {
		return fmt.Errorf("get app context: %w", err)
	}

	b, err := notification.NewBuilder(vm, appCtxObj, channelID)
	if err != nil {
		return fmt.Errorf("new builder: %w", err)
	}
	if _, err := b.SetSmallIcon1_1(iconDialogInfo); err != nil {
		return fmt.Errorf("set icon: %w", err)
	}
	if _, err := b.SetContentTitle("Alarm scheduled!"); err != nil {
		return fmt.Errorf("set title: %w", err)
	}
	if _, err := b.SetContentText("Alarm + notification from Go via JNI"); err != nil {
		return fmt.Errorf("set text: %w", err)
	}
	if _, err := b.SetAutoCancel(true); err != nil {
		return fmt.Errorf("set auto cancel: %w", err)
	}
	notif, err := b.Build()
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}
	if err := nmgr.Notify2(1, notif); err != nil {
		return fmt.Errorf("notify: %w", err)
	}
	fmt.Fprintln(output, "notification posted!")

	return nil
}
