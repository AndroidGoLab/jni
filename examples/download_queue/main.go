//go:build android

// Command download_queue demonstrates the DownloadManager API. It shows
// status constants, network constants, column names, and visibility
// settings available via the typed wrapper.
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
	"github.com/AndroidGoLab/jni/app/download"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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

	mgr, err := download.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("download.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== DownloadManager ===")
	fmt.Fprintln(output, "Service obtained OK")

	// --- Print status constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Status constants:")
	fmt.Fprintf(output, "  Pending:    %d\n", download.StatusPending)
	fmt.Fprintf(output, "  Running:    %d\n", download.StatusRunning)
	fmt.Fprintf(output, "  Paused:     %d\n", download.StatusPaused)
	fmt.Fprintf(output, "  Successful: %d\n", download.StatusSuccessful)
	fmt.Fprintf(output, "  Failed:     %d\n", download.StatusFailed)

	// --- Print network constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Network constants:")
	fmt.Fprintf(output, "  MOBILE: %d\n", download.NetworkMobile)
	fmt.Fprintf(output, "  WIFI:   %d\n", download.NetworkWifi)

	// --- Print column constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Column names:")
	fmt.Fprintf(output, "  ID:     %s\n", download.ColumnId)
	fmt.Fprintf(output, "  Title:  %s\n", download.ColumnTitle)
	fmt.Fprintf(output, "  Status: %s\n", download.ColumnStatus)
	fmt.Fprintf(output, "  Size:   %s\n", download.ColumnTotalSizeBytes)
	fmt.Fprintf(output, "  URI:    %s\n", download.ColumnUri)

	// --- Print visibility constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Visibility constants:")
	fmt.Fprintf(output, "  Hidden:      %d\n", download.VisibilityHidden)
	fmt.Fprintf(output, "  Visible:     %d\n", download.VisibilityVisible)
	fmt.Fprintf(output, "  NotifyDone:  %d\n", download.VisibilityVisibleNotifyCompleted)

	fmt.Fprintln(output, "\ndownload_queue complete")
	return nil
}
