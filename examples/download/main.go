//go:build android

// Command download demonstrates using the Android DownloadManager API.
// It obtains the DownloadManager system service, prints status constants,
// and queries all downloads by status.
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
	"github.com/AndroidGoLab/jni/app/download"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/examples/common/ui"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
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

	fmt.Fprintln(output, "=== DownloadManager ===")

	// Obtain the DownloadManager system service.
	mgr, err := download.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("download.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "Service obtained OK")

	// Print status constants.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Status constants:")
	fmt.Fprintf(output, "  Pending:    %d\n", download.StatusPending)
	fmt.Fprintf(output, "  Running:    %d\n", download.StatusRunning)
	fmt.Fprintf(output, "  Paused:     %d\n", download.StatusPaused)
	fmt.Fprintf(output, "  Successful: %d\n", download.StatusSuccessful)
	fmt.Fprintf(output, "  Failed:     %d\n", download.StatusFailed)

	// Create a DownloadManager.Query to query all downloads,
	// then call mgr.Query() with it.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Querying downloads...")

	var queryCount int32
	queryErr := vm.Do(func(env *jni.Env) error {
		// Construct new DownloadManager.Query() via JNI.
		queryCls, err := env.FindClass("android/app/DownloadManager$Query")
		if err != nil {
			return fmt.Errorf("find Query class: %w", err)
		}
		initMid, err := env.GetMethodID(queryCls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("get Query init: %w", err)
		}
		queryObj, err := env.NewObject(queryCls, initMid)
		if err != nil {
			return fmt.Errorf("new Query: %w", err)
		}

		// Call mgr.Query(queryObj) to get a Cursor.
		cursorObj, err := mgr.Query(queryObj)
		if err != nil {
			return fmt.Errorf("mgr.Query: %w", err)
		}
		if cursorObj == nil {
			fmt.Fprintln(output, "  Query returned nil cursor")
			return nil
		}

		// Get cursor count via Cursor.getCount().
		cursorCls, err := env.FindClass("android/database/Cursor")
		if err != nil {
			return fmt.Errorf("find Cursor class: %w", err)
		}
		getCountMid, err := env.GetMethodID(cursorCls, "getCount", "()I")
		if err != nil {
			return fmt.Errorf("get getCount: %w", err)
		}
		queryCount, err = env.CallIntMethod(cursorObj, getCountMid)
		if err != nil {
			return fmt.Errorf("getCount: %w", err)
		}

		// Close the cursor.
		closeMid, err := env.GetMethodID(cursorCls, "close", "()V")
		if err == nil {
			_ = env.CallVoidMethod(cursorObj, closeMid)
		}

		return nil
	})

	if queryErr != nil {
		fmt.Fprintf(output, "  Query error: %v\n", queryErr)
	} else {
		fmt.Fprintf(output, "  Total downloads: %d\n", queryCount)
	}

	// Print visibility constants.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Visibility constants:")
	fmt.Fprintf(output, "  Hidden:      %d\n", download.VisibilityHidden)
	fmt.Fprintf(output, "  Visible:     %d\n", download.VisibilityVisible)
	fmt.Fprintf(output, "  NotifyDone:  %d\n", download.VisibilityVisibleNotifyCompleted)

	// Print column name constants.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Column names:")
	fmt.Fprintf(output, "  ID:     %s\n", download.ColumnId)
	fmt.Fprintf(output, "  Title:  %s\n", download.ColumnTitle)
	fmt.Fprintf(output, "  Status: %s\n", download.ColumnStatus)
	fmt.Fprintf(output, "  Size:   %s\n", download.ColumnTotalSizeBytes)
	fmt.Fprintf(output, "  URI:    %s\n", download.ColumnUri)

	return nil
}
