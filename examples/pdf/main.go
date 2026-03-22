//go:build android

// Command pdf demonstrates the graphics/pdf package by creating a Bitmap,
// Canvas, and Paint, drawing shapes and text, and querying object properties.
// It also opens a minimal PDF file with PdfRenderer and queries pages.
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
	"os"
	"path/filepath"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/graphics/pdf"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(activity.vm)),
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
}

// minimalPDF is a valid single-page PDF (72x72 point, blank white).
var minimalPDF = []byte(`%PDF-1.0
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj
3 0 obj<</Type/Page/MediaBox[0 0 72 72]/Parent 2 0 R/Resources<<>>>>endobj
xref
0 4
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
trailer<</Size 4/Root 1 0 R>>
startxref
206
%%EOF
`)

func run(vm *jni.VM, output *bytes.Buffer) error {
	fmt.Fprintln(output, "=== Graphics/PDF Demo ===")
	ui.RenderOutput()

	// --- Bitmap/Canvas/Paint drawing API ---

	// Get ARGB_8888 bitmap config.
	argb8888, err := pdf.ARGB8888(vm)
	if err != nil {
		return fmt.Errorf("ARGB8888: %w", err)
	}
	fmt.Fprintf(output, "ARGB_8888: %v\n", argb8888)
	ui.RenderOutput()

	// Create a 200x200 Bitmap.
	bmpHelper := pdf.Bitmap{VM: vm}
	bitmapObj, err := bmpHelper.CreateBitmap3_10(200, 200, argb8888)
	if err != nil {
		return fmt.Errorf("CreateBitmap: %w", err)
	}
	bmp := pdf.Bitmap{VM: vm, Obj: bitmapObj}
	fmt.Fprintln(output, "Bitmap created OK")

	// Query bitmap dimensions and properties.
	bmpW, _ := bmp.GetWidth()
	bmpH, _ := bmp.GetHeight()
	fmt.Fprintf(output, "Bitmap: %dx%d\n", bmpW, bmpH)

	byteCount, _ := bmp.GetByteCount()
	fmt.Fprintf(output, "ByteCount: %d\n", byteCount)

	allocCount, _ := bmp.GetAllocationByteCount()
	fmt.Fprintf(output, "AllocByteCount: %d\n", allocCount)

	bmpDensity, _ := bmp.GetDensity()
	fmt.Fprintf(output, "Density: %d\n", bmpDensity)

	rowBytes, _ := bmp.GetRowBytes()
	fmt.Fprintf(output, "RowBytes: %d\n", rowBytes)

	hasAlpha, _ := bmp.HasAlpha()
	fmt.Fprintf(output, "HasAlpha: %v\n", hasAlpha)

	isMut, _ := bmp.IsMutable()
	fmt.Fprintf(output, "IsMutable: %v\n", isMut)

	isPremul, _ := bmp.IsPremultiplied()
	fmt.Fprintf(output, "IsPremultiplied: %v\n", isPremul)

	isRecycled, _ := bmp.IsRecycled()
	fmt.Fprintf(output, "IsRecycled: %v\n", isRecycled)

	genID, _ := bmp.GetGenerationId()
	fmt.Fprintf(output, "GenerationId: %d\n", genID)
	ui.RenderOutput()

	// Create a Canvas from the bitmap.
	canvas, err := pdf.NewCanvas(vm, bitmapObj)
	if err != nil {
		return fmt.Errorf("NewCanvas: %w", err)
	}
	defer vm.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(canvas.Obj)
		return nil
	})
	fmt.Fprintln(output, "Canvas created OK")

	cw, _ := canvas.GetWidth()
	ch, _ := canvas.GetHeight()
	fmt.Fprintf(output, "Canvas: %dx%d\n", cw, ch)

	cDensity, _ := canvas.GetDensity()
	fmt.Fprintf(output, "Canvas density: %d\n", cDensity)

	saveCount, _ := canvas.GetSaveCount()
	fmt.Fprintf(output, "SaveCount: %d\n", saveCount)

	maxBmpW, _ := canvas.GetMaximumBitmapWidth()
	maxBmpH, _ := canvas.GetMaximumBitmapHeight()
	fmt.Fprintf(output, "MaxBmp: %dx%d\n", maxBmpW, maxBmpH)

	isHwAccel, _ := canvas.IsHardwareAccelerated()
	fmt.Fprintf(output, "IsHwAccel: %v\n", isHwAccel)

	isOpaque, _ := canvas.IsOpaque()
	fmt.Fprintf(output, "IsOpaque: %v\n", isOpaque)
	ui.RenderOutput()

	// Create a Paint.
	paint, err := pdf.NewPaint(vm)
	if err != nil {
		return fmt.Errorf("NewPaint: %w", err)
	}
	defer vm.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(paint.Obj)
		return nil
	})
	fmt.Fprintln(output, "Paint created OK")

	pColor, _ := paint.GetColor()
	fmt.Fprintf(output, "Color: 0x%08X\n", uint32(pColor))

	pAlpha, _ := paint.GetAlpha()
	fmt.Fprintf(output, "Alpha: %d\n", pAlpha)

	pTextSize, _ := paint.GetTextSize()
	fmt.Fprintf(output, "TextSize: %.1f\n", pTextSize)

	pStrokeW, _ := paint.GetStrokeWidth()
	fmt.Fprintf(output, "StrokeWidth: %.1f\n", pStrokeW)

	pAA, _ := paint.IsAntiAlias()
	fmt.Fprintf(output, "IsAntiAlias: %v\n", pAA)

	pFlags, _ := paint.GetFlags()
	fmt.Fprintf(output, "Flags: 0x%X\n", pFlags)

	pFontSpacing, _ := paint.GetFontSpacing()
	fmt.Fprintf(output, "FontSpacing: %.1f\n", pFontSpacing)
	ui.RenderOutput()

	// Fill background with white.
	_ = canvas.DrawColor1(-1) // 0xFFFFFFFF
	fmt.Fprintln(output, "DrawColor(white): OK")

	// Draw a red rectangle.
	_ = paint.SetColor1(-65536) // 0xFFFF0000
	_ = canvas.DrawRect5_2(10, 10, 100, 80, paint.Obj)
	fmt.Fprintln(output, "DrawRect(red): OK")

	// Draw a green circle.
	_ = paint.SetColor1(-16711936) // 0xFF00FF00
	_ = canvas.DrawCircle(150, 50, 40, paint.Obj)
	fmt.Fprintln(output, "DrawCircle(green): OK")

	// Draw a blue diagonal line.
	_ = paint.SetColor1(-16776961) // 0xFF0000FF
	_ = paint.SetStrokeWidth(3.0)
	_ = canvas.DrawLine(0, 0, 200, 200, paint.Obj)
	fmt.Fprintln(output, "DrawLine(blue): OK")

	// Draw text.
	_ = paint.SetColor1(-16777216) // 0xFF000000
	_ = paint.SetTextSize(20)
	_ = paint.SetAntiAlias(true)
	_ = canvas.DrawText4_2("Hello JNI", 20, 150, paint.Obj)
	fmt.Fprintln(output, "DrawText: OK")
	ui.RenderOutput()

	// Measure text.
	measured, _ := paint.MeasureText1_2("Hello JNI")
	fmt.Fprintf(output, "MeasureText: %.1f px\n", measured)

	// Save/restore canvas state.
	sc, _ := canvas.Save()
	fmt.Fprintf(output, "Save: count=%d\n", sc)
	_ = canvas.Restore()
	fmt.Fprintln(output, "Restore: OK")

	// Read pixel at (50,50) -- inside the red rect.
	px, _ := bmp.GetPixel(50, 50)
	fmt.Fprintf(output, "Pixel(50,50): 0x%08X\n", uint32(px))

	genID2, _ := bmp.GetGenerationId()
	fmt.Fprintf(output, "GenId(after draw): %d\n", genID2)

	// Recycle bitmap.
	_ = bmp.Recycle()
	isRecycled2, _ := bmp.IsRecycled()
	fmt.Fprintf(output, "IsRecycled(after): %v\n", isRecycled2)
	ui.RenderOutput()

	// --- PdfRenderer API ---

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "--- PdfRenderer ---")

	// Get cache directory for temp file.
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		fmt.Fprintf(output, "GetAppContext: %v\n", err)
		fmt.Fprintln(output, "PDF example complete.")
		return nil
	}
	defer ctx.Close()

	var cachePath string
	vm.Do(func(env *jni.Env) error {
		cacheDir, err := ctx.GetCacheDir()
		if err != nil || cacheDir == nil {
			return err
		}
		cls := env.GetObjectClass(cacheDir)
		mid, err := env.GetMethodID(cls, "getAbsolutePath", "()Ljava/lang/String;")
		if err != nil {
			return err
		}
		pathObj, err := env.CallObjectMethod(cacheDir, mid)
		if err != nil {
			return err
		}
		cachePath = env.GoString((*jni.String)(unsafe.Pointer(pathObj)))
		return nil
	})

	if cachePath == "" {
		fmt.Fprintln(output, "No cache dir available")
		return nil
	}

	// Write minimal PDF to temp file.
	pdfPath := filepath.Join(cachePath, "test.pdf")
	if err := os.WriteFile(pdfPath, minimalPDF, 0644); err != nil {
		return fmt.Errorf("write PDF: %w", err)
	}
	defer os.Remove(pdfPath)
	fmt.Fprintf(output, "Wrote PDF: %s\n", pdfPath)
	ui.RenderOutput()

	// Open with ParcelFileDescriptor.
	var pfdObj *jni.Object
	err = vm.Do(func(env *jni.Env) error {
		fileCls, err := env.FindClass("java/io/File")
		if err != nil {
			return fmt.Errorf("find File: %w", err)
		}
		fileInit, err := env.GetMethodID(fileCls, "<init>", "(Ljava/lang/String;)V")
		if err != nil {
			return fmt.Errorf("get File.<init>: %w", err)
		}
		jPath, err := env.NewStringUTF(pdfPath)
		if err != nil {
			return err
		}
		fileObj, err := env.NewObject(fileCls, fileInit, jni.ObjectValue(&jPath.Object))
		if err != nil {
			return fmt.Errorf("new File: %w", err)
		}
		pfdCls, err := env.FindClass("android/os/ParcelFileDescriptor")
		if err != nil {
			return fmt.Errorf("find PFD: %w", err)
		}
		openMid, err := env.GetStaticMethodID(pfdCls, "open",
			"(Ljava/io/File;I)Landroid/os/ParcelFileDescriptor;")
		if err != nil {
			return fmt.Errorf("get PFD.open: %w", err)
		}
		pfdLocal, err := env.CallStaticObjectMethod(pfdCls, openMid,
			jni.ObjectValue(fileObj), jni.IntValue(int32(pdf.ModeReadOnly)))
		if err != nil {
			return fmt.Errorf("PFD.open: %w", err)
		}
		pfdObj = env.NewGlobalRef(pfdLocal)
		return nil
	})
	if err != nil {
		return fmt.Errorf("open PFD: %w", err)
	}
	fmt.Fprintln(output, "PFD opened OK")

	// Query PFD properties.
	pfd := pdf.ParcelFileDescriptor{VM: vm, Obj: pfdObj}
	pfdSize, _ := pfd.GetStatSize()
	fmt.Fprintf(output, "PDF size: %d bytes\n", pfdSize)

	pfdFd, _ := pfd.GetFd()
	fmt.Fprintf(output, "FD: %d\n", pfdFd)
	ui.RenderOutput()

	// Create PdfRenderer.
	var renderer pdf.Renderer
	renderer.VM = vm
	err = vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("android/graphics/pdf/PdfRenderer")
		if err != nil {
			return fmt.Errorf("find PdfRenderer: %w", err)
		}
		initMid, err := env.GetMethodID(cls, "<init>",
			"(Landroid/os/ParcelFileDescriptor;)V")
		if err != nil {
			return fmt.Errorf("get PdfRenderer.<init>: %w", err)
		}
		obj, err := env.NewObject(cls, initMid, jni.ObjectValue(pfdObj))
		if err != nil {
			return fmt.Errorf("new PdfRenderer: %w", err)
		}
		renderer.Obj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		return fmt.Errorf("create renderer: %w", err)
	}
	fmt.Fprintln(output, "PdfRenderer created OK")

	// Query page count.
	pageCount, _ := renderer.GetPageCount()
	fmt.Fprintf(output, "Pages: %d\n", pageCount)

	shouldScale, _ := renderer.ShouldScaleForPrinting()
	fmt.Fprintf(output, "ScaleForPrint: %v\n", shouldScale)
	ui.RenderOutput()

	// Open page 0 and query dimensions.
	if pageCount > 0 {
		pageObj, err := renderer.OpenPage(0)
		if err != nil {
			fmt.Fprintf(output, "OpenPage(0): %v\n", err)
		} else {
			page := pdf.RendererPage{VM: vm, Obj: pageObj}

			idx, _ := page.GetIndex()
			fmt.Fprintf(output, "Page index: %d\n", idx)

			pw, _ := page.GetWidth()
			ph, _ := page.GetHeight()
			fmt.Fprintf(output, "Page: %dx%d\n", pw, ph)

			// Render to a bitmap.
			argb2, _ := pdf.ARGB8888(vm)
			bmpObj2, err := bmpHelper.CreateBitmap3_10(72, 72, argb2)
			if err != nil {
				fmt.Fprintf(output, "CreateBitmap: %v\n", err)
			} else {
				err = page.Render4_1(bmpObj2, nil, nil, int32(pdf.RenderModeForDisplay))
				if err != nil {
					fmt.Fprintf(output, "Render: %v\n", err)
				} else {
					fmt.Fprintln(output, "Rendered page 0 OK")
				}
				bmp2 := pdf.Bitmap{VM: vm, Obj: bmpObj2}
				_ = bmp2.Recycle()
			}

			_ = page.Close()
			fmt.Fprintln(output, "Page closed")
		}
	}

	// Close renderer.
	_ = renderer.Close()
	fmt.Fprintln(output, "Renderer closed")

	// Clean up global refs.
	vm.Do(func(env *jni.Env) error {
		if renderer.Obj != nil {
			env.DeleteGlobalRef(renderer.Obj)
		}
		if pfdObj != nil {
			env.DeleteGlobalRef(pfdObj)
		}
		return nil
	})

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "PDF example complete.")
	return nil
}
