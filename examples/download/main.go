//go:build android

// Command download demonstrates using the Android DownloadManager API.
//
// This example obtains the DownloadManager system service using the exported
// NewManager constructor and Close cleanup method, and shows how to use the
// exported status constants. The actual download operations (enqueueRaw,
// removeRaw, queryRaw) and helper types (downloadRequest, downloadQuery,
// cursor) are unexported and used internally by higher-level wrappers.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/app"
	"github.com/xaionaro-go/jni/app/download"
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

	// NewManager obtains the DownloadManager system service.
	mgr, err := download.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("download.NewManager: %w", err)
	}
	// Close releases the JNI global reference; always defer it.
	defer mgr.Close()

	fmt.Fprintln(&output, "DownloadManager obtained successfully")

	// Exported status constants for download state queries.
	fmt.Fprintln(&output, "Download status constants:")
	fmt.Fprintf(&output, "  StatusPending:    %d\n", download.StatusPending)
	fmt.Fprintf(&output, "  StatusRunning:    %d\n", download.StatusRunning)
	fmt.Fprintf(&output, "  StatusPaused:     %d\n", download.StatusPaused)
	fmt.Fprintf(&output, "  StatusSuccessful: %d\n", download.StatusSuccessful)
	fmt.Fprintf(&output, "  StatusFailed:     %d\n", download.StatusFailed)

	// The following methods and types are unexported (package-internal):
	//
	// Manager methods:
	//   mgr.enqueueRaw(request *jni.Object) (int64, error)
	//     Enqueues a new download and returns the download ID.
	//
	//   mgr.removeRaw(ids []int64) (int32, error)
	//     Removes downloads by their IDs; returns the count removed.
	//
	//   mgr.queryRaw(query *jni.Object) (*jni.Object, error)
	//     Queries the download manager for download status via a cursor.
	//
	// downloadRequest wraps DownloadManager.Request:
	//   NewdownloadRequest(vm *jni.VM) (*downloadRequest, error)
	//   downloadRequest.setDestinationInExternalPublicDir(dirType, subPath string)
	//   downloadRequest.setTitle(title *jni.Object)
	//   downloadRequest.setDescription(description *jni.Object)
	//   downloadRequest.setMimeType(mimeType string)
	//
	// downloadQuery wraps DownloadManager.Query:
	//   NewdownloadQuery(vm *jni.VM) (*downloadQuery, error)
	//   downloadQuery.setFilterById(ids []int64)
	//
	// cursor wraps android.database.Cursor:
	//   cursor.moveToFirst() bool
	//   cursor.getColumnIndex(columnName string) int32
	//   cursor.getInt(columnIndex int32) int32
	//   cursor.close()

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
