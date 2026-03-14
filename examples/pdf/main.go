//go:build android

// Command pdf demonstrates the constants and data structures provided by
// the pdf package, which wraps android.graphics.pdf.PdfRenderer.
//
// PdfRenderer requires a ParcelFileDescriptor backed by a real PDF file,
// so this example does not instantiate a Renderer. Instead it shows the
// available constants and documents the API surface.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/xaionaro-go/jni/graphics/pdf"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(_ *C.JavaVM) {
	run()
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}

func run() {
	// --- Constants ---
	fmt.Fprintln(&output, "Render modes:")
	fmt.Fprintf(&output, "  RenderModeForDisplay = %d\n", pdf.RenderModeForDisplay)
	fmt.Fprintf(&output, "  RenderModeForPrint   = %d\n", pdf.RenderModeForPrint)
	fmt.Fprintf(&output, "  ModeReadOnly         = %d (0x%X)\n", pdf.ModeReadOnly, pdf.ModeReadOnly)

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

	fmt.Fprintln(&output, "PDF example complete.")
}
