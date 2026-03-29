//go:build android

// Command download_queue demonstrates the DownloadManager API using the typed
// wrapper package. It obtains the download service, queries downloads,
// calls ToString, and exercises the GetMaxBytesOverMobile and
// GetRecommendedMaxBytesOverMobile static methods.
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

	fmt.Fprintln(output, "=== DownloadManager Demo ===")
	ui.RenderOutput()

	// --- Obtain DownloadManager ---
	mgr, err := download.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("download.NewManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "DownloadManager: obtained OK")
	ui.RenderOutput()

	// --- ToString ---
	mgrStr, err := mgr.ToString()
	if err != nil {
		fmt.Fprintf(output, "Manager.ToString: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Manager.ToString: %s\n", mgrStr)
	}
	ui.RenderOutput()

	// --- GetMaxBytesOverMobile (static) ---
	appCtxObj, err := ctx.GetApplicationContext()
	if err != nil {
		fmt.Fprintf(output, "GetApplicationContext: %v\n", err)
	} else {
		maxBytes, err := mgr.GetMaxBytesOverMobile(appCtxObj)
		if err != nil {
			fmt.Fprintf(output, "GetMaxBytesOverMobile: error (%v)\n", err)
		} else if maxBytes == nil {
			fmt.Fprintln(output, "GetMaxBytesOverMobile: null (no limit)")
		} else {
			fmt.Fprintln(output, "GetMaxBytesOverMobile: obtained OK")
			vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(maxBytes); return nil })
		}

		recMaxBytes, err := mgr.GetRecommendedMaxBytesOverMobile(appCtxObj)
		if err != nil {
			fmt.Fprintf(output, "GetRecommendedMaxBytesOverMobile: error (%v)\n", err)
		} else if recMaxBytes == nil {
			fmt.Fprintln(output, "GetRecommendedMaxBytesOverMobile: null (no limit)")
		} else {
			fmt.Fprintln(output, "GetRecommendedMaxBytesOverMobile: obtained OK")
			vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(recMaxBytes); return nil })
		}
	}
	ui.RenderOutput()

	// --- Query downloads with all status filters ---
	fmt.Fprintln(output, "\n=== Querying downloads ===")
	for _, statusFilter := range []struct {
		name   string
		status int
	}{
		{"Pending", download.StatusPending},
		{"Running", download.StatusRunning},
		{"Paused", download.StatusPaused},
		{"Successful", download.StatusSuccessful},
		{"Failed", download.StatusFailed},
	} {
		// Create a ManagerQuery via raw constructor (no typed constructor available).
		var queryObj *jni.GlobalRef
		err := vm.Do(func(env *jni.Env) error {
			cls, err := env.FindClass("android/app/DownloadManager$Query")
			if err != nil {
				return err
			}
			mid, err := env.GetMethodID(cls, "<init>", "()V")
			if err != nil {
				return err
			}
			obj, err := env.NewObject(cls, mid)
			if err != nil {
				return err
			}
			queryObj = env.NewGlobalRef(obj)
			return nil
		})
		if err != nil {
			fmt.Fprintf(output, "  Create Query: %v\n", err)
			continue
		}

		query := &download.ManagerQuery{VM: vm, Obj: queryObj}

		// SetFilterByStatus
		_, err = query.SetFilterByStatus(int32(statusFilter.status))
		if err != nil {
			fmt.Fprintf(output, "  SetFilterByStatus(%s): %v\n", statusFilter.name, err)
			vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(queryObj); return nil })
			continue
		}

		// Execute query
		cursorObj, err := mgr.Query(queryObj)
		if err != nil {
			fmt.Fprintf(output, "  Query(%s): %v\n", statusFilter.name, err)
		} else if cursorObj == nil {
			fmt.Fprintf(output, "  Query(%s): null cursor\n", statusFilter.name)
		} else {
			// Get count from cursor.
			var count int32
			vm.Do(func(env *jni.Env) error {
				cursorCls := env.GetObjectClass(cursorObj)
				getCountMid, err := env.GetMethodID(cursorCls, "getCount", "()I")
				if err != nil {
					return err
				}
				count, err = env.CallIntMethod(cursorObj, getCountMid)
				if err != nil {
					return err
				}
				// Close cursor.
				closeMid, err := env.GetMethodID(cursorCls, "close", "()V")
				if err != nil {
					return err
				}
				return env.CallVoidMethod(cursorObj, closeMid)
			})
			fmt.Fprintf(output, "  %s downloads: %d\n", statusFilter.name, count)
			vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(cursorObj); return nil })
		}
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(queryObj); return nil })
	}
	ui.RenderOutput()

	// --- Query all downloads (no filter) ---
	fmt.Fprintln(output, "\n=== All downloads ===")
	var allQueryObj *jni.GlobalRef
	err = vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("android/app/DownloadManager$Query")
		if err != nil {
			return err
		}
		mid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return err
		}
		obj, err := env.NewObject(cls, mid)
		if err != nil {
			return err
		}
		allQueryObj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		fmt.Fprintf(output, "Create all-query: %v\n", err)
	} else {
		cursorObj, err := mgr.Query(allQueryObj)
		if err != nil {
			fmt.Fprintf(output, "Query(all): %v\n", err)
		} else if cursorObj != nil {
			var count int32
			vm.Do(func(env *jni.Env) error {
				cursorCls := env.GetObjectClass(cursorObj)
				getCountMid, err := env.GetMethodID(cursorCls, "getCount", "()I")
				if err != nil {
					return err
				}
				count, err = env.CallIntMethod(cursorObj, getCountMid)
				if err != nil {
					return err
				}
				closeMid, err := env.GetMethodID(cursorCls, "close", "()V")
				if err != nil {
					return err
				}
				return env.CallVoidMethod(cursorObj, closeMid)
			})
			fmt.Fprintf(output, "Total downloads: %d\n", count)
			vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(cursorObj); return nil })
		}
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(allQueryObj); return nil })
	}
	ui.RenderOutput()

	// --- Print constants for reference ---
	fmt.Fprintln(output, "\n=== Download constants ===")
	fmt.Fprintf(output, "Status: Pending=%d Running=%d Paused=%d Successful=%d Failed=%d\n",
		download.StatusPending, download.StatusRunning, download.StatusPaused,
		download.StatusSuccessful, download.StatusFailed)
	fmt.Fprintf(output, "Network: Mobile=%d WiFi=%d\n",
		download.NetworkMobile, download.NetworkWifi)
	fmt.Fprintf(output, "Visibility: Hidden=%d Visible=%d NotifyDone=%d NotifyOnlyDone=%d\n",
		download.VisibilityHidden, download.VisibilityVisible,
		download.VisibilityVisibleNotifyCompleted, download.VisibilityVisibleNotifyOnlyCompletion)
	fmt.Fprintf(output, "Columns: ID=%s Title=%s Status=%s Size=%s URI=%s\n",
		download.ColumnId, download.ColumnTitle, download.ColumnStatus,
		download.ColumnTotalSizeBytes, download.ColumnUri)
	fmt.Fprintf(output, "Actions: Complete=%s Clicked=%s ViewDownloads=%s\n",
		download.ActionDownloadComplete, download.ActionNotificationClicked,
		download.ActionViewDownloads)
	fmt.Fprintf(output, "Errors: Unknown=%d FileError=%d HttpData=%d Space=%d Redirects=%d\n",
		download.ErrorUnknown, download.ErrorFileError, download.ErrorHttpDataError,
		download.ErrorInsufficientSpace, download.ErrorTooManyRedirects)
	fmt.Fprintf(output, "Paused: Unknown=%d WiFi=%d Network=%d Retry=%d\n",
		download.PausedUnknown, download.PausedQueuedForWifi,
		download.PausedWaitingForNetwork, download.PausedWaitingToRetry)

	fmt.Fprintln(output, "\ndownload_queue complete.")
	return nil
}
