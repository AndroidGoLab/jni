//go:build android

// Command power demonstrates using the PowerManager API. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the PowerManager system service, queries the
// interactive and power-save states, and shows the PartialWakeLock
// constant. Wake locks are created and managed via unexported methods
// that wrap the JNI layer; the WakeLock.Release method is exported.
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
	"github.com/AndroidGoLab/jni/os/power"
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

	// --- Constants ---
	fmt.Fprintf(output, "PartialWakeLock = %d\n", power.PartialWakeLock)

	// --- NewManager ---
	mgr, err := power.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("power.NewManager: %v", err)
	}

	// --- IsInteractive ---
	interactive, err := mgr.IsInteractive()
	if err != nil {
		return fmt.Errorf("IsInteractive: %v", err)
	}
	fmt.Fprintf(output, "screen interactive: %v\n", interactive)

	// --- IsPowerSaveMode ---
	powerSave, err := mgr.IsPowerSaveMode()
	if err != nil {
		return fmt.Errorf("IsPowerSaveMode: %v", err)
	}
	fmt.Fprintf(output, "power save mode: %v\n", powerSave)

	// --- WakeLock ---
	// Wake locks are created via mgr.newWakeLock (unexported):
	//   mgr.newWakeLock(flags int32, tag string) (*jni.Object, error)
	//
	// Use power.PartialWakeLock as the flags argument to keep the CPU
	// running while allowing the screen to turn off.
	//
	// The WakeLock type provides:
	//   wl.acquireMs(timeoutMs int64) error  [unexported]
	//   wl.Release() error                   [exported]
	//
	// Always acquire with a timeout to prevent battery drain, and
	// release when the background work is done.

	fmt.Fprintln(output, "Power example complete.")
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
