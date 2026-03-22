//go:build android

// Command clipboard demonstrates the Android ClipboardManager API.
// It sets text on the clipboard, reads it back, and displays the result.
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
	"github.com/AndroidGoLab/jni/content/clipboard"
	"github.com/AndroidGoLab/jni/examples/common/ui"
)

// charSequenceToString calls CharSequence.toString() via JNI and returns the Go string.
func charSequenceToString(vm *jni.VM, csObj *jni.Object) string {
	if csObj == nil {
		return ""
	}
	var result string
	vm.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(csObj)
		mid, err := env.GetMethodID(cls, "toString", "()Ljava/lang/String;")
		if err != nil {
			return nil
		}
		strObj, err := env.CallObjectMethod(csObj, mid)
		if err != nil {
			return nil
		}
		result = env.GoString((*jni.String)(unsafe.Pointer(strObj)))
		return nil
	})
	return result
}

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(activity.vm)),
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
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

	mgr, err := clipboard.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("clipboard.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Clipboard Demo ===")
	fmt.Fprintln(output)

	// Check initial clipboard state.
	hasClip, err := mgr.HasPrimaryClip()
	if err != nil {
		return fmt.Errorf("hasPrimaryClip: %w", err)
	}
	fmt.Fprintf(output, "has clip: %v\n", hasClip)

	// Set text via the deprecated but simple setText API.
	const testText = "Hello from Go JNI!"
	if err := mgr.SetText(testText); err != nil {
		return fmt.Errorf("setText: %w", err)
	}
	fmt.Fprintf(output, "set: %q\n", testText)

	// Read it back. GetText returns a CharSequence (*jni.Object), so we
	// call toString() to get a Go string.
	gotObj, err := mgr.GetText()
	if err != nil {
		return fmt.Errorf("getText: %w", err)
	}
	got := charSequenceToString(vm, gotObj)
	fmt.Fprintf(output, "got: %q\n", got)

	// Verify round-trip.
	fmt.Fprintln(output)
	if got == testText {
		fmt.Fprintln(output, "round-trip OK")
	} else {
		fmt.Fprintln(output, "MISMATCH!")
	}

	// Confirm clip is present after write.
	hasClip, err = mgr.HasPrimaryClip()
	if err != nil {
		return fmt.Errorf("hasPrimaryClip: %w", err)
	}
	fmt.Fprintf(output, "has clip: %v\n", hasClip)

	return nil
}
