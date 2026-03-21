//go:build android

// Package exampleui provides the shared Activity lifecycle handling for
// JNI examples. It creates a text display on NativeActivity's rendering
// surface using Canvas via JNI — no Java code, no Android widgets.
package exampleui

import (
	"bytes"
	"fmt"
	"strings"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/graphics/pdf"
)

/*
#include <android/native_activity.h>
#include <android/native_window.h>
#include <string.h>

// Render a bitmap's pixels onto the native window surface.
static int renderBitmapToWindow(ANativeWindow* window, void* pixels, int width, int height) {
    ANativeWindow_setBuffersGeometry(window, width, height, AHARDWAREBUFFER_FORMAT_R8G8B8A8_UNORM);
    ANativeWindow_Buffer buf;
    if (ANativeWindow_lock(window, &buf, NULL) != 0) {
        return -1;
    }
    // Copy row by row (buf.stride may differ from width).
    for (int y = 0; y < height && y < buf.height; y++) {
        memcpy((char*)buf.bits + y * buf.stride * 4,
               (char*)pixels + y * width * 4,
               width * 4 < buf.stride * 4 ? width * 4 : buf.stride * 4);
    }
    ANativeWindow_unlockAndPost(window);
    return 0;
}
*/
import "C"

const (
	// Android ARGB color constants.
	colorWhite  = -1          // 0xFFFFFFFF
	colorDkGray = -12303292   // 0xFF444444
	textSize    = float32(40) // sp
	lineHeight  = float32(48) // px between text lines
	leftMargin  = float32(32) // px from left edge
	topMargin   = float32(120) // px from top (below status bar)
	bottomPad   = 100          // px reserved at bottom

	// PackageManager.GET_META_DATA
	pmGetMetaData = 128

	// Permission request code (arbitrary).
	permissionRequestCode = 1
)

// RunFunc is the example's entry point.
type RunFunc func(vm *jni.VM, output *bytes.Buffer) error

var (
	runFunc          RunFunc
	vm               *jni.VM
	activityRef      *jni.Object
	nativeWindow     *C.ANativeWindow
	outputBuf        bytes.Buffer
	exampleStarted   bool
	permissionsAsked bool
	windowWidth      int
	windowHeight     int
)

// Register sets the example's run function. Call from init().
func Register(fn RunFunc) {
	runFunc = fn
}

// OnCreate is called when the NativeActivity is created.
func OnCreate(
	cvm *jni.VM,
	activity *jni.Object,
) {
	vm = cvm
	activityRef = activity
	outputBuf.Reset()
	exampleStarted = false
}

// OnNativeWindowCreated is called when the rendering surface is ready.
func OnNativeWindowCreated(windowPtr unsafe.Pointer) {
	nativeWindow = (*C.ANativeWindow)(windowPtr)
	windowWidth = int(C.ANativeWindow_getWidth(nativeWindow))
	windowHeight = int(C.ANativeWindow_getHeight(nativeWindow))

	renderText("Running example…")

	if exampleStarted {
		return
	}

	// Check and request permissions before running.
	if activityRef != nil && !permissionsAsked {
		var needed []string
		vm.Do(func(env *jni.Env) error {
			var err error
			needed, err = getUngrantedPermissions(env, activityRef)
			if err != nil {
				fmt.Fprintf(&outputBuf, "permissions check: %v\n", err)
			}
			return nil
		})
		if len(needed) > 0 {
			permissionsAsked = true
			renderText("Requesting permissions…")
			vm.Do(func(env *jni.Env) error {
				requestPermissions(env, activityRef, needed)
				return nil
			})
			return
		}
	}

	startExample()
}

func startExample() {
	if exampleStarted {
		return
	}
	exampleStarted = true

	go func() {
		if runFunc != nil {
			if err := runFunc(vm, &outputBuf); err != nil {
				fmt.Fprintf(&outputBuf, "ERROR: %v\n", err)
			}
		}
		RenderOutput()
	}()
}

// RenderOutput re-renders the current output buffer to the screen.
// Call from background goroutines after appending to the shared buffer.
func RenderOutput() {
	text := outputBuf.String()
	if text == "" {
		text = "(no output)"
	}
	renderText(text)
}

// OnResume is called when the activity resumes (e.g. after permission dialog).
func OnResume(activity *jni.Object) {
	if nativeWindow != nil && outputBuf.Len() > 0 {
		renderText(outputBuf.String())
	}

	// After permission dialog, try to start the example.
	if permissionsAsked && !exampleStarted && nativeWindow != nil {
		startExample()
	}
}

