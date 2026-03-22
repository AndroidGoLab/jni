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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/app/job"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(activity.vm)),
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
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

	// Print all result code constants.
	fmt.Fprintln(output, "Result codes:")
	fmt.Fprintf(output, "  ResultSuccess = %d\n", job.ResultSuccess)
	fmt.Fprintf(output, "  ResultFailure = %d\n", job.ResultFailure)

	// Print all network type constants.
	fmt.Fprintln(output, "Network types:")
	fmt.Fprintf(output, "  NetworkTypeNone       = %d\n", job.NetworkTypeNone)
	fmt.Fprintf(output, "  NetworkTypeAny        = %d\n", job.NetworkTypeAny)
	fmt.Fprintf(output, "  NetworkTypeUnmetered  = %d\n", job.NetworkTypeUnmetered)
	fmt.Fprintf(output, "  NetworkTypeNotRoaming = %d\n", job.NetworkTypeNotRoaming)
	fmt.Fprintf(output, "  NetworkTypeCellular   = %d\n", job.NetworkTypeCellular)

	// Print backoff policy constants.
	fmt.Fprintln(output, "Backoff policies:")
	fmt.Fprintf(output, "  BackoffPolicyLinear      = %d\n", job.BackoffPolicyLinear)
	fmt.Fprintf(output, "  BackoffPolicyExponential = %d\n", job.BackoffPolicyExponential)

	// CancelInAllNamespaces cancels all jobs in all namespaces.
	if err := sched.CancelInAllNamespaces(); err != nil {
		fmt.Fprintf(output, "CancelInAllNamespaces: %v (expected on older APIs)\n", err)
	} else {
		fmt.Fprintln(output, "cancelled all jobs in all namespaces")
	}

	fmt.Fprintln(output, "JobScheduler example complete")
	return nil
}
