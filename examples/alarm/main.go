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
	ctx, err := getAppContext(vm)
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

	// Schedule the alarm inside a single vm.Do so all JNI local refs
	// (Intent, PendingIntent) remain valid on the same thread.
	err = vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}

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

		return mgr.Set(int32(alarm.RtcWakeup), triggerAt.UnixMilli(), pendingObj)
	})
	if err != nil {
		return fmt.Errorf("schedule alarm: %w", err)
	}
	fmt.Fprintf(output, "alarm scheduled for %s\n", triggerAt.Format(time.RFC3339))

	var appCtxRef *jni.GlobalRef
	err = vm.Do(func(env *jni.Env) error {
		appCtxObj, err := ctx.GetApplicationContext()
		if err != nil {
			return err
		}
		appCtxRef = env.NewGlobalRef(appCtxObj)
		return nil
	})
	if err != nil {
		return fmt.Errorf("get app context: %w", err)
	}

	fmt.Fprintf(output, "alarm will fire in %s\n", delay)
	go func() {
		time.Sleep(delay)
		if err := postNotification(vm, appCtxRef); err != nil {
			fmt.Fprintf(output, "notification error: %v\n", err)
		}
		if err := playAlarmRingtone(vm, appCtxRef); err != nil {
			fmt.Fprintf(output, "ringtone error: %v\n", err)
		}
		fmt.Fprintf(output, "alarm fired!\n")
		exampleui.RenderOutput()
	}()

	return nil
}

func postNotification(vm *jni.VM, appCtx *jni.GlobalRef) error {
	ctx := &app.Context{VM: vm, Obj: appCtx}
	nmgr, err := notification.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("notification.NewManager: %w", err)
	}
	defer nmgr.Close()

	return vm.Do(func(env *jni.Env) error {
		// Create and register notification channel.
		channelCls, err := env.FindClass("android/app/NotificationChannel")
		if err != nil {
			return err
		}
		initMid, err := env.GetMethodID(channelCls, "<init>",
			"(Ljava/lang/String;Ljava/lang/CharSequence;I)V")
		if err != nil {
			return err
		}
		jID, _ := env.NewStringUTF(notificationChannelID)
		jName, _ := env.NewStringUTF(notificationChannelName)
		ch, err := env.NewObject(channelCls, initMid,
			jni.ObjectValue(&jID.Object),
			jni.ObjectValue(&jName.Object),
			jni.IntValue(int32(notification.ImportanceHigh)))
		if err != nil {
			return err
		}
		if err := nmgr.CreateNotificationChannel(ch); err != nil {
			return err
		}

		// Build notification.
		bCls, err := env.FindClass("android/app/Notification$Builder")
		if err != nil {
			return err
		}
		bInit, err := env.GetMethodID(bCls, "<init>",
			"(Landroid/content/Context;Ljava/lang/String;)V")
		if err != nil {
			return err
		}
		b, err := env.NewObject(bCls, bInit,
			jni.ObjectValue(appCtx), jni.ObjectValue(&jID.Object))
		if err != nil {
			return err
		}

		setIcon, _ := env.GetMethodID(bCls, "setSmallIcon",
			"(I)Landroid/app/Notification$Builder;")
		_, _ = env.CallObjectMethod(b, setIcon, jni.IntValue(iconDialogInfo))

		setTitle, _ := env.GetMethodID(bCls, "setContentTitle",
			"(Ljava/lang/CharSequence;)Landroid/app/Notification$Builder;")
		jT, _ := env.NewStringUTF("Alarm fired!")
		_, _ = env.CallObjectMethod(b, setTitle, jni.ObjectValue(&jT.Object))

		setText, _ := env.GetMethodID(bCls, "setContentText",
			"(Ljava/lang/CharSequence;)Landroid/app/Notification$Builder;")
		jTx, _ := env.NewStringUTF("From Go via JNI, zero Java")
		_, _ = env.CallObjectMethod(b, setText, jni.ObjectValue(&jTx.Object))

		buildMid, _ := env.GetMethodID(bCls, "build",
			"()Landroid/app/Notification;")
		notif, err := env.CallObjectMethod(b, buildMid)
		if err != nil {
			return err
		}
		return nmgr.Notify2(1, notif)
	})
}

func playAlarmRingtone(vm *jni.VM, appCtx *jni.GlobalRef) error {
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

func getAppContext(vm *jni.VM) (*app.Context, error) {
	var ctx app.Context
	ctx.VM = vm
	err := vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}
		atClass, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return err
		}
		mid, err := env.GetStaticMethodID(atClass, "currentApplication", "()Landroid/app/Application;")
		if err != nil {
			return err
		}
		appObj, err := env.CallStaticObjectMethod(atClass, mid)
		if err != nil {
			return err
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
