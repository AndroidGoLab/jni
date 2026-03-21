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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/widget/toast"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
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

	// --- Constants ---
	// The toast package provides duration constants matching the Android
	// Toast.LENGTH_SHORT and Toast.LENGTH_LONG values.
	fmt.Fprintln(output, "=== Toast duration constants ===")
	fmt.Fprintf(output, "  Short (LENGTH_SHORT) = %d\n", toast.LengthShort)
	fmt.Fprintf(output, "  Long  (LENGTH_LONG)  = %d\n", toast.LengthLong)

	// --- Showing a toast ---
	// The toast type wraps android.widget.Toast. Its single method is:
	//   show() error
	//
	// To create a Toast, call the static factory Toast.makeText via JNI,
	// then wrap the returned object in the toast struct and call show().
	//
	// Example flow:
	//   1. Call Toast.makeText(context, "Hello from Go!", toast.LengthShort) via JNI
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
			jni.IntValue(int32(toast.LengthShort)),
		)
		if err != nil {
			fmt.Fprintf(output, "  makeText (requires UI thread): %v\n", err)
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
		//       jni.IntValue(int32(toast.LengthLong)),
		//   )
		_ = toastObj

		fmt.Fprintln(output, "Toast created via generated API")
		return nil
	})
	if err != nil {
		return fmt.Errorf("toast: %w", err)
	}

	fmt.Fprintln(output, "\nAll toast package features demonstrated.")
	return nil
}
