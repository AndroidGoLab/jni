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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/app"
	"github.com/xaionaro-go/jni/os/power"
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

	// --- Constants ---
	fmt.Fprintf(&output, "PartialWakeLock = %d\n", power.PartialWakeLock)

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
	fmt.Fprintf(&output, "screen interactive: %v\n", interactive)

	// --- IsPowerSaveMode ---
	powerSave, err := mgr.IsPowerSaveMode()
	if err != nil {
		return fmt.Errorf("IsPowerSaveMode: %v", err)
	}
	fmt.Fprintf(&output, "power save mode: %v\n", powerSave)

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

	fmt.Fprintln(&output, "Power example complete.")
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
