//go:build android

// Command usage demonstrates using the Android UsageStatsManager
// system service, wrapped by the usage package. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// The usage package wraps android.app.usage.UsageStatsManager and
// provides the Stats data class, Interval type constants, and
// standby bucket constants. It requires the
// android.permission.PACKAGE_USAGE_STATS permission.
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/app/usage"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
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

	mgr, err := usage.NewStatsManager(ctx)
	if err != nil {
		return fmt.Errorf("usage.NewStatsManager: %w", err)
	}

	// Check if an app is inactive.
	inactive, err := mgr.IsAppInactive("com.example.app")
	if err != nil {
		return fmt.Errorf("IsAppInactive: %w", err)
	}
	fmt.Fprintf(output, "com.example.app inactive: %v\n", inactive)

	// Get the app's standby bucket.
	bucket, err := mgr.GetAppStandbyBucket()
	if err != nil {
		return fmt.Errorf("GetAppStandbyBucket: %w", err)
	}
	fmt.Fprintf(output, "app standby bucket: %d\n", bucket)

	// Manager also provides the unexported method:
	//   queryUsageStatsRaw(intervalType, beginTime, endTime)
	//     -- queries usage statistics for a time range.

	// --- Interval Type Constants ---
	// These correspond to UsageStatsManager.INTERVAL_* values:
	fmt.Fprintf(output, "IntervalDaily:   %d\n", usage.IntervalDaily)
	fmt.Fprintf(output, "IntervalWeekly:  %d\n", usage.IntervalWeekly)
	fmt.Fprintf(output, "IntervalMonthly: %d\n", usage.IntervalMonthly)
	fmt.Fprintf(output, "IntervalYearly:  %d\n", usage.IntervalYearly)

	// --- Standby Bucket Constants ---
	// These correspond to UsageStatsManager.STANDBY_BUCKET_* values:
	fmt.Fprintf(output, "StandbyBucketActive:     %d\n", usage.StandbyBucketActive)
	fmt.Fprintf(output, "StandbyBucketWorkingSet: %d\n", usage.StandbyBucketWorkingSet)
	fmt.Fprintf(output, "StandbyBucketFrequent:   %d\n", usage.StandbyBucketFrequent)
	fmt.Fprintf(output, "StandbyBucketRare:       %d\n", usage.StandbyBucketRare)
	fmt.Fprintf(output, "StandbyBucketRestricted: %d\n", usage.StandbyBucketRestricted)

	_ = mgr
	return nil
}
