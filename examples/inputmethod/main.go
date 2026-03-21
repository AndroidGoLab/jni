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
	"github.com/AndroidGoLab/jni/view/inputmethod"
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

	mgr, err := inputmethod.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("inputmethod.NewManager: %v", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "InputMethodManager obtained")

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
	fmt.Fprintln(output, "InputMethodManager ready for keyboard control")

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
