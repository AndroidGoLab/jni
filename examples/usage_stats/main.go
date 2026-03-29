//go:build android

// Command usage_stats uses UsageStatsManager to query app usage
// statistics and shows daily usage concepts for recent apps.
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
	"github.com/AndroidGoLab/jni/app/usage"
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

// standbyBucketName converts a bucket value to a name.
func standbyBucketName(bucket int32) string {
	switch int(bucket) {
	case usage.StandbyBucketActive:
		return "ACTIVE"
	case usage.StandbyBucketWorkingSet:
		return "WORKING_SET"
	case usage.StandbyBucketFrequent:
		return "FREQUENT"
	case usage.StandbyBucketRare:
		return "RARE"
	case usage.StandbyBucketRestricted:
		return "RESTRICTED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", bucket)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Usage Stats ===")

	mgr, err := usage.NewStatsManager(ctx)
	if err != nil {
		return fmt.Errorf("usage.NewStatsManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "UsageStatsManager: obtained")

	// Get app standby bucket.
	bucket, err := mgr.GetAppStandbyBucket()
	if err != nil {
		fmt.Fprintf(output, "StandbyBucket: %v\n", err)
	} else {
		fmt.Fprintf(output, "StandbyBucket: %s (%d)\n", standbyBucketName(bucket), bucket)
	}

	// Check if some common apps are inactive.
	fmt.Fprintln(output, "\nApp Inactive Status:")
	apps := []string{
		"com.android.settings",
		"com.android.chrome",
		"com.google.android.gms",
		"com.example.nonexistent",
	}

	for _, pkg := range apps {
		inactive, err := mgr.IsAppInactive(pkg)
		if err != nil {
			fmt.Fprintf(output, "  %-30s ERR: %v\n", pkg, err)
		} else {
			fmt.Fprintf(output, "  %-30s inactive=%v\n", pkg, inactive)
		}
	}

	// Interval type constants.
	fmt.Fprintln(output, "\nInterval Constants:")
	fmt.Fprintf(output, "  INTERVAL_DAILY:   %d\n", usage.IntervalDaily)
	fmt.Fprintf(output, "  INTERVAL_WEEKLY:  %d\n", usage.IntervalWeekly)
	fmt.Fprintf(output, "  INTERVAL_MONTHLY: %d\n", usage.IntervalMonthly)
	fmt.Fprintf(output, "  INTERVAL_YEARLY:  %d\n", usage.IntervalYearly)

	// Standby bucket constants.
	fmt.Fprintln(output, "\nStandby Bucket Constants:")
	fmt.Fprintf(output, "  ACTIVE:      %d\n", usage.StandbyBucketActive)
	fmt.Fprintf(output, "  WORKING_SET: %d\n", usage.StandbyBucketWorkingSet)
	fmt.Fprintf(output, "  FREQUENT:    %d\n", usage.StandbyBucketFrequent)
	fmt.Fprintf(output, "  RARE:        %d\n", usage.StandbyBucketRare)
	fmt.Fprintf(output, "  RESTRICTED:  %d\n", usage.StandbyBucketRestricted)

	fmt.Fprintln(output, "\nUsage stats complete.")
	return nil
}
