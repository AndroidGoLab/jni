//go:build android

// Command location_geofence gets the last known location, defines a geofence
// center point, computes the distance to it, and reports whether the device
// is inside or outside the zone. Uses pure Go math on location data.
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

// Geofence parameters: center point and radius in meters.
const (
	geofenceLat    = 37.4220   // Google HQ latitude
	geofenceLon    = -122.0841 // Google HQ longitude
	geofenceRadius = 500.0     // radius in meters
	earthRadius    = 6371000.0 // Earth radius in meters
)

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

	fmt.Fprintln(output, "=== Geofence Configuration ===")
	fmt.Fprintf(output, "  Center: lat=%.4f lon=%.4f\n", geofenceLat, geofenceLon)
	fmt.Fprintf(output, "  Radius: %.0f m\n", geofenceRadius)

	// Try to get a location from any available provider.
	providers := []string{location.GpsProvider, location.NetworkProvider, location.PassiveProvider}

	fmt.Fprintln(output, "\n=== Searching for location ===")
	var loc *location.ExtractedLocation
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
		vm.Do(func(env *jni.Env) error {
			loc, err = location.ExtractLocation(env, locObj)
			return err
		})
		if err != nil {
			fmt.Fprintf(output, "  %s: extract error: %v\n", provider, err)
			continue
		}
		fmt.Fprintf(output, "  %s: lat=%.6f lon=%.6f\n", provider, loc.Latitude, loc.Longitude)
		break
	}

	if loc == nil {
		fmt.Fprintln(output, "\nNo cached location available.")
		fmt.Fprintln(output, "Using simulated position: lat=37.4225 lon=-122.0840")
		loc = &location.ExtractedLocation{
			Latitude:  37.4225,
			Longitude: -122.0840,
			Provider:  "simulated",
		}
	}

	fmt.Fprintln(output, "\n=== Geofence Check ===")
	fmt.Fprintf(output, "  Device:   lat=%.6f lon=%.6f (%s)\n",
		loc.Latitude, loc.Longitude, loc.Provider)

	dist := haversine(loc.Latitude, loc.Longitude, geofenceLat, geofenceLon)
	inside := dist <= geofenceRadius

	fmt.Fprintf(output, "  Distance: %.1f m\n", dist)
	if inside {
		fmt.Fprintf(output, "  Status:   INSIDE geofence (%.0f m radius)\n", geofenceRadius)
	} else {
		fmt.Fprintf(output, "  Status:   OUTSIDE geofence (%.0f m radius)\n", geofenceRadius)
	}

	fmt.Fprintln(output, "\nLocation geofence example completed.")
	return nil
}

// haversine computes the distance in meters between two lat/lon points
// using the haversine formula.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}
