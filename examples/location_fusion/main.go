//go:build android

// Command location_fusion demonstrates the LocationManager API using typed
// wrappers. It lists provider constants, checks provider availability,
// and retrieves last known locations from each provider.
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

	// --- Provider constants ---
	fmt.Fprintln(output, "=== Provider Constants ===")
	fmt.Fprintf(output, "  GPS:     %q\n", location.GpsProvider)
	fmt.Fprintf(output, "  Network: %q\n", location.NetworkProvider)
	fmt.Fprintf(output, "  Passive: %q\n", location.PassiveProvider)

	// --- Location enabled ---
	locEnabled, err := mgr.IsLocationEnabled()
	if err != nil {
		fmt.Fprintf(output, "\nIsLocationEnabled: %v\n", err)
	} else {
		fmt.Fprintf(output, "\nLocation enabled: %v\n", locEnabled)
	}

	// --- Provider status ---
	knownProviders := []string{
		location.GpsProvider,
		location.NetworkProvider,
		location.PassiveProvider,
	}

	fmt.Fprintln(output, "\n=== Provider Status ===")
	for _, provider := range knownProviders {
		enabled, _ := mgr.IsProviderEnabled(provider)
		has, _ := mgr.HasProvider(provider)
		fmt.Fprintf(output, "  %s: enabled=%v exists=%v\n", provider, enabled, has)
	}

	// --- Collect last known location from each provider ---
	type providerResult struct {
		Name    string
		Loc     *location.ExtractedLocation
		Err     string
		Enabled bool
	}

	fmt.Fprintln(output, "\n=== Fusion: Location from each provider ===")
	var results []providerResult
	for _, provider := range knownProviders {
		r := providerResult{Name: provider}
		r.Enabled, _ = mgr.IsProviderEnabled(provider)
		if !r.Enabled {
			r.Err = "not enabled"
			results = append(results, r)
			continue
		}

		locObj, err := mgr.GetLastKnownLocation(provider)
		if err != nil {
			r.Err = err.Error()
			results = append(results, r)
			continue
		}
		if locObj == nil || locObj.Ref() == 0 {
			r.Err = "no cached location"
			results = append(results, r)
			continue
		}

		var loc *location.ExtractedLocation
		vm.Do(func(env *jni.Env) error {
			loc, err = location.ExtractLocation(env, locObj)
			return err
		})
		if err != nil {
			r.Err = fmt.Sprintf("extract: %v", err)
		} else {
			r.Loc = loc
		}
		results = append(results, r)
	}

	for _, r := range results {
		if r.Err != "" {
			fmt.Fprintf(output, "\n  [%s] %s\n", r.Name, r.Err)
			continue
		}
		fixAge := time.Since(time.UnixMilli(r.Loc.Time))
		fmt.Fprintf(output, "\n  [%s]\n", r.Name)
		fmt.Fprintf(output, "    lat=%.6f lon=%.6f\n", r.Loc.Latitude, r.Loc.Longitude)
		fmt.Fprintf(output, "    alt=%.1f m  acc=%.1f m\n", r.Loc.Altitude, float64(r.Loc.Accuracy))
		fmt.Fprintf(output, "    age=%s\n", fixAge.Round(time.Second))
	}

	// Compare results if we got multiple locations.
	fmt.Fprintln(output, "\n=== Comparison ===")
	var locs []*location.ExtractedLocation
	var locNames []string
	for _, r := range results {
		if r.Loc != nil {
			locs = append(locs, r.Loc)
			locNames = append(locNames, r.Name)
		}
	}

	if len(locs) < 2 {
		fmt.Fprintln(output, "Need at least 2 provider results to compare.")
	} else {
		for i := 0; i < len(locs); i++ {
			for j := i + 1; j < len(locs); j++ {
				dist := haversine(
					locs[i].Latitude, locs[i].Longitude,
					locs[j].Latitude, locs[j].Longitude,
				)
				fmt.Fprintf(output, "  %s vs %s: %.1f m apart\n",
					locNames[i], locNames[j], dist)
			}
		}

		// Pick the most accurate.
		bestIdx := 0
		for i, loc := range locs {
			if loc.Accuracy < locs[bestIdx].Accuracy || locs[bestIdx].Accuracy == 0 {
				bestIdx = i
			}
		}
		fmt.Fprintf(output, "\n  Best accuracy: %s (%.1f m)\n",
			locNames[bestIdx], float64(locs[bestIdx].Accuracy))
	}

	fmt.Fprintln(output, "\nLocation fusion example completed.")
	return nil
}

const earthRadius = 6371000.0

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
