//go:build android

// Command resolver demonstrates querying content providers via the
// Android ContentResolver, wrapped by the resolver package. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// The resolver package wraps android.content.ContentResolver and
// android.database.Cursor. It provides Cursor methods for iterating
// query results and reading column values. The Resolver and Cursor
// types require proper resource cleanup via Close().
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/content/resolver"
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

	// The resolver package wraps android.content.ContentResolver.
	//
	// Resolver provides unexported methods for content provider access:
	//   queryRaw(uri, projection, selection, selectionArgs, sortOrder)
	//     -- queries a content provider and returns a raw Cursor JNI object.
	//   openFileDescriptorRaw(uri, mode)
	//     -- opens a file descriptor to a content URI.
	//
	// These are intended to be wrapped by higher-level helpers.

	// Cursor wraps android.database.Cursor with exported methods
	// for reading query results:
	//   Close()                            -- releases the JNI global reference.
	//   GetString(columnIndex int32)       -- reads a string column.
	//   GetInt(columnIndex int32)          -- reads an int32 column.
	//   GetLong(columnIndex int32)         -- reads an int64 column.
	//   GetColumnIndex(columnName string)  -- resolves a column name to its index.
	//
	// The unexported moveToNext() advances the cursor to the next row.

	// Demonstrate Cursor exported method signatures.
	// In a real app, the cursor would be obtained from a query via
	// higher-level helpers built on top of the Resolver.
	fmt.Fprintln(&output, "Resolver API: queryRaw, openFileDescriptorRaw")
	fmt.Fprintln(&output, "Cursor API: Close, GetString, GetInt, GetLong, GetColumnIndex")

	// The parcelFD type (unexported) wraps android.os.ParcelFileDescriptor
	// with methods getFd() and detachFd() for obtaining raw file descriptors
	// from content URIs opened via openFileDescriptorRaw.
	fmt.Fprintln(&output, "ParcelFD API: getFd, detachFd")

	// The Resolver and Uri types are the main exported types.
	var r resolver.Resolver
	var u resolver.Uri
	_, _ = r, u
	fmt.Fprintln(&output, "Resolver example complete.")
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
