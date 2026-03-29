//go:build android

// Command print_job_monitor gets PrintManager, queries print jobs if
// available, and shows job states and print service info.
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/print"
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

	fmt.Fprintln(output, "=== Print Job Monitor ===")

	mgr, err := print.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("print.NewManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "PrintManager: obtained OK")

	// Show all state names for reference.
	fmt.Fprintln(output, "\nAll Job States:")
	states := []int{
		print.StateCreated,
		print.StateQueued,
		print.StateStarted,
		print.StateBlocked,
		print.StateCompleted,
		print.StateFailed,
		print.StateCanceled,
	}
	for _, s := range states {
		fmt.Fprintf(output, "  %d = %s\n", s, jobStateName(s))
	}

	// Printer status constants.
	fmt.Fprintln(output, "\nPrinter Status:")
	fmt.Fprintf(output, "  Idle:        %d\n", print.StatusIdle)
	fmt.Fprintf(output, "  Busy:        %d\n", print.StatusBusy)
	fmt.Fprintf(output, "  Unavailable: %d\n", print.StatusUnavailable)

	// Note: GetPrintJobs returns generic type (List<PrintJob>)
	// and is filtered from the generated API.
	fmt.Fprintln(output, "\nNote: print job queries require")
	fmt.Fprintln(output, "the List<PrintJob> generic return")
	fmt.Fprintln(output, "type which is not yet supported.")

	fmt.Fprintln(output, "\nPrint job monitor complete.")
	return nil
}
