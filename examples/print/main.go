//go:build android

// Command print demonstrates using the PrintManager API. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the PrintManager system service and shows all
// print job state constants. The Manager provides methods for querying
// print jobs and starting new ones. The printJob and printJobInfo data
// classes (unexported) hold job metadata extracted from JNI objects.
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
	"github.com/AndroidGoLab/jni/print"
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
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- Constants ---
	// JobState is a typed constant representing print job states.
	fmt.Fprintln(output, "Print job state constants (type JobState):")
	fmt.Fprintf(output, "  JobCreated   = %d\n", print.StateCreated)
	fmt.Fprintf(output, "  JobQueued    = %d\n", print.StateQueued)
	fmt.Fprintf(output, "  JobStarted   = %d\n", print.StateStarted)
	fmt.Fprintf(output, "  JobBlocked   = %d\n", print.StateBlocked)
	fmt.Fprintf(output, "  JobCompleted = %d\n", print.StateCompleted)
	fmt.Fprintf(output, "  JobFailed    = %d\n", print.StateFailed)
	fmt.Fprintf(output, "  JobCanceled  = %d\n", print.StateCanceled)

	// --- NewManager ---
	mgr, err := print.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("print.NewManager: %w", err)
	}
	_ = mgr

	// --- Manager methods (unexported) ---
	//
	//   mgr.getPrintJobsRaw() (*jni.Object, error)
	//     Query all print jobs visible to this application.
	//
	//   mgr.printRaw(jobName string, adapter, attributes *jni.Object)
	//     Start a new print job with a PrintDocumentAdapter.

	// --- Data classes (unexported) ---
	//
	// printJobInfo{
	//   Label string
	//   State JobState
	// }
	// Extracted from android.print.PrintJobInfo via extractprintJobInfo.
	//
	// printJob{
	//   Info printJobInfo
	// }
	// Extracted from android.print.PrintJob via extractprintJob.

	// --- JobState typed constant ---
	// JobState values can be used in switch statements.
	state := print.StateCompleted
	switch state {
	case print.StateCreated:
		fmt.Fprintln(output, "job: created")
	case print.StateQueued:
		fmt.Fprintln(output, "job: queued")
	case print.StateStarted:
		fmt.Fprintln(output, "job: started")
	case print.StateBlocked:
		fmt.Fprintln(output, "job: blocked")
	case print.StateCompleted:
		fmt.Fprintln(output, "job: completed")
	case print.StateFailed:
		fmt.Fprintln(output, "job: failed")
	case print.StateCanceled:
		fmt.Fprintln(output, "job: canceled")
	}
	return nil
}
