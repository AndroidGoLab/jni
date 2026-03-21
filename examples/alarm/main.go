//go:build android

// Command alarm demonstrates scheduling alarms via the Android
// AlarmManager system service and playing the alarm ringtone — all
// from Go via JNI, with no custom Java code.
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
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/app/alarm"
	"github.com/AndroidGoLab/jni/app/notification"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/media/ringtone"
)

const (
	// PendingIntent.FLAG_IMMUTABLE
	flagImmutable = 67108864

	// android.R.drawable.ic_dialog_info
	iconDialogInfo = 17301620

	// Ringtone playback duration before auto-stop.
	ringtoneDuration = 5 * time.Second

	notificationChannelID   = "alarm"
	notificationChannelName = "Alarm"
)

func main() {}

func init() { exampleui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	println("ANativeActivity_onCreate called")
	exampleui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	exampleui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	exampleui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := alarm.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("alarm.NewManager: %w", err)
	}
	defer mgr.Close()

	canSchedule, err := mgr.CanScheduleExactAlarms()
	if err != nil {
		return fmt.Errorf("CanScheduleExactAlarms: %w", err)
	}
	fmt.Fprintf(output, "can schedule exact alarms: %v\n", canSchedule)
	fmt.Fprintf(output, "alarm types: RTC_WAKEUP=%d, RTC=%d\n",
		alarm.RtcWakeup, alarm.Rtc)

	delay := 60 * time.Second
	triggerAt := time.Now().Add(delay)

	// All typed wrappers return GlobalRefs, so no vm.Do nesting needed.
	intent, err := app.NewIntent(vm)
	if err != nil {
		return fmt.Errorf("new intent: %w", err)
	}
	if _, err := intent.SetAction("center.dx.jni.examples.alarm.FIRED"); err != nil {
		return fmt.Errorf("set action: %w", err)
	}

	pi := app.PendingIntent{VM: vm}
	pendingObj, err := pi.GetBroadcast(ctx.Obj, 0, intent.Obj, flagImmutable)
	if err != nil {
		return fmt.Errorf("getBroadcast: %w", err)
	}

	if err := mgr.Set(int32(alarm.RtcWakeup), triggerAt.UnixMilli(), pendingObj); err != nil {
		return fmt.Errorf("schedule alarm: %w", err)
	}
	fmt.Fprintf(output, "alarm scheduled for %s\n", triggerAt.Format(time.RFC3339))

	// GetApplicationContext returns a GlobalRef via the generated wrapper.
	appCtxObj, err := ctx.GetApplicationContext()
	if err != nil {
		return fmt.Errorf("get app context: %w", err)
	}

	fmt.Fprintf(output, "alarm will fire in %s\n", delay)
	go func() {
		time.Sleep(delay)
		if err := postNotification(vm, appCtxObj); err != nil {
			fmt.Fprintf(output, "notification error: %v\n", err)
		}
		if err := playAlarmRingtone(vm, appCtxObj); err != nil {
			fmt.Fprintf(output, "ringtone error: %v\n", err)
		}
		fmt.Fprintf(output, "alarm fired!\n")
		exampleui.RenderOutput()
	}()

	return nil
}

func postNotification(vm *jni.VM, appCtx *jni.Object) error {
	ctx := &app.Context{VM: vm, Obj: appCtx}
	nmgr, err := notification.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("notification.NewManager: %w", err)
	}
	defer nmgr.Close()

	// Create and register notification channel.
	ch, err := notification.NewChannel(
		vm, notificationChannelID, notificationChannelName,
		int32(notification.ImportanceHigh),
	)
	if err != nil {
		return fmt.Errorf("new channel: %w", err)
	}
	if err := nmgr.CreateNotificationChannel(ch.Obj); err != nil {
		return fmt.Errorf("create channel: %w", err)
	}

	// Build notification via typed Builder wrapper.
	b, err := notification.NewBuilder(vm, appCtx, notificationChannelID)
	if err != nil {
		return fmt.Errorf("new builder: %w", err)
	}
	if _, err := b.SetSmallIcon1_1(iconDialogInfo); err != nil {
		return fmt.Errorf("set icon: %w", err)
	}
	if _, err := b.SetContentTitle("Alarm fired!"); err != nil {
		return fmt.Errorf("set title: %w", err)
	}
	if _, err := b.SetContentText("From Go via JNI, zero Java"); err != nil {
		return fmt.Errorf("set text: %w", err)
	}
	notif, err := b.Build()
	if err != nil {
		return fmt.Errorf("build notification: %w", err)
	}
	return nmgr.Notify2(1, notif)
}

func playAlarmRingtone(vm *jni.VM, appCtx *jni.Object) error {
	rm := ringtone.Manager{VM: vm}

	uri, err := rm.GetDefaultUri(int32(ringtone.TypeAlarm))
	if err != nil {
		return fmt.Errorf("get alarm URI: %w", err)
	}
	if uri == nil || uri.Ref() == 0 {
		// Fall back to notification ringtone if no alarm ringtone is set.
		uri, err = rm.GetDefaultUri(int32(ringtone.TypeNotification))
		if err != nil {
			return fmt.Errorf("get notification URI: %w", err)
		}
	}

	rtObj, err := rm.GetRingtone2((*jni.Object)(appCtx), uri)
	if err != nil {
		return fmt.Errorf("getRingtone: %w", err)
	}
	if rtObj == nil || rtObj.Ref() == 0 {
		return fmt.Errorf("getRingtone returned nil")
	}

	var rt ringtone.Ringtone
	rt.VM = vm
	vm.Do(func(env *jni.Env) error {
		rt.Obj = env.NewGlobalRef(rtObj)
		return nil
	})

	if err := rt.Play(); err != nil {
		return fmt.Errorf("play: %w", err)
	}

	go func() {
		time.Sleep(ringtoneDuration)
		_ = rt.Stop()
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(rt.Obj)
			return nil
		})
	}()

	return nil
}
