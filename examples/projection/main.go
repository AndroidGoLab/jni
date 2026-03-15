//go:build android

// Command projection demonstrates using the MediaProjection API. It is built
// as a c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the MediaProjectionManager system service and
// describes the screen capture workflow: creating a capture intent,
// obtaining a MediaProjection from the activity result, registering a
// projectionCallback, and creating a virtual display.
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
	"github.com/AndroidGoLab/jni/media/projection"
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

	// --- NewManager ---
	mgr, err := projection.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("projection.NewManager: %v", err)
	}
	defer mgr.Close()

	// --- Manager methods (unexported) ---
	// The screen capture workflow:
	//
	// 1. Create a screen capture intent:
	//    intent, err := mgr.createScreenCaptureIntent()
	//    Pass this intent to startActivityForResult.
	//
	// 2. Obtain the projection from the result:
	//    projObj, err := mgr.getMediaProjection(resultCode, resultData)

	// --- Projection (unexported methods) ---
	// The Projection type wraps android.media.projection.MediaProjection:
	//
	//   proj.stop()
	//     Stop the media projection.
	//
	//   proj.registerCallback(callback, handler *jni.Object) error
	//     Register a callback for projection lifecycle events.
	//
	//   proj.createVirtualDisplayRaw(name, width, height, dpi, flags,
	//     surface, callback, handler) (*jni.Object, error)
	//     Create a virtual display for screen capture.

	// --- projectionCallback (unexported) ---
	// Registered via registerprojectionCallback to handle stop events:
	//
	//   projectionCallback{
	//     OnStop func()
	//   }
	//   proxy, cleanup, err := registerprojectionCallback(env, cb)

	// --- VirtualDisplay (unexported methods) ---
	// The VirtualDisplay type wraps android.hardware.display.VirtualDisplay:
	//
	//   vd.release()
	//     Release the virtual display when screen capture is done.

	fmt.Fprintf(&output, "MediaProjectionManager obtained from context\n")
	fmt.Fprintln(&output, "Projection example complete.")
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
