//go:build android

// Command preferences_config demonstrates Android SharedPreferences as a
// typed key-value store. It writes config values (string, int, bool, float,
// long), reads them back, and verifies correctness.
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
	"math"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/content/preferences"
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

	// Get SharedPreferences with a unique name, MODE_PRIVATE=0.
	spObj, err := ctx.GetSharedPreferences("go_jni_config_demo", 0)
	if err != nil {
		return fmt.Errorf("getSharedPreferences: %w", err)
	}

	sp := preferences.SharedPreferences{VM: vm, Obj: spObj}

	fmt.Fprintln(output, "=== Preferences Config ===")
	fmt.Fprintln(output)

	// --- Write phase ---
	fmt.Fprintln(output, "-- Writing config --")

	editorObj, err := sp.Edit()
	if err != nil {
		return fmt.Errorf("edit: %w", err)
	}
	editor := preferences.SharedPreferencesEditor{VM: vm, Obj: editorObj}

	type configEntry struct {
		key string
		fn  func() error
	}

	entries := []configEntry{
		{"app_name", func() error { _, err := editor.PutString("app_name", "GoJNI Config Demo"); return err }},
		{"version", func() error { _, err := editor.PutInt("version", 7); return err }},
		{"debug_mode", func() error { _, err := editor.PutBoolean("debug_mode", true); return err }},
		{"scale_factor", func() error { _, err := editor.PutFloat("scale_factor", 1.75); return err }},
		{"last_sync_ms", func() error { _, err := editor.PutLong("last_sync_ms", 1711700000000); return err }},
		{"theme", func() error { _, err := editor.PutString("theme", "dark"); return err }},
		{"max_retries", func() error { _, err := editor.PutInt("max_retries", 3); return err }},
		{"notifications", func() error { _, err := editor.PutBoolean("notifications", false); return err }},
	}

	for _, e := range entries {
		if err := e.fn(); err != nil {
			return fmt.Errorf("put %s: %w", e.key, err)
		}
		fmt.Fprintf(output, "  put: %s\n", e.key)
	}

	ok, err := editor.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	fmt.Fprintf(output, "commit: %v\n", ok)
	fmt.Fprintln(output)

	// --- Read phase ---
	fmt.Fprintln(output, "-- Reading config --")

	appName, err := sp.GetString("app_name", "")
	if err != nil {
		return fmt.Errorf("getString(app_name): %w", err)
	}
	fmt.Fprintf(output, "  app_name: %q\n", appName)

	version, err := sp.GetInt("version", 0)
	if err != nil {
		return fmt.Errorf("getInt(version): %w", err)
	}
	fmt.Fprintf(output, "  version: %d\n", version)

	debugMode, err := sp.GetBoolean("debug_mode", false)
	if err != nil {
		return fmt.Errorf("getBoolean(debug_mode): %w", err)
	}
	fmt.Fprintf(output, "  debug_mode: %v\n", debugMode)

	scaleFactor, err := sp.GetFloat("scale_factor", 0)
	if err != nil {
		return fmt.Errorf("getFloat(scale_factor): %w", err)
	}
	fmt.Fprintf(output, "  scale_factor: %.2f\n", scaleFactor)

	lastSync, err := sp.GetLong("last_sync_ms", 0)
	if err != nil {
		return fmt.Errorf("getLong(last_sync_ms): %w", err)
	}
	fmt.Fprintf(output, "  last_sync_ms: %d\n", lastSync)

	theme, err := sp.GetString("theme", "")
	if err != nil {
		return fmt.Errorf("getString(theme): %w", err)
	}
	fmt.Fprintf(output, "  theme: %q\n", theme)

	maxRetries, err := sp.GetInt("max_retries", 0)
	if err != nil {
		return fmt.Errorf("getInt(max_retries): %w", err)
	}
	fmt.Fprintf(output, "  max_retries: %d\n", maxRetries)

	notif, err := sp.GetBoolean("notifications", true)
	if err != nil {
		return fmt.Errorf("getBoolean(notifications): %w", err)
	}
	fmt.Fprintf(output, "  notifications: %v\n", notif)
	fmt.Fprintln(output)

	// --- Verify phase ---
	fmt.Fprintln(output, "-- Verification --")

	allOk := true
	check := func(name string, pass bool) {
		status := "OK"
		if !pass {
			status = "FAIL"
			allOk = false
		}
		fmt.Fprintf(output, "  %s: %s\n", name, status)
	}

	check("app_name", appName == "GoJNI Config Demo")
	check("version", version == 7)
	check("debug_mode", debugMode == true)
	check("scale_factor", math.Abs(float64(scaleFactor)-1.75) < 0.01)
	check("last_sync_ms", lastSync == 1711700000000)
	check("theme", theme == "dark")
	check("max_retries", maxRetries == 3)
	check("notifications", notif == false)
	fmt.Fprintln(output)

	if allOk {
		fmt.Fprintln(output, "All checks passed.")
	} else {
		fmt.Fprintln(output, "Some checks FAILED!")
	}

	// --- Key existence ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "-- Key existence --")

	has, err := sp.Contains("app_name")
	if err != nil {
		return fmt.Errorf("contains: %w", err)
	}
	fmt.Fprintf(output, "  has app_name: %v\n", has)

	hasMissing, err := sp.Contains("nonexistent_key")
	if err != nil {
		return fmt.Errorf("contains: %w", err)
	}
	fmt.Fprintf(output, "  has nonexistent_key: %v\n", hasMissing)

	// --- Cleanup: remove a key ---
	editorObj2, err := sp.Edit()
	if err != nil {
		return fmt.Errorf("edit: %w", err)
	}
	editor2 := preferences.SharedPreferencesEditor{VM: vm, Obj: editorObj2}

	if _, err := editor2.Remove("debug_mode"); err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	ok2, err := editor2.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	fmt.Fprintf(output, "  removed debug_mode, commit: %v\n", ok2)

	hasRemoved, err := sp.Contains("debug_mode")
	if err != nil {
		return fmt.Errorf("contains: %w", err)
	}
	fmt.Fprintf(output, "  has debug_mode (after remove): %v\n", hasRemoved)

	return nil
}
