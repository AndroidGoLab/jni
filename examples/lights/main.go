//go:build android

// Command lights demonstrates the LightsManager JNI bindings.
// It obtains the LightsManager system service, queries available lights
// via GetLights, and prints all properties of each Light using GetId,
// GetName, GetType, GetOrdinal, HasBrightnessControl, HasRgbControl,
// and GetLightState.
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/hardware/lights"
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

// lightTypeName returns a human-readable name for a light type constant.
func lightTypeName(t int32) string {
	switch int(t) {
	case lights.LightTypeInput:
		return "Input"
	case lights.LightTypeKeyboardBacklight:
		return "KeyboardBacklight"
	case lights.LightTypeMicrophone:
		return "Microphone"
	case lights.LightTypePlayerId:
		return "PlayerId"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== LightsManager ===")

	// Obtain LightsManager via getSystemService.
	// LightsManager requires API 31+, so handle errors gracefully.
	svcObj, err := ctx.GetSystemService("lights")
	if err != nil {
		fmt.Fprintf(output, "\nNot available: %v\n", err)
		fmt.Fprintln(output, "(Requires API 31+)")
		return nil
	}

	if svcObj == nil || svcObj.Ref() == 0 {
		fmt.Fprintln(output, "\nNot available (null)")
		fmt.Fprintln(output, "(Requires API 31+)")
		return nil
	}

	mgr := lights.Manager{VM: vm, Obj: svcObj}
	defer vm.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(mgr.Obj)
		return nil
	})

	fmt.Fprintln(output, "Service obtained OK")

	// Filtered: GetLights returns generic type (List<Light>)
	// lightsListObj, err := mgr.GetLights()
	// ...
	fmt.Fprintln(output, "(GetLights filtered: returns generic type)")

	return nil
}
