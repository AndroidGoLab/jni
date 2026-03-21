//go:build android

// Command battery demonstrates Android battery status constants. It is
// built as a c-shared library and packaged into an APK using the
// shared apk.mk infrastructure.
//
// Battery information is obtained from the sticky broadcast
// android.intent.action.BATTERY_CHANGED via Intent extras. The battery
// package provides typed constants for interpreting those extras.
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
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/os/battery"
)

func main() {}

func init() { exampleui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
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

	// Battery status constants (android.os.BatteryManager.BATTERY_STATUS_*).
	// The Status type is a named int for type safety.
	fmt.Fprintf(output, "status: unknown=%d, charging=%d, discharging=%d, not_charging=%d, full=%d\n",
		battery.BatteryStatusUnknown, battery.BatteryStatusCharging, battery.BatteryStatusDischarging,
		battery.BatteryStatusNotCharging, battery.BatteryStatusFull)

	// Plugged state constants (android.os.BatteryManager.BATTERY_PLUGGED_*).
	fmt.Fprintf(output, "plugged: ac=%d, usb=%d, wireless=%d\n",
		battery.BatteryPluggedAc, battery.BatteryPluggedUsb,
		battery.BatteryPluggedWireless)

	// In a real app, battery info is read from the BATTERY_CHANGED sticky
	// broadcast. Register a BroadcastReceiver via app.Context.RegisterReceiverRaw
	// or read the sticky intent directly:
	//
	//   intent, _ := ctx.RegisterReceiverRaw(nil, batteryFilter)
	//   status := battery.Status(intent.GetIntExtra("status", 0))
	//   level := intent.GetIntExtra("level", 0)
	//   scale := intent.GetIntExtra("scale", 100)
	//   pct := float64(level) / float64(scale) * 100
	//
	// Available extras: level, scale, status, temperature, voltage, plugged,
	// health, technology, present.

	// Demonstrate Status type usage.
	status := battery.BatteryStatusCharging
	switch status {
	case battery.BatteryStatusUnknown:
		fmt.Fprintln(output, "battery status: unknown")
	case battery.BatteryStatusCharging:
		fmt.Fprintln(output, "battery status: charging")
	case battery.BatteryStatusDischarging:
		fmt.Fprintln(output, "battery status: discharging")
	case battery.BatteryStatusNotCharging:
		fmt.Fprintln(output, "battery status: not charging")
	case battery.BatteryStatusFull:
		fmt.Fprintln(output, "battery status: full")
	}

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
