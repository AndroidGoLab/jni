//go:build android

// Command resolver_call_log queries the Android call log content provider
// (content://call_log/calls). It reads phone number, duration, and type
// (incoming/outgoing/missed) and computes basic statistics.
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

func callTypeName(t int32) string {
	switch t {
	case 1:
		return "incoming"
	case 2:
		return "outgoing"
	case 3:
		return "missed"
	case 4:
		return "voicemail"
	case 5:
		return "rejected"
	case 6:
		return "blocked"
	default:
		return fmt.Sprintf("type(%d)", t)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Call Log ===")

	resolverObj, err := ctx.GetContentResolver()
	if err != nil {
		return fmt.Errorf("GetContentResolver: %w", err)
	}

	cr := resolver.ContentResolver{VM: vm, Obj: resolverObj}

	uriHelper := resolver.Uri{VM: vm}
	callsURI, err := uriHelper.Parse("content://call_log/calls")
	if err != nil {
		return fmt.Errorf("Uri.Parse: %w", err)
	}

	cursorObj, err := cr.Query4(callsURI, nil, nil, nil)
	if err != nil {
		fmt.Fprintf(output, "Query call log: %v\n", err)
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
	fmt.Fprintf(output, "Total call records: %d\n", count)

	// Find column indices.
	numberIdx, _ := cursor.GetColumnIndex("number")
	durationIdx, _ := cursor.GetColumnIndex("duration")
	typeIdx, _ := cursor.GetColumnIndex("type")

	fmt.Fprintf(output, "Columns: number=%d, duration=%d, type=%d\n",
		numberIdx, durationIdx, typeIdx)
	fmt.Fprintln(output)

	ok, err := cursor.MoveToFirst()
	if err != nil || !ok {
		fmt.Fprintln(output, "(no call records)")
		return nil
	}

	// Iterate and collect stats.
	maxRows := int32(15)
	if count < maxRows {
		maxRows = count
	}

	var totalDuration int64
	var incoming, outgoing, missed int32

	fmt.Fprintf(output, "Recent %d calls:\n", maxRows)
	for row := int32(0); row < maxRows; row++ {
		number, _ := cursor.GetString(numberIdx)
		duration, _ := cursor.GetInt(durationIdx)
		callType, _ := cursor.GetInt(typeIdx)

		totalDuration += int64(duration)
		switch callType {
		case 1:
			incoming++
		case 2:
			outgoing++
		case 3:
			missed++
		}

		fmt.Fprintf(output, "  [%d] %s (%s, %ds)\n",
			row+1, number, callTypeName(callType), duration)

		if row < maxRows-1 {
			moved, err := cursor.MoveToNext()
			if err != nil || !moved {
				break
			}
		}
	}

	// Stats summary.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Stats (sampled):")
	fmt.Fprintf(output, "  incoming: %d\n", incoming)
	fmt.Fprintf(output, "  outgoing: %d\n", outgoing)
	fmt.Fprintf(output, "  missed:   %d\n", missed)
	fmt.Fprintf(output, "  total duration: %ds\n", totalDuration)
	if maxRows > 0 {
		fmt.Fprintf(output, "  avg duration: %ds\n", totalDuration/int64(maxRows))
	}

	return nil
}
