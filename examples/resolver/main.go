//go:build android

// Command resolver demonstrates querying content providers via the
// Android ContentResolver, wrapped by the resolver package.
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
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/content/resolver"
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

	fmt.Fprintln(output, "=== ContentResolver ===")

	// Get the ContentResolver from the app context.
	resolverObj, err := ctx.GetContentResolver()
	if err != nil {
		return fmt.Errorf("GetContentResolver: %w", err)
	}

	cr := resolver.ContentResolver{VM: vm, Obj: resolverObj}

	// Parse a content URI for the Settings.System provider,
	// which is accessible without special permissions.
	uriHelper := resolver.Uri{VM: vm}
	settingsURI, err := uriHelper.Parse("content://settings/system")
	if err != nil {
		return fmt.Errorf("Uri.Parse: %w", err)
	}

	// Query with a Bundle argument (API 26+): pass nil for all args.
	cursorObj, err := cr.Query4(settingsURI, nil, nil, nil)
	if err != nil {
		fmt.Fprintf(output, "Query settings: %v\n", err)
		return nil
	}
	if cursorObj == nil {
		fmt.Fprintln(output, "Query returned null cursor")
		return nil
	}

	cursor := resolver.Cursor{
		VM:  vm,
		Obj: cursorObj,
	}
	defer func() { _ = cursor.Close() }()

	count, err := cursor.GetCount()
	if err != nil {
		return fmt.Errorf("cursor.GetCount: %w", err)
	}

	colCount, err := cursor.GetColumnCount()
	if err != nil {
		return fmt.Errorf("cursor.GetColumnCount: %w", err)
	}

	fmt.Fprintf(output, "Settings rows: %d\n", count)
	fmt.Fprintf(output, "Columns: %d\n", colCount)

	// Print column names.
	for c := int32(0); c < colCount; c++ {
		name, err := cursor.GetColumnName(c)
		if err != nil {
			continue
		}
		fmt.Fprintf(output, "  col[%d]: %s\n", c, name)
	}

	// Print first few rows.
	maxRows := int32(5)
	if count < maxRows {
		maxRows = count
	}

	ok, err := cursor.MoveToFirst()
	if err != nil || !ok {
		return nil
	}

	fmt.Fprintf(output, "\nFirst %d rows:\n", maxRows)
	for row := int32(0); row < maxRows; row++ {
		var vals []string
		for c := int32(0); c < colCount; c++ {
			s, err := cursor.GetString(c)
			if err != nil {
				s = "(err)"
			}
			vals = append(vals, s)
		}
		fmt.Fprintf(output, "  %v\n", vals)

		if row < maxRows-1 {
			moved, err := cursor.MoveToNext()
			if err != nil || !moved {
				break
			}
		}
	}

	return nil
}
