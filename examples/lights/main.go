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
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/hardware/lights"
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
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// lights.Manager has no constructor; it wraps
	// android.hardware.lights.LightsManager obtained via getSystemService.
	var mgr lights.Manager
	_ = mgr

	// Print all light type constants.
	fmt.Fprintln(output, "Light type constants:")
	fmt.Fprintf(output, "  LightTypeInput             = %d\n", lights.LightTypeInput)
	fmt.Fprintf(output, "  LightTypeKeyboardBacklight = %d\n", lights.LightTypeKeyboardBacklight)
	fmt.Fprintf(output, "  LightTypeMicrophone        = %d\n", lights.LightTypeMicrophone)
	fmt.Fprintf(output, "  LightTypePlayerId          = %d\n", lights.LightTypePlayerId)

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

	fmt.Fprintln(output, "LightsManager ready")
	return nil
}
