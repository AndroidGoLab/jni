//go:build android

// Command lights demonstrates the LightsManager JNI bindings. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the LightsManager system service and prints
// all light type constants. The package wraps
// android.hardware.lights.LightsManager and provides the Light data
// class with ID, Name, and Type fields.
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
	"github.com/AndroidGoLab/jni/hardware/lights"
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

	mgr, err := lights.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("lights.NewManager: %v", err)
	}
	_ = mgr

	// Print all light type constants.
	fmt.Fprintln(&output, "Light type constants:")
	fmt.Fprintf(&output, "  LightTypeInput             = %d\n", lights.LightTypeInput)
	fmt.Fprintf(&output, "  LightTypeKeyboardBacklight = %d\n", lights.LightTypeKeyboardBacklight)
	fmt.Fprintf(&output, "  LightTypeMicrophone        = %d\n", lights.LightTypeMicrophone)
	fmt.Fprintf(&output, "  LightTypePlayerId          = %d\n", lights.LightTypePlayerId)

	// The Light data class (exported) has these fields:
	//   ID   int    - unique light identifier
	//   Name string - human-readable light name
	//   Type int    - one of the LightType* constants above

	// Package-internal Manager methods:
	//   getLightsRaw()   - returns Java List of Light objects
	//   openSessionRaw() - opens a LightsSession for controlling lights
	//
	// The Session type provides:
	//   requestLightsRaw(request) - apply a lights request
	//   closeRaw()                - close the Java session
	//   Close()                   - release the Go global reference
	//
	// lightStateBuilder builds a LightState:
	//   setColor(argb int32)     - set the ARGB color
	//   setPlayerId(id int32)    - set player ID for player lights
	//   build()                  - produce the LightState
	//
	// lightsRequestBuilder builds a LightsRequest:
	//   addLight(light, state)   - add a light with desired state
	//   clearLight(light)        - clear a previously set light
	//   build()                  - produce the LightsRequest

	fmt.Fprintln(&output, "LightsManager ready")
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
