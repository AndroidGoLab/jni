//go:build android

// Command job demonstrates the JobScheduler JNI bindings. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the JobScheduler system service and demonstrates
// the exported Cancel and CancelAll methods along with all available
// constants. The jobInfoBuilder type (created via NewjobInfoBuilder)
// provides package-internal methods for configuring job constraints.
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
	"github.com/AndroidGoLab/jni/app/job"
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

	sched, err := job.NewScheduler(ctx)
	if err != nil {
		return fmt.Errorf("job.NewScheduler: %v", err)
	}
	defer sched.Close()

	// Print all result code constants.
	fmt.Fprintln(&output, "Result codes:")
	fmt.Fprintf(&output, "  ResultSuccess = %d\n", job.ResultSuccess)
	fmt.Fprintf(&output, "  ResultFailure = %d\n", job.ResultFailure)

	// Print all network type constants.
	fmt.Fprintln(&output, "Network types:")
	fmt.Fprintf(&output, "  NetworkNone       = %d\n", job.NetworkNone)
	fmt.Fprintf(&output, "  NetworkAny        = %d\n", job.NetworkAny)
	fmt.Fprintf(&output, "  NetworkUnmetered  = %d\n", job.NetworkUnmetered)
	fmt.Fprintf(&output, "  NetworkNotRoaming = %d\n", job.NetworkNotRoaming)
	fmt.Fprintf(&output, "  NetworkCellular   = %d\n", job.NetworkCellular)

	// Print backoff policy constants.
	fmt.Fprintln(&output, "Backoff policies:")
	fmt.Fprintf(&output, "  BackoffPolicyLinear      = %d\n", job.BackoffPolicyLinear)
	fmt.Fprintf(&output, "  BackoffPolicyExponential = %d\n", job.BackoffPolicyExponential)

	// Cancel a specific job by ID.
	if err := sched.Cancel(42); err != nil {
		return fmt.Errorf("Cancel: %v", err)
	}
	fmt.Fprintln(&output, "cancelled job 42")

	// Cancel all pending jobs for this application.
	if err := sched.CancelAll(); err != nil {
		return fmt.Errorf("CancelAll: %v", err)
	}
	fmt.Fprintln(&output, "cancelled all jobs")

	// NewjobInfoBuilder(vm) creates a jobInfoBuilder that wraps
	// android.app.job.JobInfo.Builder. Its package-internal methods:
	//   setRequiredNetworkType(networkType int32)
	//   setRequiresCharging(requiresCharging bool)
	//   setRequiresDeviceIdle(requiresDeviceIdle bool)
	//   setRequiresBatteryNotLow(requiresBatteryNotLow bool)
	//   setRequiresStorageNotLow(requiresStorageNotLow bool)
	//   setPeriodic(intervalMillis int64)
	//   setMinimumLatency(minLatencyMillis int64)
	//   setOverrideDeadline(maxExecutionDelayMillis int64)
	//   setPersisted(isPersisted bool)
	//   setBackoffCriteria(initialBackoffMillis int64, backoffPolicy int32)
	//   build() - produces the JobInfo object
	//
	// Package-internal Scheduler methods:
	//   scheduleRaw(job) - schedule a job, returns result code
	//   getPendingJobRaw(jobId) - get a pending job by ID
	//   getAllPendingJobsRaw() - get all pending jobs
	//
	// The jobInfoJava data class extracts fields from a Java JobInfo:
	//   ID, Service, NetworkType, RequireCharging, RequireDeviceIdle,
	//   RequireBatteryNotLow, RequireStorageNotLow, IntervalMillis,
	//   MinLatencyMillis, MaxDelayMillis, Persisted,
	//   InitialBackoffMillis, BackoffPolicy
	builder, err := job.NewjobInfoBuilder(ctx.VM)
	if err != nil {
		return fmt.Errorf("job.NewjobInfoBuilder: %v", err)
	}
	_ = builder
	fmt.Fprintln(&output, "job info builder created")
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
