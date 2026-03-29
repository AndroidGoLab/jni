//go:build android

// Command location_reminder combines location check with alarm: gets location,
// checks if near a target point, and queries the AlarmManager. Shows
// cross-package integration between location and app/alarm packages.
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
	"github.com/AndroidGoLab/jni/app/alarm"
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

// Target location (reminder trigger point).
const (
	targetLat   = 40.7128  // New York City
	targetLon   = -74.0060
	triggerDist = 1000.0   // 1 km radius
	earthRadius = 6371000.0
)

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- Location Part ---
	fmt.Fprintln(output, "=== Location Check ===")
	fmt.Fprintf(output, "Target: lat=%.4f lon=%.4f (radius %.0f m)\n",
		targetLat, targetLon, triggerDist)

	locMgr, err := location.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("location.NewManager: %w", err)
	}
	defer locMgr.Close()

	var loc *location.ExtractedLocation
	providers := []string{location.GpsProvider, location.NetworkProvider, location.PassiveProvider}
	for _, provider := range providers {
		enabled, _ := locMgr.IsProviderEnabled(provider)
		if !enabled {
			continue
		}
		locObj, err := locMgr.GetLastKnownLocation(provider)
		if err != nil || locObj == nil || locObj.Ref() == 0 {
			continue
		}
		vm.Do(func(env *jni.Env) error {
			loc, err = location.ExtractLocation(env, locObj)
			return err
		})
		if err == nil && loc != nil {
			break
		}
	}

	if loc == nil {
		fmt.Fprintln(output, "No cached location. Using simulated position.")
		loc = &location.ExtractedLocation{
			Latitude:  40.7580,
			Longitude: -73.9855,
			Provider:  "simulated",
		}
	}

	fmt.Fprintf(output, "Current: lat=%.6f lon=%.6f (%s)\n",
		loc.Latitude, loc.Longitude, loc.Provider)

	dist := haversine(loc.Latitude, loc.Longitude, targetLat, targetLon)
	nearTarget := dist <= triggerDist
	fmt.Fprintf(output, "Distance to target: %.0f m\n", dist)

	if nearTarget {
		fmt.Fprintln(output, "Status: NEAR target - reminder would trigger!")
	} else {
		fmt.Fprintf(output, "Status: FAR from target (%.0f m away)\n", dist-triggerDist)
	}

	// --- Alarm Part ---
	fmt.Fprintln(output, "\n=== Alarm Manager ===")

	alarmMgr, err := alarm.NewManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "alarm.NewManager: %v\n", err)
		fmt.Fprintln(output, "(AlarmManager not available)")
	} else {
		defer alarmMgr.Close()

		canSchedule, err := alarmMgr.CanScheduleExactAlarms()
		if err != nil {
			fmt.Fprintf(output, "CanScheduleExactAlarms: %v\n", err)
		} else {
			fmt.Fprintf(output, "Can schedule exact alarms: %v\n", canSchedule)
		}

		nextAlarmObj, err := alarmMgr.GetNextAlarmClock()
		if err != nil {
			fmt.Fprintf(output, "GetNextAlarmClock: %v\n", err)
		} else if nextAlarmObj == nil || nextAlarmObj.Ref() == 0 {
			fmt.Fprintln(output, "No upcoming alarm clock set")
		} else {
			var info *alarm.AlarmClockInfo
			vm.Do(func(env *jni.Env) error {
				info, err = alarm.ExtractAlarmClockInfo(env, nextAlarmObj)
				return err
			})
			if err == nil && info != nil {
				t := time.UnixMilli(info.TriggerTime)
				fmt.Fprintf(output, "Next alarm: %s\n", t.Format(time.RFC3339))
			}
		}

		fmt.Fprintln(output, "\nAlarm type constants:")
		fmt.Fprintf(output, "  RTC_WAKEUP:             %d\n", alarm.RtcWakeup)
		fmt.Fprintf(output, "  ELAPSED_REALTIME_WAKEUP: %d\n", alarm.ElapsedRealtimeWakeup)

		fmt.Fprintln(output, "\nAlarm intervals:")
		fmt.Fprintf(output, "  INTERVAL_HOUR:    %d ms\n", alarm.IntervalHour)
		fmt.Fprintf(output, "  INTERVAL_DAY:     %d ms\n", alarm.IntervalDay)
	}

	fmt.Fprintln(output, "\nLocation reminder example completed.")
	return nil
}

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
