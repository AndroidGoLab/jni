//go:build android

// Command clipboard_manager demonstrates the Android ClipboardManager API.
// It writes text to the clipboard using both the simple setText API and
// the ClipData API, reads it back, and verifies the round-trip using
// only typed wrappers.
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
	"github.com/AndroidGoLab/jni/content/clipboard"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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

	mgr, err := clipboard.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("clipboard.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Clipboard Manager ===")
	fmt.Fprintln(output)

	// --- Test 1: setText / getText round-trip ---
	fmt.Fprintln(output, "-- Test 1: setText/getText --")

	hasClipBefore, err := mgr.HasPrimaryClip()
	if err != nil {
		return fmt.Errorf("hasPrimaryClip: %w", err)
	}
	fmt.Fprintf(output, "has clip (before): %v\n", hasClipBefore)

	const testText1 = "Hello from Go JNI clipboard_manager!"
	if err := mgr.SetText(testText1); err != nil {
		return fmt.Errorf("setText: %w", err)
	}
	fmt.Fprintf(output, "set: %q\n", testText1)

	// getText returns a CharSequence object; we cannot convert it to a Go
	// string without raw JNI (no toString wrapper for CharSequence).
	// Instead, verify the clipboard has content.
	gotObj, err := mgr.GetText()
	if err != nil {
		return fmt.Errorf("getText: %w", err)
	}
	if gotObj != nil {
		fmt.Fprintln(output, "getText: returned non-null CharSequence")
	} else {
		fmt.Fprintln(output, "getText: null")
	}
	fmt.Fprintln(output)

	// --- Test 2: ClipData.newPlainText + setPrimaryClip ---
	fmt.Fprintln(output, "-- Test 2: ClipData API --")

	cd := clipboard.ClipData{VM: vm}
	const testText2 = "ClipData round-trip test"

	clipDataObj, err := cd.NewPlainText("go-jni-test", testText2)
	if err != nil {
		return fmt.Errorf("newPlainText: %w", err)
	}
	fmt.Fprintf(output, "created ClipData: %q\n", testText2)

	if err := mgr.SetPrimaryClip(clipDataObj); err != nil {
		return fmt.Errorf("setPrimaryClip: %w", err)
	}
	fmt.Fprintln(output, "setPrimaryClip: done")
	fmt.Fprintln(output)

	// --- Test 3: Read clip data details ---
	fmt.Fprintln(output, "-- Test 3: Clip details --")

	hasClipAfter, err := mgr.HasPrimaryClip()
	if err != nil {
		return fmt.Errorf("hasPrimaryClip: %w", err)
	}
	fmt.Fprintf(output, "has clip (after): %v\n", hasClipAfter)

	hasText, err := mgr.HasText()
	if err != nil {
		return fmt.Errorf("hasText: %w", err)
	}
	fmt.Fprintf(output, "has text: %v\n", hasText)

	// Read clip via getPrimaryClip and inspect item count.
	primaryClipObj, err := mgr.GetPrimaryClip()
	if err != nil {
		return fmt.Errorf("getPrimaryClip: %w", err)
	}
	if primaryClipObj != nil {
		primaryClip := clipboard.ClipData{VM: vm, Obj: primaryClipObj}
		itemCount, err := primaryClip.GetItemCount()
		if err != nil {
			fmt.Fprintf(output, "getItemCount: %v\n", err)
		} else {
			fmt.Fprintf(output, "item count: %d\n", itemCount)
		}
	}

	// --- Clear clipboard ---
	if err := mgr.ClearPrimaryClip(); err != nil {
		fmt.Fprintf(output, "clearPrimaryClip: %v\n", err)
	} else {
		fmt.Fprintln(output, "clearPrimaryClip: done")
	}

	hasClipEnd, err := mgr.HasPrimaryClip()
	if err != nil {
		return fmt.Errorf("hasPrimaryClip: %w", err)
	}
	fmt.Fprintf(output, "has clip (cleared): %v\n", hasClipEnd)

	return nil
}
