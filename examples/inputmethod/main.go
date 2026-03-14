//go:build android

// Command inputmethod demonstrates the InputMethodManager JNI bindings.
// It is built as a c-shared library and packaged into an APK using the
// shared apk.mk infrastructure.
//
// This example obtains the InputMethodManager system service. The
// package wraps android.view.inputmethod.InputMethodManager and provides
// methods for programmatically showing and hiding the soft keyboard:
//   - showSoftInput(view, flags) - show the soft keyboard for a view
//   - hideSoftInputFromWindow(token, flags) - hide the soft keyboard
//
// These methods are package-internal and intended to be composed into
// higher-level APIs. NewManager and Close are the exported entry points.
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
	"github.com/xaionaro-go/jni/view/inputmethod"
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

	mgr, err := inputmethod.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("inputmethod.NewManager: %v", err)
	}
	defer mgr.Close()

	fmt.Fprintln(&output, "InputMethodManager obtained")

	// The manager wraps android.view.inputmethod.InputMethodManager.
	//
	// Package-internal methods (for composition into higher-level APIs):
	//   showSoftInput(view *jni.Object, flags int32) (bool, error)
	//     Shows the soft keyboard for the given View.
	//     flags=0 means no special behavior.
	//
	//   hideSoftInputFromWindow(windowToken *jni.Object, flags int32) (bool, error)
	//     Hides the soft keyboard using the window token obtained from
	//     view.getWindowToken(). flags=0 means no special behavior.
	fmt.Fprintln(&output, "InputMethodManager ready for keyboard control")

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
