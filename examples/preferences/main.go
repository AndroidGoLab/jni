//go:build android

// Command preferences demonstrates the Android SharedPreferences API.
// It writes several typed values, reads them back, and displays the results.
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
	"github.com/AndroidGoLab/jni/content/preferences"
	"github.com/AndroidGoLab/jni/exampleui"
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
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// Context.getSharedPreferences("go_jni_demo", MODE_PRIVATE=0)
	spObj, err := ctx.GetSharedPreferences("go_jni_demo", 0)
	if err != nil {
		return fmt.Errorf("getSharedPreferences: %w", err)
	}

	sp := preferences.SharedPreferences{VM: vm, Obj: spObj}

	fmt.Fprintln(output, "=== SharedPreferences ===")
	fmt.Fprintln(output)

	// Write values via the editor.
	editorObj, err := sp.Edit()
	if err != nil {
		return fmt.Errorf("edit: %w", err)
	}
	editor := preferences.SharedPreferencesEditor{VM: vm, Obj: editorObj}

	if _, err := editor.PutString("greeting", "Hello from Go!"); err != nil {
		return fmt.Errorf("putString: %w", err)
	}
	if _, err := editor.PutInt("counter", 42); err != nil {
		return fmt.Errorf("putInt: %w", err)
	}
	if _, err := editor.PutBoolean("enabled", true); err != nil {
		return fmt.Errorf("putBoolean: %w", err)
	}
	if _, err := editor.PutFloat("ratio", 3.14); err != nil {
		return fmt.Errorf("putFloat: %w", err)
	}
	if _, err := editor.PutLong("timestamp", 1700000000); err != nil {
		return fmt.Errorf("putLong: %w", err)
	}

	ok, err := editor.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	fmt.Fprintf(output, "commit: %v\n", ok)
	fmt.Fprintln(output)

	// Read values back.
	greeting, err := sp.GetString("greeting", "")
	if err != nil {
		return fmt.Errorf("getString: %w", err)
	}
	fmt.Fprintf(output, "greeting: %q\n", greeting)

	counter, err := sp.GetInt("counter", 0)
	if err != nil {
		return fmt.Errorf("getInt: %w", err)
	}
	fmt.Fprintf(output, "counter: %d\n", counter)

	enabled, err := sp.GetBoolean("enabled", false)
	if err != nil {
		return fmt.Errorf("getBoolean: %w", err)
	}
	fmt.Fprintf(output, "enabled: %v\n", enabled)

	ratio, err := sp.GetFloat("ratio", 0)
	if err != nil {
		return fmt.Errorf("getFloat: %w", err)
	}
	fmt.Fprintf(output, "ratio: %.2f\n", ratio)

	ts, err := sp.GetLong("timestamp", 0)
	if err != nil {
		return fmt.Errorf("getLong: %w", err)
	}
	fmt.Fprintf(output, "timestamp: %d\n", ts)

	// Check key existence.
	fmt.Fprintln(output)
	has, err := sp.Contains("greeting")
	if err != nil {
		return fmt.Errorf("contains: %w", err)
	}
	fmt.Fprintf(output, "has greeting: %v\n", has)

	hasMissing, err := sp.Contains("nonexistent")
	if err != nil {
		return fmt.Errorf("contains: %w", err)
	}
	fmt.Fprintf(output, "has missing: %v\n", hasMissing)

	return nil
}
