//go:build android

// Command pdf demonstrates the constants and data structures provided by
// the pdf package, which wraps android.graphics.pdf.PdfRenderer.
//
// PdfRenderer requires a ParcelFileDescriptor backed by a real PDF file,
// so this example does not instantiate a Renderer. Instead it shows the
// available constants and documents the API surface.
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
	"unsafe"
	"bytes"
	"fmt"

	"github.com/AndroidGoLab/jni/graphics/pdf"
	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
)

func main() {}

func init() { exampleui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	exampleui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	exampleui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	exampleui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	// --- Constants ---
	fmt.Fprintln(output, "Render modes:")
	fmt.Fprintf(output, "  RenderModeForDisplay = %d\n", pdf.RenderModeForDisplay)
	fmt.Fprintf(output, "  RenderModeForPrint   = %d\n", pdf.RenderModeForPrint)
	fmt.Fprintf(output, "  ModeReadOnly         = %d (0x%X)\n", pdf.ModeReadOnly, pdf.ModeReadOnly)

	// --- Renderer ---
	// To create a Renderer you need a ParcelFileDescriptor pointing at a
	// real PDF file:
	//
	//   renderer, err := pdf.NewRenderer(vm)
	//   defer renderer.Close()
	//
	//   pageCount, _ := renderer.GetPageCount()
	//   renderer.closeRaw()  // closes the Java-side document

	// --- Page ---
	// Pages are opened via renderer.openPageRaw(index) (unexported).
	// Each Page exposes exported methods:
	//
	//   page.Close()       - releases the Go-side JNI reference
	//   page.GetWidth()    int32
	//   page.GetHeight()   int32
	//   page.renderRaw(bitmap, destClip, transform, renderMode) [unexported]
	//   page.closeRaw()    [unexported, closes Java-side]
	//
	// Render modes for page.renderRaw:
	//   pdf.RenderModeForDisplay (screen rendering)
	//   pdf.RenderModeForPrint   (print-quality rendering)

	// --- ParcelFileDescriptor ---
	// The parcelFileDescriptor type (unexported) provides:
	//   openRaw(file, mode) to open a file descriptor.
	// Use pdf.ModeReadOnly to open a PDF file in read-only mode.

	// --- Bitmap ---
	// The bitmap type (unexported) provides:
	//   createBitmapRaw(width, height, config) to create a rendering target.
	//   copyPixelsToBufferRaw(buffer) to extract pixel data.

	fmt.Fprintln(output, "PDF example complete.")
	return nil
}
