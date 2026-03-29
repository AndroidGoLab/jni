//go:build android

// Command lights_led demonstrates the LightsManager API for LED control.
// It obtains the LightsManager system service and describes the typed
// wrapper API surface for LED control.
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

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Lights LED Control ===")
	fmt.Fprintln(output, "")

	// LightsManager requires API 31+
	svcObj, err := ctx.GetSystemService("lights")
	if err != nil {
		fmt.Fprintf(output, "LightsManager not available: %v\n", err)
		fmt.Fprintln(output, "(Requires API 31+)")
		fmt.Fprintln(output, "")
		printAPIOverview(output)
		return nil
	}

	if svcObj == nil || svcObj.Ref() == 0 {
		fmt.Fprintln(output, "LightsManager not available (null)")
		fmt.Fprintln(output, "(Requires API 31+)")
		fmt.Fprintln(output, "")
		printAPIOverview(output)
		return nil
	}

	mgr := lights.Manager{VM: vm, Obj: svcObj}

	fmt.Fprintln(output, "LightsManager obtained OK")
	fmt.Fprintln(output, "")

	// GetLights returns a List<Light> object. Since there is no typed
	// java.util.List wrapper, we check if the call succeeds rather than
	// iterating elements via raw JNI.
	lightsListObj, err := mgr.GetLights()
	if err != nil {
		fmt.Fprintf(output, "GetLights: error: %v\n", err)
	} else if lightsListObj == nil {
		fmt.Fprintln(output, "GetLights: (null)")
	} else {
		fmt.Fprintln(output, "GetLights: returned list object OK")
	}

	// Show LED control API overview
	fmt.Fprintln(output, "")
	printAPIOverview(output)

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Lights LED example complete.")
	return nil
}

func printAPIOverview(output *bytes.Buffer) {
	fmt.Fprintln(output, "--- LED Control API Overview ---")
	fmt.Fprintln(output, "  LightsManager (android.hardware.lights.LightsManager):")
	fmt.Fprintln(output, "    GetLights() -> List<Light>")
	fmt.Fprintln(output, "    GetLightState(light) -> LightState")
	fmt.Fprintln(output, "    OpenSession() -> LightsSession")
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "  Light properties:")
	fmt.Fprintln(output, "    GetId, GetName, GetType, GetOrdinal")
	fmt.Fprintln(output, "    HasBrightnessControl, HasRgbControl")
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "  Light types:")
	fmt.Fprintf(output, "    Input=%d, KeyboardBacklight=%d\n",
		lights.LightTypeInput, lights.LightTypeKeyboardBacklight)
	fmt.Fprintf(output, "    Microphone=%d, PlayerId=%d\n",
		lights.LightTypeMicrophone, lights.LightTypePlayerId)
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "  LED control workflow:")
	fmt.Fprintln(output, "    1. session := mgr.OpenSession()")
	fmt.Fprintln(output, "    2. builder := LightStateBuilder{}")
	fmt.Fprintln(output, "    3. builder.SetColor(0xFFFF0000) // red")
	fmt.Fprintln(output, "    4. state := builder.Build()")
	fmt.Fprintln(output, "    5. reqBuilder.AddLight(light, state)")
	fmt.Fprintln(output, "    6. session.RequestLights(reqBuilder.Build())")
	fmt.Fprintln(output, "    7. session.Close()")
}
