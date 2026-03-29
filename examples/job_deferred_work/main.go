//go:build android

// Command job_deferred_work demonstrates the JobScheduler API.
// It shows job scheduler constants, checks pending jobs, and
// demonstrates the cancel and cancelAll methods.
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
	"github.com/AndroidGoLab/jni/app/job"
	"github.com/AndroidGoLab/jni/examples/common/ui"
)

const (
	testJobID = 42
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

	sched, err := job.NewScheduler(ctx)
	if err != nil {
		return fmt.Errorf("job.NewScheduler: %w", err)
	}
	defer sched.Close()

	// --- Print constants ---
	fmt.Fprintln(output, "=== JobScheduler Constants ===")
	fmt.Fprintf(output, "ResultSuccess = %d\n", job.ResultSuccess)
	fmt.Fprintf(output, "ResultFailure = %d\n", job.ResultFailure)
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Network types:")
	fmt.Fprintf(output, "  NONE        = %d\n", job.NetworkTypeNone)
	fmt.Fprintf(output, "  ANY         = %d\n", job.NetworkTypeAny)
	fmt.Fprintf(output, "  UNMETERED   = %d\n", job.NetworkTypeUnmetered)
	fmt.Fprintf(output, "  NOT_ROAMING = %d\n", job.NetworkTypeNotRoaming)
	fmt.Fprintf(output, "  CELLULAR    = %d\n", job.NetworkTypeCellular)
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Backoff policies:")
	fmt.Fprintf(output, "  LINEAR      = %d\n", job.BackoffPolicyLinear)
	fmt.Fprintf(output, "  EXPONENTIAL = %d\n", job.BackoffPolicyExponential)
	fmt.Fprintln(output, "")
	fmt.Fprintf(output, "DEFAULT_INITIAL_BACKOFF = %d ms\n", job.DefaultInitialBackoffMillis)
	fmt.Fprintf(output, "MAX_BACKOFF_DELAY       = %d ms\n", job.MaxBackoffDelayMillis)

	// --- Check pending job ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Check Pending Job ===")
	pendingJob, err := sched.GetPendingJob(testJobID)
	if err != nil {
		fmt.Fprintf(output, "GetPendingJob(%d): %v\n", testJobID, err)
	} else if pendingJob != nil && pendingJob.Ref() != 0 {
		fmt.Fprintf(output, "job %d is pending\n", testJobID)
	} else {
		fmt.Fprintf(output, "job %d not found in pending\n", testJobID)
	}

	// --- Get all pending jobs ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== All Pending Jobs ===")
	allPending, err := sched.GetAllPendingJobs()
	if err != nil {
		fmt.Fprintf(output, "GetAllPendingJobs: %v\n", err)
	} else if allPending != nil && allPending.Ref() != 0 {
		fmt.Fprintln(output, "pending jobs list obtained")
	} else {
		fmt.Fprintln(output, "no pending jobs")
	}

	// --- Check canRunUserInitiatedJobs ---
	canRun, err := sched.CanRunUserInitiatedJobs()
	if err != nil {
		fmt.Fprintf(output, "\nCanRunUserInitiatedJobs: %v\n", err)
	} else {
		fmt.Fprintf(output, "\ncan run user-initiated jobs: %v\n", canRun)
	}

	// --- Cancel a specific job ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Cancel Job ===")
	if err := sched.Cancel(testJobID); err != nil {
		fmt.Fprintf(output, "Cancel(%d): %v\n", testJobID, err)
	} else {
		fmt.Fprintf(output, "job %d cancelled\n", testJobID)
	}

	fmt.Fprintln(output, "\njob_deferred_work complete")
	return nil
}
