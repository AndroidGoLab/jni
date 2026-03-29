//go:build android

// Command test_harness is an automated test concept: calls multiple API
// endpoints in sequence, verifies each returns valid data, and reports
// pass/fail for each. Tests: build info, battery, display, environment,
// storage, keyguard, power.
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
	"github.com/AndroidGoLab/jni/os/battery"
	"github.com/AndroidGoLab/jni/os/build"
	"github.com/AndroidGoLab/jni/os/keyguard"
	"github.com/AndroidGoLab/jni/os/power"
	"github.com/AndroidGoLab/jni/os/storage"
	"github.com/AndroidGoLab/jni/view/display"
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

type testResult struct {
	name   string
	passed bool
	detail string
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Test Harness ===")

	var results []testResult

	// Test 1: Build Info
	func() {
		info, err := build.GetBuildInfo(vm)
		if err != nil {
			results = append(results, testResult{"BuildInfo", false, err.Error()})
			return
		}
		if info.Model == "" {
			results = append(results, testResult{"BuildInfo", false, "empty model"})
			return
		}
		results = append(results, testResult{"BuildInfo", true, info.Model})
	}()

	// Test 2: Version Info
	func() {
		ver, err := build.GetVersionInfo(vm)
		if err != nil {
			results = append(results, testResult{"VersionInfo", false, err.Error()})
			return
		}
		if ver.SDKInt <= 0 {
			results = append(results, testResult{"VersionInfo", false, "invalid SDK"})
			return
		}
		results = append(results, testResult{"VersionInfo", true, fmt.Sprintf("SDK=%d", ver.SDKInt)})
	}()

	// Test 3: Battery
	func() {
		mgr, err := battery.NewManager(ctx)
		if err != nil {
			results = append(results, testResult{"Battery", false, err.Error()})
			return
		}
		defer mgr.Close()
		level, err := mgr.GetIntProperty(int32(battery.BatteryPropertyCapacity))
		if err != nil {
			results = append(results, testResult{"Battery", false, err.Error()})
			return
		}
		if level < 0 || level > 100 {
			results = append(results, testResult{"Battery", false, fmt.Sprintf("bad level %d", level)})
			return
		}
		results = append(results, testResult{"Battery", true, fmt.Sprintf("%d%%", level)})
	}()

	// Test 4: Display
	func() {
		wm, err := display.NewWindowManager(ctx)
		if err != nil {
			results = append(results, testResult{"Display", false, err.Error()})
			return
		}
		defer wm.Close()
		dispObj, err := wm.GetDefaultDisplay()
		if err != nil {
			results = append(results, testResult{"Display", false, err.Error()})
			return
		}
		if dispObj == nil || dispObj.Ref() == 0 {
			results = append(results, testResult{"Display", false, "null display"})
			return
		}
		disp := display.Display{VM: vm, Obj: dispObj}
		w, _ := disp.GetWidth()
		h, _ := disp.GetHeight()
		if w <= 0 || h <= 0 {
			results = append(results, testResult{"Display", false, fmt.Sprintf("bad size %dx%d", w, h)})
			return
		}
		results = append(results, testResult{"Display", true, fmt.Sprintf("%dx%d", w, h)})
	}()

	// Test 5: Storage
	func() {
		mgr, err := storage.NewManager(ctx)
		if err != nil {
			results = append(results, testResult{"Storage", false, err.Error()})
			return
		}
		defer mgr.Close()
		results = append(results, testResult{"Storage", true, "service obtained"})
	}()

	// Test 6: Keyguard
	func() {
		mgr, err := keyguard.NewManager(ctx)
		if err != nil {
			results = append(results, testResult{"Keyguard", false, err.Error()})
			return
		}
		defer mgr.Close()

		locked, err := mgr.IsKeyguardLocked()
		if err != nil {
			results = append(results, testResult{"Keyguard", false, err.Error()})
			return
		}
		results = append(results, testResult{"Keyguard", true, fmt.Sprintf("locked=%v", locked)})
	}()

	// Test 7: Power
	func() {
		mgr, err := power.NewManager(ctx)
		if err != nil {
			results = append(results, testResult{"Power", false, err.Error()})
			return
		}
		defer mgr.Close()

		interactive, err := mgr.IsInteractive()
		if err != nil {
			results = append(results, testResult{"Power", false, err.Error()})
			return
		}
		results = append(results, testResult{"Power", true, fmt.Sprintf("interactive=%v", interactive)})
	}()

	// Print results.
	passed := 0
	failed := 0
	for _, r := range results {
		status := "PASS"
		if !r.passed {
			status = "FAIL"
			failed++
		} else {
			passed++
		}
		fmt.Fprintf(output, "  [%s] %-12s %s\n", status, r.name, r.detail)
	}

	fmt.Fprintf(output, "\nResults: %d passed, %d failed, %d total\n", passed, failed, len(results))
	fmt.Fprintln(output, "\nTest harness complete.")
	return nil
}
