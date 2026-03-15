//go:build android

// Command camera demonstrates using the Android Camera2 CameraManager API.
// It is built as a c-shared library and packaged into an APK using the
// shared apk.mk infrastructure.
//
// This example obtains the CameraManager system service and shows how the
// exported NewManager constructor and Close cleanup method are used.
// The camera package also provides unexported methods for torch control
// (setTorchMode, registerTorchCallback, unregisterTorchCallback) and a
// torchCallback type with an OnTorchModeChanged callback, which are used
// internally by higher-level wrappers.
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
	"github.com/AndroidGoLab/jni/hardware/camera"
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

	// NewManager obtains the CameraManager system service.
	mgr, err := camera.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("camera.NewManager: %v", err)
	}
	// Close releases the JNI global reference; always defer it.
	defer mgr.Close()

	fmt.Fprintln(&output, "CameraManager obtained successfully")

	// The following methods are unexported (package-internal) and are
	// used by higher-level Go wrappers that build on this package:
	//
	//   mgr.setTorchMode(cameraId string, enabled bool) error
	//     Turns the torch (flashlight) for the given camera on or off.
	//
	//   mgr.registerTorchCallback(callback *jni.Object, handler *jni.Object) error
	//     Registers a TorchCallback to receive torch mode change events.
	//
	//   mgr.unregisterTorchCallback(callback *jni.Object) error
	//     Unregisters a previously registered TorchCallback.
	//
	// The torchCallback struct provides a Go-friendly callback interface:
	//
	//   torchCallback{
	//       OnTorchModeChanged: func(cameraId string, enabled bool) { ... },
	//   }
	//
	// It is registered via the unexported registertorchCallback function,
	// which creates a Java proxy implementing CameraManager.TorchCallback.

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
