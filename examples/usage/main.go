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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/app/usage"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	vm := jni.VMFromPtr(unsafe.Pointer(cvm))
	if err := run(vm); err != nil {
		fmt.Fprintf(&output, "ERROR: %v\n", err)
	}
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}

func run(vm *jni.VM) error {
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
	fmt.Fprintf(&output, "com.example.app inactive: %v\n", inactive)

	// Get the app's standby bucket.
	bucket, err := mgr.GetAppStandbyBucket()
	if err != nil {
		return fmt.Errorf("GetAppStandbyBucket: %w", err)
	}
	fmt.Fprintf(&output, "app standby bucket: %d\n", bucket)

	// Manager also provides the unexported method:
	//   queryUsageStatsRaw(intervalType, beginTime, endTime)
	//     -- queries usage statistics for a time range.

	// --- Interval Type Constants ---
	// These correspond to UsageStatsManager.INTERVAL_* values:
	fmt.Fprintf(&output, "Daily:   %d\n", usage.Daily)
	fmt.Fprintf(&output, "Weekly:  %d\n", usage.Weekly)
	fmt.Fprintf(&output, "Monthly: %d\n", usage.Monthly)
	fmt.Fprintf(&output, "Yearly:  %d\n", usage.Yearly)

	// --- Standby Bucket Constants ---
	// These correspond to UsageStatsManager.STANDBY_BUCKET_* values:
	fmt.Fprintf(&output, "StandbyActive:     %d\n", usage.StandbyActive)
	fmt.Fprintf(&output, "StandbyWorkingSet: %d\n", usage.StandbyWorkingSet)
	fmt.Fprintf(&output, "StandbyFrequent:   %d\n", usage.StandbyFrequent)
	fmt.Fprintf(&output, "StandbyRare:       %d\n", usage.StandbyRare)
	fmt.Fprintf(&output, "StandbyRestricted: %d\n", usage.StandbyRestricted)

	// --- Stats Data Class ---
	// Stats holds data extracted from android.app.usage.UsageStats:
	var stats usage.Stats
	fmt.Fprintf(&output, "Stats.PackageName:         %q\n", stats.PackageName)
	fmt.Fprintf(&output, "Stats.FirstTimestamp:      %d\n", stats.FirstTimestamp)
	fmt.Fprintf(&output, "Stats.LastTimestamp:       %d\n", stats.LastTimestamp)
	fmt.Fprintf(&output, "Stats.TotalTimeVisible:    %d\n", stats.TotalTimeVisible)
	fmt.Fprintf(&output, "Stats.TotalTimeForeground: %d\n", stats.TotalTimeForeground)

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