// renderText draws text onto the NativeActivity surface using a JNI
// Bitmap + Canvas for text layout, then blits to ANativeWindow.
func renderText(text string) {
	if vm == nil || nativeWindow == nil || windowWidth == 0 || windowHeight == 0 {
		return
	}

	// Get ARGB_8888 config for bitmap creation.
	argb8888, err := pdf.ARGB8888(vm)
	if err != nil {
		return
	}

	// Create Bitmap.createBitmap(width, height, ARGB_8888).
	// Static methods use a zero-value receiver for the VM reference.
	bmpHelper := pdf.Bitmap{VM: vm}
	bitmapObj, err := bmpHelper.CreateBitmap3_10(
		int32(windowWidth), int32(windowHeight), argb8888,
	)
	if err != nil {
		return
	}
	bmp := pdf.Bitmap{VM: vm, Obj: bitmapObj}

	// Create Canvas from Bitmap.
	canvas, err := pdf.NewCanvas(vm, bitmapObj)
	if err != nil {
		return
	}

	// Fill background white.
	if err := canvas.DrawColor1(colorWhite); err != nil {
		return
	}

	// Create Paint and configure it.
	paint, err := pdf.NewPaint(vm)
	if err != nil {
		return
	}
	if err := paint.SetColor1(colorDkGray); err != nil {
		return
	}
	if err := paint.SetTextSize(textSize); err != nil {
		return
	}
	if err := paint.SetAntiAlias(true); err != nil {
		return
	}

	// Set monospace typeface (best-effort; ignore errors on older APIs).
	if monoObj, err := pdf.MonospaceTypeface(vm); err == nil && monoObj != nil {
		_, _ = paint.SetTypeface(monoObj)
	}

	// Draw each line of text.
	lines := strings.Split(text, "\n")
	y := topMargin
	for _, line := range lines {
		if line == "" {
			y += lineHeight
			continue
		}
		_ = canvas.DrawText4_2(line, leftMargin, y, paint.Obj)
		y += lineHeight
		if y > float32(windowHeight-bottomPad) {
			break
		}
	}

	// Extract bitmap pixels and blit to ANativeWindow.
	// NewIntArray and GetIntArrayRegion require direct env access.
	pixelCount := windowWidth * windowHeight
	intBuf := make([]int32, pixelCount)

	vm.Do(func(env *jni.Env) error {
		pixelArray := env.NewIntArray(int32(pixelCount))
		if pixelArray == nil {
			return fmt.Errorf("NewIntArray returned nil")
		}
		if err := bmp.GetPixels(
			&pixelArray.Object,
			0, int32(windowWidth),
			0, 0,
			int32(windowWidth), int32(windowHeight),
		); err != nil {
			return err
		}
		env.GetIntArrayRegion(pixelArray, 0, int32(pixelCount), unsafe.Pointer(&intBuf[0]))
		return nil
	})

	// Blit to ANativeWindow.
	C.renderBitmapToWindow(
		nativeWindow,
		unsafe.Pointer(&intBuf[0]),
		C.int(windowWidth),
		C.int(windowHeight),
	)

	// Recycle bitmap.
	_ = bmp.Recycle()
}

func getUngrantedPermissions(
	env *jni.Env,
	activity *jni.Object,
) ([]string, error) {
	actCls := env.GetObjectClass(activity)

	getPMMid, err := env.GetMethodID(actCls, "getPackageManager",
		"()Landroid/content/pm/PackageManager;")
	if err != nil {
		return nil, nil
	}
	pm, err := env.CallObjectMethod(activity, getPMMid)
	if err != nil {
		return nil, nil
	}

	getPkgMid, err := env.GetMethodID(actCls, "getPackageName", "()Ljava/lang/String;")
	if err != nil {
		return nil, nil
	}
	pkgObj, err := env.CallObjectMethod(activity, getPkgMid)
	if err != nil {
		return nil, nil
	}
	pkgName := env.GoString((*jni.String)(unsafe.Pointer(pkgObj)))

	pmCls := env.GetObjectClass(pm)
	getAIMid, err := env.GetMethodID(pmCls, "getApplicationInfo",
		"(Ljava/lang/String;I)Landroid/content/pm/ApplicationInfo;")
	if err != nil {
		return nil, nil
	}
	jPkg, _ := env.NewStringUTF(pkgName)
	ai, err := env.CallObjectMethod(pm, getAIMid,
		jni.ObjectValue(&jPkg.Object), jni.IntValue(pmGetMetaData))
	if err != nil {
		return nil, nil
	}

	aiCls := env.GetObjectClass(ai)
	metaFid, err := env.GetFieldID(aiCls, "metaData", "Landroid/os/Bundle;")
	if err != nil {
		return nil, nil
	}
	metaObj := env.GetObjectField(ai, metaFid)
	if metaObj == nil || metaObj.Ref() == 0 {
		return nil, nil
	}

	bundleCls := env.GetObjectClass(metaObj)
	getStrMid, err := env.GetMethodID(bundleCls, "getString",
		"(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;")
	if err != nil {
		return nil, nil
	}
	jKey, _ := env.NewStringUTF("example.permissions")
	jEmpty, _ := env.NewStringUTF("")
	csvObj, err := env.CallObjectMethod(metaObj, getStrMid,
		jni.ObjectValue(&jKey.Object), jni.ObjectValue(&jEmpty.Object))
	if err != nil {
		return nil, nil
	}
	csv := env.GoString((*jni.String)(unsafe.Pointer(csvObj)))
	if csv == "" {
		return nil, nil
	}

	perms := strings.Split(csv, ",")
	var needed []string

	checkMid, err := env.GetMethodID(actCls, "checkSelfPermission", "(Ljava/lang/String;)I")
	if err != nil {
		return nil, nil
	}
	for _, perm := range perms {
		jPerm, _ := env.NewStringUTF(perm)
		result, err := env.CallIntMethod(activity, checkMid, jni.ObjectValue(&jPerm.Object))
		if err != nil {
			continue
		}
		if result != 0 {
			needed = append(needed, perm)
		}
	}

	return needed, nil
}

func requestPermissions(
	env *jni.Env,
	activity *jni.Object,
	perms []string,
) {
	actCls := env.GetObjectClass(activity)
	reqMid, err := env.GetMethodID(actCls, "requestPermissions", "([Ljava/lang/String;I)V")
	if err != nil {
		return
	}
	strCls, err := env.FindClass("java/lang/String")
	if err != nil {
		return
	}
	arr, err := env.NewObjectArray(int32(len(perms)), strCls, nil)
	if err != nil || arr == nil {
		return
	}
	for i, p := range perms {
		jP, _ := env.NewStringUTF(p)
		_ = env.SetObjectArrayElement(arr, int32(i), &jP.Object)
	}
	_ = env.CallVoidMethod(activity, reqMid,
		jni.ObjectValue(&arr.Object), jni.IntValue(permissionRequestCode))
}
