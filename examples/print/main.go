//go:build android

// Command print demonstrates using the PrintManager API.
// It obtains the PrintManager system service, queries existing print jobs,
// and prints their details.
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
	"github.com/AndroidGoLab/jni/print"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
}

// jobStateName returns a human-readable name for a print job state.
func jobStateName(state int) string {
	switch state {
	case print.StateCreated:
		return "Created"
	case print.StateQueued:
		return "Queued"
	case print.StateStarted:
		return "Started"
	case print.StateBlocked:
		return "Blocked"
	case print.StateCompleted:
		return "Completed"
	case print.StateFailed:
		return "Failed"
	case print.StateCanceled:
		return "Canceled"
	default:
		return fmt.Sprintf("Unknown(%d)", state)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== PrintManager ===")

	// Print job state constants.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Job state constants:")
	fmt.Fprintf(output, "  Created:   %d\n", print.StateCreated)
	fmt.Fprintf(output, "  Queued:    %d\n", print.StateQueued)
	fmt.Fprintf(output, "  Started:   %d\n", print.StateStarted)
	fmt.Fprintf(output, "  Blocked:   %d\n", print.StateBlocked)
	fmt.Fprintf(output, "  Completed: %d\n", print.StateCompleted)
	fmt.Fprintf(output, "  Failed:    %d\n", print.StateFailed)
	fmt.Fprintf(output, "  Canceled:  %d\n", print.StateCanceled)

	// Obtain the PrintManager system service.
	mgr, err := print.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("print.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "\nService obtained OK")

	// Filtered: GetPrintJobs returns generic type (List<PrintJob>)
	// jobsListObj, err := mgr.GetPrintJobs()
	// ...

	return nil
}
