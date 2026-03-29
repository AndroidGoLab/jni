//go:build android

// Command resolver_sms queries the Android SMS content provider
// (content://sms/inbox). It reads sender address, body text, and date
// for recent messages.
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
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/content/resolver"
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

	fmt.Fprintln(output, "=== SMS Inbox ===")

	resolverObj, err := ctx.GetContentResolver()
	if err != nil {
		return fmt.Errorf("GetContentResolver: %w", err)
	}

	cr := resolver.ContentResolver{VM: vm, Obj: resolverObj}

	uriHelper := resolver.Uri{VM: vm}
	smsURI, err := uriHelper.Parse("content://sms/inbox")
	if err != nil {
		return fmt.Errorf("Uri.Parse: %w", err)
	}

	cursorObj, err := cr.Query4(smsURI, nil, nil, nil)
	if err != nil {
		fmt.Fprintf(output, "Query SMS: %v\n", err)
		return nil
	}
	if cursorObj == nil {
		fmt.Fprintln(output, "Query returned null cursor")
		return nil
	}

	cursor := resolver.Cursor{VM: vm, Obj: cursorObj}
	defer func() { _ = cursor.Close() }()

	count, err := cursor.GetCount()
	if err != nil {
		return fmt.Errorf("cursor.GetCount: %w", err)
	}
	fmt.Fprintf(output, "Total inbox messages: %d\n", count)

	// Print column names.
	colCount, err := cursor.GetColumnCount()
	if err == nil {
		fmt.Fprintf(output, "Columns: %d\n", colCount)
		for c := int32(0); c < colCount && c < 10; c++ {
			name, err := cursor.GetColumnName(c)
			if err != nil {
				continue
			}
			fmt.Fprintf(output, "  col[%d]: %s\n", c, name)
		}
	}
	fmt.Fprintln(output)

	// Find column indices.
	addressIdx, _ := cursor.GetColumnIndex("address")
	bodyIdx, _ := cursor.GetColumnIndex("body")
	dateIdx, _ := cursor.GetColumnIndex("date")

	ok, err := cursor.MoveToFirst()
	if err != nil || !ok {
		fmt.Fprintln(output, "(no messages)")
		return nil
	}

	maxRows := int32(10)
	if count < maxRows {
		maxRows = count
	}

	fmt.Fprintf(output, "Recent %d messages:\n", maxRows)
	for row := int32(0); row < maxRows; row++ {
		address, _ := cursor.GetString(addressIdx)
		body, _ := cursor.GetString(bodyIdx)
		dateLong, _ := cursor.GetLong(dateIdx)

		// Truncate body for display.
		if len(body) > 60 {
			body = body[:60] + "..."
		}

		dateStr := time.UnixMilli(dateLong).Format("2006-01-02 15:04")

		fmt.Fprintf(output, "  [%d] From: %s\n", row+1, address)
		fmt.Fprintf(output, "       Date: %s\n", dateStr)
		fmt.Fprintf(output, "       Body: %s\n", body)

		if row < maxRows-1 {
			moved, err := cursor.MoveToNext()
			if err != nil || !moved {
				break
			}
		}
	}

	return nil
}
