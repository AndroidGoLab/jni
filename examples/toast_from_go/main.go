//go:build android

// Command toast_from_go demonstrates showing Toast messages from pure Go.
// It uses the generated toast package to create both short and long duration
// toasts via the Toast.makeText static factory and the typed wrapper API.
package main

/*
#include <android/native_activity.h>
extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static void _setCallbacks(ANativeActivity* a) { a->callbacks->onResume = _onResume; a->callbacks->onNativeWindowCreated = _onWindowCreated; }
static uintptr_t _getVM(ANativeActivity* a) { return (uintptr_t)a->vm; }
static uintptr_t _getClazz(ANativeActivity* a) { return (uintptr_t)a->clazz; }
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/widget/toast"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(C._getVM(activity))),
		jni.ObjectFromUintptr(uintptr(C._getClazz(activity))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(C._getClazz(activity))),
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

	// --- Duration constants ---
	fmt.Fprintln(output, "=== Toast Duration Constants ===")
	fmt.Fprintf(output, "LENGTH_SHORT = %d\n", toast.LengthShort)
	fmt.Fprintf(output, "LENGTH_LONG  = %d\n", toast.LengthLong)

	// --- Show a short toast ---
	// Use the static MakeText3_1 method via the typed wrapper.
	// Note: Toast.makeText and show() require a Looper-prepared thread
	// (the UI thread). Since NativeActivity examples may run on a
	// background thread, this may fail gracefully.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Short Toast ===")

	helper := toast.Toast{VM: vm}
	shortObj, err := helper.MakeText3_1(ctx.Obj, "Hello from Go (short)!", int32(toast.LengthShort))
	if err != nil {
		fmt.Fprintf(output, "makeText (short): %v\n", err)
	} else {
		shortToast := toast.Toast{VM: vm, Obj: shortObj}
		dur, err := shortToast.GetDuration()
		if err != nil {
			fmt.Fprintf(output, "getDuration: %v\n", err)
		} else {
			fmt.Fprintf(output, "duration: %d (LENGTH_SHORT)\n", dur)
		}

		if err := shortToast.Show(); err != nil {
			fmt.Fprintf(output, "show (short): %v\n", err)
		} else {
			fmt.Fprintln(output, "short toast shown!")
		}
	}

	// --- Show a long toast ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Long Toast ===")

	longObj, err := helper.MakeText3_1(ctx.Obj, "Hello from Go (long)!", int32(toast.LengthLong))
	if err != nil {
		fmt.Fprintf(output, "makeText (long): %v\n", err)
	} else {
		longToast := toast.Toast{VM: vm, Obj: longObj}
		dur, err := longToast.GetDuration()
		if err != nil {
			fmt.Fprintf(output, "getDuration: %v\n", err)
		} else {
			fmt.Fprintf(output, "duration: %d (LENGTH_LONG)\n", dur)
		}

		if err := longToast.Show(); err != nil {
			fmt.Fprintf(output, "show (long): %v\n", err)
		} else {
			fmt.Fprintln(output, "long toast shown!")
		}
	}

	// --- Toast via constructor + setText ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Toast via Constructor ===")

	t, err := toast.NewToast(vm, ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "NewToast: %v\n", err)
	} else {
		if err := t.SetText1_1("Built with NewToast + setText"); err != nil {
			fmt.Fprintf(output, "setText: %v\n", err)
		}
		if err := t.SetDuration(int32(toast.LengthShort)); err != nil {
			fmt.Fprintf(output, "setDuration: %v\n", err)
		}
		if err := t.Show(); err != nil {
			fmt.Fprintf(output, "show: %v\n", err)
		} else {
			fmt.Fprintln(output, "constructor toast shown!")
		}
	}

	fmt.Fprintln(output, "\ntoast_from_go complete")
	return nil
}
