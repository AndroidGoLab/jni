//go:build android

// Command display demonstrates the Android Display and WindowManager API.
// It is built as a c-shared library and packaged into an APK using the
// shared apk.mk infrastructure.
//
// This package wraps android.view.WindowManager, android.view.Display, and
// android.util.DisplayMetrics. All types in this package are unexported
// (windowManager, display, displayMetrics), so they are not directly usable
// from external code. This example documents their structure and methods
// for reference.
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
	_ "github.com/xaionaro-go/jni/view/display"
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
	_, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}

	fmt.Fprintln(&output, "The display package provides unexported JNI bindings for:")
	fmt.Fprintln(&output, "  - android.view.WindowManager")
	fmt.Fprintln(&output, "  - android.view.Display")
	fmt.Fprintln(&output, "  - android.util.DisplayMetrics")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "All types are unexported and used by higher-level wrappers.")

	// windowManager wraps android.view.WindowManager.
	//
	// Constructor (unexported):
	//   NewwindowManager(ctx *app.Context) (*windowManager, error)
	//     Obtains the WindowManager system service via "window".
	//
	// Methods (unexported):
	//   windowManager.getDefaultDisplay() (*jni.Object, error)
	//     Returns the default Display object.

	// display wraps android.view.Display.
	//
	// Methods (unexported):
	//   display.getRotation() int32
	//     Returns the rotation of the screen from its "natural" orientation.
	//
	//   display.getRefreshRate() float32
	//     Returns the refresh rate of the display in frames per second.
	//
	//   display.getMetrics(metrics *jni.Object)
	//     Fills the given DisplayMetrics with the display's metrics.
	//
	//   display.getRealMetrics(metrics *jni.Object)
	//     Fills the given DisplayMetrics with the real display metrics,
	//     including areas of the screen that are occupied by system decorations.

	// displayMetrics wraps android.util.DisplayMetrics.
	//
	// Constructor (unexported):
	//   NewdisplayMetrics(vm *jni.VM) (*displayMetrics, error)
	//     Creates a new DisplayMetrics instance.
	//
	// The underlying Java object has fields widthPixels, heightPixels, and
	// densityDpi, which can be read via JNI field access after calling
	// display.getMetrics() or display.getRealMetrics() to populate them.

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
