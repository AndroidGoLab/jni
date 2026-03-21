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
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/app/usage"
)

func main() {}

func init() { exampleui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	exampleui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	exampleui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	exampleui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := usage.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("usage.NewManager: %w", err)
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

// getAppContext obtains an Android Context via ActivityThread.currentApplication().
func getAppContext(vm *jni.VM) (*app.Context, error) {
	var ctx app.Context
	ctx.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}

		atClass, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread: %w", err)
		}

		curAppMid, err := env.GetStaticMethodID(atClass, "currentApplication", "()Landroid/app/Application;")
		if err != nil {
			return fmt.Errorf("get currentApplication: %w", err)
		}
		appObj, err := env.CallStaticObjectMethod(atClass, curAppMid)
		if err != nil {
			return fmt.Errorf("call currentApplication: %w", err)
		}
		if appObj == nil || appObj.Ref() == 0 {
			return fmt.Errorf("currentApplication returned null")
		}

		ctx.Obj = env.NewGlobalRef(appObj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}
