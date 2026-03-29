//go:build android

// Command location_passive uses the passive location provider to get a
// location without requesting new fixes. The passive provider piggybacks
// on location requests made by other apps.
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

	fmt.Fprintln(output, "=== Passive Location Provider ===")
	fmt.Fprintf(output, "Provider constant: %q\n", location.PassiveProvider)

	// Check passive provider status.
	passiveEnabled, err := mgr.IsProviderEnabled(location.PassiveProvider)
	if err != nil {
		fmt.Fprintf(output, "IsProviderEnabled: %v\n", err)
	} else {
		fmt.Fprintf(output, "Passive provider enabled: %v\n", passiveEnabled)
	}

	// Also check GPS/network for comparison.
	fmt.Fprintln(output, "\n=== All Provider Status ===")
	for _, p := range []string{location.GpsProvider, location.NetworkProvider, location.PassiveProvider} {
		enabled, err := mgr.IsProviderEnabled(p)
		if err != nil {
			fmt.Fprintf(output, "  %s: error: %v\n", p, err)
		} else {
			fmt.Fprintf(output, "  %s: enabled=%v\n", p, enabled)
		}
	}

	// Get last known location from passive provider.
	fmt.Fprintln(output, "\n=== Passive Location ===")
	locObj, err := mgr.GetLastKnownLocation(location.PassiveProvider)
	if err != nil {
		fmt.Fprintf(output, "GetLastKnownLocation(passive): %v\n", err)
	} else if locObj == nil || locObj.Ref() == 0 {
		fmt.Fprintln(output, "No cached passive location available.")
		fmt.Fprintln(output, "(Passive provider reuses fixes from other apps.)")
		fmt.Fprintln(output, "Try running a maps app first, then retry.")
	} else {
		var loc *location.ExtractedLocation
		vm.Do(func(env *jni.Env) error {
			loc, err = location.ExtractLocation(env, locObj)
			return err
		})
		if err != nil {
			fmt.Fprintf(output, "ExtractLocation: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Provider:  %s\n", loc.Provider)
			fmt.Fprintf(output, "  Latitude:  %.6f\n", loc.Latitude)
			fmt.Fprintf(output, "  Longitude: %.6f\n", loc.Longitude)
			fmt.Fprintf(output, "  Altitude:  %.1f m\n", loc.Altitude)
			fmt.Fprintf(output, "  Accuracy:  %.1f m\n", float64(loc.Accuracy))

			fixAge := time.Since(time.UnixMilli(loc.Time))
			fmt.Fprintf(output, "  Fix time:  %s\n",
				time.UnixMilli(loc.Time).UTC().Format(time.RFC3339))
			fmt.Fprintf(output, "  Fix age:   %s\n", fixAge.Round(time.Second))
		}
	}

	// Compare with GPS last known for reference.
	fmt.Fprintln(output, "\n=== GPS Last Known (for comparison) ===")
	gpsObj, err := mgr.GetLastKnownLocation(location.GpsProvider)
	if err != nil {
		fmt.Fprintf(output, "GetLastKnownLocation(gps): %v\n", err)
	} else if gpsObj == nil || gpsObj.Ref() == 0 {
		fmt.Fprintln(output, "No cached GPS location.")
	} else {
		var gps *location.ExtractedLocation
		vm.Do(func(env *jni.Env) error {
			gps, err = location.ExtractLocation(env, gpsObj)
			return err
		})
		if err == nil {
			fmt.Fprintf(output, "  lat=%.6f lon=%.6f (acc=%.1f m)\n",
				gps.Latitude, gps.Longitude, float64(gps.Accuracy))
		}
	}

	fmt.Fprintln(output, "\nPassive location example completed.")
	return nil
}
