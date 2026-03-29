//go:build android

// Command gps_speed_altitude gets a location with speed and altitude data
// and displays GPS speed (m/s and km/h) and altitude from the Location
// object fields.
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

	fmt.Fprintln(output, "=== GPS Speed & Altitude ===")

	// Check GPS provider status.
	gpsEnabled, err := mgr.IsProviderEnabled(location.GpsProvider)
	if err != nil {
		fmt.Fprintf(output, "IsProviderEnabled(gps): %v\n", err)
	} else {
		fmt.Fprintf(output, "GPS enabled: %v\n", gpsEnabled)
	}

	// Try each provider for a location with speed/altitude data.
	providers := []string{location.GpsProvider, location.NetworkProvider, location.PassiveProvider}
	var loc *location.ExtractedLocation

	fmt.Fprintln(output, "\n=== Querying providers ===")
	for _, provider := range providers {
		enabled, _ := mgr.IsProviderEnabled(provider)
		if !enabled {
			fmt.Fprintf(output, "  %s: not enabled\n", provider)
			continue
		}
		locObj, err := mgr.GetLastKnownLocation(provider)
		if err != nil {
			fmt.Fprintf(output, "  %s: %v\n", provider, err)
			continue
		}
		if locObj == nil || locObj.Ref() == 0 {
			fmt.Fprintf(output, "  %s: no cached location\n", provider)
			continue
		}
		var extracted *location.ExtractedLocation
		vm.Do(func(env *jni.Env) error {
			extracted, err = location.ExtractLocation(env, locObj)
			return err
		})
		if err != nil {
			fmt.Fprintf(output, "  %s: extract error: %v\n", provider, err)
			continue
		}
		fmt.Fprintf(output, "  %s: found location data\n", provider)
		loc = extracted
		break
	}

	if loc == nil {
		fmt.Fprintln(output, "\nNo cached location. Using simulated data.")
		loc = &location.ExtractedLocation{
			Latitude:  48.8566,
			Longitude: 2.3522,
			Altitude:  35.0,
			Speed:     12.5,
			Bearing:   270.0,
			Accuracy:  10.0,
			Time:      time.Now().UnixMilli(),
			Provider:  "simulated",
		}
	}

	fmt.Fprintln(output, "\n=== Location Details ===")
	fmt.Fprintf(output, "  Provider:  %s\n", loc.Provider)
	fmt.Fprintf(output, "  Latitude:  %.6f\n", loc.Latitude)
	fmt.Fprintf(output, "  Longitude: %.6f\n", loc.Longitude)

	fmt.Fprintln(output, "\n=== Altitude ===")
	fmt.Fprintf(output, "  Altitude:  %.1f m\n", loc.Altitude)

	fmt.Fprintln(output, "\n=== Speed ===")
	speedMS := float64(loc.Speed)
	speedKMH := speedMS * 3.6
	fmt.Fprintf(output, "  Speed:     %.2f m/s\n", speedMS)
	fmt.Fprintf(output, "  Speed:     %.2f km/h\n", speedKMH)

	fmt.Fprintln(output, "\n=== Bearing ===")
	fmt.Fprintf(output, "  Bearing:   %.1f deg\n", float64(loc.Bearing))
	fmt.Fprintf(output, "  Direction: %s\n", bearingToCardinal(float64(loc.Bearing)))

	fmt.Fprintln(output, "\n=== Accuracy ===")
	fmt.Fprintf(output, "  Accuracy:  %.1f m\n", float64(loc.Accuracy))

	fixTime := time.UnixMilli(loc.Time)
	fmt.Fprintf(output, "\nFix time: %s\n", fixTime.UTC().Format(time.RFC3339))

	fmt.Fprintln(output, "\nGPS speed/altitude example completed.")
	return nil
}

// bearingToCardinal converts a bearing in degrees to a cardinal direction.
func bearingToCardinal(bearing float64) string {
	dirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	idx := int((bearing + 22.5) / 45.0)
	return dirs[idx%8]
}
