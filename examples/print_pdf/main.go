//go:build android

// Command print_pdf uses PrintManager to show print service availability
// and demonstrate the print job API surface.
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

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Print PDF ===")

	// Obtain the PrintManager system service.
	mgr, err := print.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("print.NewManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "PrintManager: obtained OK")

	// Show print job state constants.
	fmt.Fprintln(output, "\nPrint Job State Constants:")
	fmt.Fprintf(output, "  Created:   %d\n", print.StateCreated)
	fmt.Fprintf(output, "  Queued:    %d\n", print.StateQueued)
	fmt.Fprintf(output, "  Started:   %d\n", print.StateStarted)
	fmt.Fprintf(output, "  Blocked:   %d\n", print.StateBlocked)
	fmt.Fprintf(output, "  Completed: %d\n", print.StateCompleted)
	fmt.Fprintf(output, "  Failed:    %d\n", print.StateFailed)
	fmt.Fprintf(output, "  Canceled:  %d\n", print.StateCanceled)

	// Show printer status constants.
	fmt.Fprintln(output, "\nPrinter Status Constants:")
	fmt.Fprintf(output, "  Idle:        %d\n", print.StatusIdle)
	fmt.Fprintf(output, "  Busy:        %d\n", print.StatusBusy)
	fmt.Fprintf(output, "  Unavailable: %d\n", print.StatusUnavailable)

	// Show print attribute constants.
	fmt.Fprintln(output, "\nPrint Attribute Constants:")
	fmt.Fprintf(output, "  ColorModeColor:       %d\n", print.ColorModeColor)
	fmt.Fprintf(output, "  ColorModeMonochrome:  %d\n", print.ColorModeMonochrome)
	fmt.Fprintf(output, "  DuplexModeNone:       %d\n", print.DuplexModeNone)
	fmt.Fprintf(output, "  DuplexModeLongEdge:   %d\n", print.DuplexModeLongEdge)
	fmt.Fprintf(output, "  DuplexModeShortEdge:  %d\n", print.DuplexModeShortEdge)
	fmt.Fprintf(output, "  ContentTypeDocument:  %d\n", print.ContentTypeDocument)
	fmt.Fprintf(output, "  ContentTypePhoto:     %d\n", print.ContentTypePhoto)

	fmt.Fprintln(output, "\nPrint PDF complete.")
	return nil
}
