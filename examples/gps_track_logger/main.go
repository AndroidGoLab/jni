//go:build android

// Command gps_track_logger checks the GPS provider status and retrieves
// the last known GPS location. It formats the result as a GPX-style
// coordinate, using only typed wrappers for LocationManager.
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
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/location"
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

	mgr, err := location.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("location.NewManager: %w", err)
	}
	defer mgr.Close()

	enabled, err := mgr.IsProviderEnabled(location.GpsProvider)
	if err != nil {
		return fmt.Errorf("IsProviderEnabled: %w", err)
	}
	fmt.Fprintf(output, "GPS provider enabled: %v\n", enabled)

	if !enabled {
		fmt.Fprintln(output, "GPS provider is not enabled. Enable GPS and retry.")
		fmt.Fprintln(output, "GPS track logger example completed.")
		return nil
	}

	// --- GNSS hardware info ---
	fmt.Fprintln(output, "\n=== GNSS Hardware ===")
	hwModel, err := mgr.GetGnssHardwareModelName()
	if err != nil {
		fmt.Fprintf(output, "GNSS model: %v\n", err)
	} else {
		fmt.Fprintf(output, "GNSS model: %s\n", hwModel)
	}

	hwYear, err := mgr.GetGnssYearOfHardware()
	if err != nil {
		fmt.Fprintf(output, "GNSS year: %v\n", err)
	} else {
		fmt.Fprintf(output, "GNSS year: %d\n", hwYear)
	}

	// --- Last known GPS location ---
	fmt.Fprintln(output, "\n=== Last Known GPS Location ===")
	locObj, err := mgr.GetLastKnownLocation(location.GpsProvider)
	if err != nil {
		fmt.Fprintf(output, "GetLastKnownLocation: %v\n", err)
		fmt.Fprintln(output, "\nGPS track logger example completed.")
		return nil
	}
	if locObj == nil || locObj.Ref() == 0 {
		fmt.Fprintln(output, "No cached GPS location available.")
		fmt.Fprintln(output, "\nGPS track logger example completed.")
		return nil
	}

	var loc *location.ExtractedLocation
	vm.Do(func(env *jni.Env) error {
		loc, err = location.ExtractLocation(env, locObj)
		return err
	})
	if err != nil {
		return fmt.Errorf("extract location: %w", err)
	}

	t := time.UnixMilli(loc.Time).UTC().Format(time.RFC3339)
	fmt.Fprintln(output, "=== GPX Track Point ===")
	fmt.Fprintf(output, "  <trkpt lat=\"%.6f\" lon=\"%.6f\">\n", loc.Latitude, loc.Longitude)
	fmt.Fprintf(output, "    <ele>%.1f</ele>\n", loc.Altitude)
	fmt.Fprintf(output, "    <time>%s</time>\n", t)
	fmt.Fprintf(output, "  </trkpt>\n")

	fixAge := time.Since(time.UnixMilli(loc.Time))
	fmt.Fprintf(output, "\nFix age: %s\n", fixAge.Round(time.Second))
	fmt.Fprintf(output, "Accuracy: %.1f m\n", float64(loc.Accuracy))
	fmt.Fprintf(output, "Speed: %.1f m/s\n", float64(loc.Speed))
	fmt.Fprintf(output, "Bearing: %.1f deg\n", float64(loc.Bearing))

	fmt.Fprintln(output, "\nGPS track logger example completed.")
	return nil
}
