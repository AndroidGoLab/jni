//go:build android

// Command resolver_contacts queries the Android contacts content provider
// (content://com.android.contacts/contacts) via ContentResolver and lists
// contact display names using cursor iteration.
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

	fmt.Fprintln(output, "=== Contacts ===")

	resolverObj, err := ctx.GetContentResolver()
	if err != nil {
		return fmt.Errorf("GetContentResolver: %w", err)
	}

	cr := resolver.ContentResolver{VM: vm, Obj: resolverObj}

	uriHelper := resolver.Uri{VM: vm}
	contactsURI, err := uriHelper.Parse("content://com.android.contacts/contacts")
	if err != nil {
		return fmt.Errorf("Uri.Parse: %w", err)
	}

	cursorObj, err := cr.Query4(contactsURI, nil, nil, nil)
	if err != nil {
		fmt.Fprintf(output, "Query contacts: %v\n", err)
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
	fmt.Fprintf(output, "Total contacts: %d\n", count)

	colCount, err := cursor.GetColumnCount()
	if err != nil {
		return fmt.Errorf("cursor.GetColumnCount: %w", err)
	}
	fmt.Fprintf(output, "Columns: %d\n", colCount)

	// Find the display_name column index.
	displayNameIdx, err := cursor.GetColumnIndex("display_name")
	if err != nil {
		return fmt.Errorf("GetColumnIndex(display_name): %w", err)
	}
	fmt.Fprintf(output, "display_name col: %d\n", displayNameIdx)
	fmt.Fprintln(output)

	// Iterate through contacts.
	ok, err := cursor.MoveToFirst()
	if err != nil || !ok {
		fmt.Fprintln(output, "(no contacts)")
		return nil
	}

	maxRows := int32(20)
	if count < maxRows {
		maxRows = count
	}

	fmt.Fprintf(output, "First %d contacts:\n", maxRows)
	for row := int32(0); row < maxRows; row++ {
		name, err := cursor.GetString(displayNameIdx)
		if err != nil {
			name = "(err)"
		}
		fmt.Fprintf(output, "  [%d] %s\n", row+1, name)

		if row < maxRows-1 {
			moved, err := cursor.MoveToNext()
			if err != nil || !moved {
				break
			}
		}
	}

	return nil
}
