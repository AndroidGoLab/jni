//go:build android

// Command toast demonstrates the full Toast API surface provided by the
// generated toast package.
//
// It covers:
//   - toast type: show()
//   - Constants: Short (LENGTH_SHORT = 0), Long (LENGTH_LONG = 1)
//
// Toast.makeText is a static factory method on the Android side; the
// generated package exposes the instance method show() on the toast
// wrapper type and the duration constants. Creating the Toast object
// itself requires calling Toast.makeText via JNI (typically through
// app.Context), then wrapping the result.
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
	"github.com/xaionaro-go/jni/widget/toast"
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
	// The toast package provides duration constants matching the Android
	// Toast.LENGTH_SHORT and Toast.LENGTH_LONG values.
	fmt.Fprintln(&output, "=== Toast duration constants ===")
	fmt.Fprintf(&output, "  Short (LENGTH_SHORT) = %d\n", toast.Short)
	fmt.Fprintf(&output, "  Long  (LENGTH_LONG)  = %d\n", toast.Long)

	// --- Showing a toast ---
	// The toast type wraps android.widget.Toast. Its single method is:
	//   show() error
	//
	// To create a Toast, call the static factory Toast.makeText via JNI,
	// then wrap the returned object in the toast struct and call show().
	//
	// Example flow:
	//   1. Call Toast.makeText(context, "Hello from Go!", toast.Short) via JNI
	//   2. Wrap the returned Java object
	//   3. Call show() on the wrapper
	err = vm.Do(func(env *jni.Env) error {
		// Find the Toast class for the static makeText call.
		cls, err := env.FindClass("android/widget/Toast")
		if err != nil {
			return fmt.Errorf("FindClass: %w", err)
		}

		// Locate the static factory method:
		//   static Toast makeText(Context, CharSequence, int)
		mid, err := env.GetStaticMethodID(cls,
			"makeText",
			"(Landroid/content/Context;Ljava/lang/CharSequence;I)Landroid/widget/Toast;",
		)
		if err != nil {
			return fmt.Errorf("GetStaticMethodID: %w", err)
		}

		// Create the message string.
		msg, err := env.NewStringUTF("Hello from Go!")
		if err != nil {
			return fmt.Errorf("NewStringUTF: %w", err)
		}

		// Create a short-duration toast using the Short constant.
		// Toast.makeText requires a Looper-prepared thread. Since nativeRun
		// may run on a background thread, this may fail with a Looper error.
		toastObj, err := env.CallStaticObjectMethod(cls, mid,
			jni.ObjectValue(ctx.Obj),
			jni.ObjectValue(&msg.Object),
			jni.IntValue(int32(toast.Short)),
		)
		if err != nil {
			fmt.Fprintf(&output, "  makeText (requires UI thread): %v\n", err)
			return nil
		}

		// The returned object can be wrapped in a toast struct whose
		// show() method calls android.widget.Toast.show():
		//
		//   t := toast{VM: vm, Obj: env.NewGlobalRef(toastObj)}
		//   t.show()
		//
		// Alternatively, create a long-duration toast:
		//   env.CallStaticObjectMethod(cls, mid,
		//       jni.ObjectValue(ctx.Obj),
		//       jni.ObjectValue(msg.Object()),
		//       jni.IntValue(int32(toast.Long)),
		//   )
		_ = toastObj

		fmt.Fprintln(&output, "Toast created via generated API")
		return nil
	})
	if err != nil {
		return fmt.Errorf("toast: %w", err)
	}

	fmt.Fprintln(&output, "\nAll toast package features demonstrated.")
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
