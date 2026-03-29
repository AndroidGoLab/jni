//go:build android

// Command mediastore_gallery queries the Android MediaStore for images.
// It lists image file names, sizes, and dates, showing the first N entries
// as a gallery manifest.
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

func humanSize(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== MediaStore Gallery ===")

	resolverObj, err := ctx.GetContentResolver()
	if err != nil {
		return fmt.Errorf("GetContentResolver: %w", err)
	}

	cr := resolver.ContentResolver{VM: vm, Obj: resolverObj}

	// Query the external images MediaStore content URI.
	uriHelper := resolver.Uri{VM: vm}
	imagesURI, err := uriHelper.Parse("content://media/external/images/media")
	if err != nil {
		return fmt.Errorf("Uri.Parse: %w", err)
	}

	cursorObj, err := cr.Query4(imagesURI, nil, nil, nil)
	if err != nil {
		fmt.Fprintf(output, "Query images: %v\n", err)
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
	fmt.Fprintf(output, "Total images: %d\n", count)
	fmt.Fprintln(output)

	// Find column indices.
	nameIdx, _ := cursor.GetColumnIndex("_display_name")
	sizeIdx, _ := cursor.GetColumnIndex("_size")
	dateIdx, _ := cursor.GetColumnIndex("date_added")
	widthIdx, _ := cursor.GetColumnIndex("width")
	heightIdx, _ := cursor.GetColumnIndex("height")
	mimeIdx, _ := cursor.GetColumnIndex("mime_type")

	ok, err := cursor.MoveToFirst()
	if err != nil || !ok {
		fmt.Fprintln(output, "(no images)")
		return nil
	}

	maxRows := int32(15)
	if count < maxRows {
		maxRows = count
	}

	var totalSize int64

	fmt.Fprintf(output, "Gallery manifest (%d entries):\n", maxRows)
	fmt.Fprintln(output)
	for row := int32(0); row < maxRows; row++ {
		name, _ := cursor.GetString(nameIdx)
		size, _ := cursor.GetLong(sizeIdx)
		dateAdded, _ := cursor.GetLong(dateIdx)
		width, _ := cursor.GetInt(widthIdx)
		height, _ := cursor.GetInt(heightIdx)
		mime, _ := cursor.GetString(mimeIdx)

		totalSize += size

		dateStr := time.Unix(dateAdded, 0).Format("2006-01-02")

		fmt.Fprintf(output, "  [%d] %s\n", row+1, name)
		fmt.Fprintf(output, "       %s | %dx%d | %s | %s\n",
			humanSize(size), width, height, mime, dateStr)

		if row < maxRows-1 {
			moved, err := cursor.MoveToNext()
			if err != nil || !moved {
				break
			}
		}
	}

	fmt.Fprintln(output)
	fmt.Fprintf(output, "Total size (sampled %d): %s\n", maxRows, humanSize(totalSize))

	return nil
}
