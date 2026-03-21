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
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/hardware/lights"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
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

	fmt.Fprintln(output, "Service obtained OK")

	// GetLights - returns java.util.List<Light>.
	lightsListObj, err := mgr.GetLights()
	if err != nil {
		fmt.Fprintf(output, "GetLights: %v\n", err)
		return nil
	}

	// Iterate the List<Light> and call every getter.
	var lightCount int32
	err = vm.Do(func(env *jni.Env) error {
		if lightsListObj == nil {
			fmt.Fprintln(output, "GetLights returned nil")
			return nil
		}

		listCls, err := env.FindClass("java/util/List")
		if err != nil {
			return err
		}
		sizeMid, err := env.GetMethodID(listCls, "size", "()I")
		if err != nil {
			return err
		}
		getMid, err := env.GetMethodID(listCls, "get", "(I)Ljava/lang/Object;")
		if err != nil {
			return err
		}

		lightCount, err = env.CallIntMethod(lightsListObj, sizeMid)
		if err != nil {
			return err
		}

		fmt.Fprintf(output, "Lights found: %d\n", lightCount)

		for i := int32(0); i < lightCount; i++ {
			elemObj, err := env.CallObjectMethod(lightsListObj, getMid, jni.IntValue(i))
			if err != nil || elemObj == nil {
				continue
			}

			// Wrap as a lights.Light to use typed accessors.
			light := lights.Light{
				VM:  vm,
				Obj: env.NewGlobalRef(elemObj),
			}

			fmt.Fprintf(output, "\n  Light #%d:\n", i)

			// GetId
			id, err := light.GetId()
			if err != nil {
				fmt.Fprintf(output, "    ID: err: %v\n", err)
			} else {
				fmt.Fprintf(output, "    ID: %d\n", id)
			}

			// GetName
			name, err := light.GetName()
			if err != nil {
				fmt.Fprintf(output, "    Name: err: %v\n", err)
			} else {
				fmt.Fprintf(output, "    Name: %s\n", name)
			}

			// GetType
			typ, err := light.GetType()
			if err != nil {
				fmt.Fprintf(output, "    Type: err: %v\n", err)
			} else {
				fmt.Fprintf(output, "    Type: %s (%d)\n", lightTypeName(typ), typ)
			}

			// GetOrdinal
			ordinal, err := light.GetOrdinal()
			if err != nil {
				fmt.Fprintf(output, "    Ordinal: err: %v\n", err)
			} else {
				fmt.Fprintf(output, "    Ordinal: %d\n", ordinal)
			}

			// HasBrightnessControl
			hasBrightness, err := light.HasBrightnessControl()
			if err != nil {
				fmt.Fprintf(output, "    Brightness: err: %v\n", err)
			} else {
				fmt.Fprintf(output, "    Brightness: %v\n", hasBrightness)
			}

			// HasRgbControl
			hasRgb, err := light.HasRgbControl()
			if err != nil {
				fmt.Fprintf(output, "    RGB: err: %v\n", err)
			} else {
				fmt.Fprintf(output, "    RGB: %v\n", hasRgb)
			}

			// GetLightState - returns LightState for this light.
			stateObj, err := mgr.GetLightState(light.Obj)
			if err != nil {
				fmt.Fprintf(output, "    State: err: %v\n", err)
			} else if stateObj == nil {
				fmt.Fprintf(output, "    State: nil\n")
			} else {
				ls := lights.LightState{VM: vm, Obj: stateObj}

				color, err := ls.GetColor()
				if err != nil {
					fmt.Fprintf(output, "    Color: err: %v\n", err)
				} else {
					fmt.Fprintf(output, "    Color: 0x%08X\n", uint32(color))
				}

				playerId, err := ls.GetPlayerId()
				if err != nil {
					fmt.Fprintf(output, "    PlayerID: err: %v\n", err)
				} else {
					fmt.Fprintf(output, "    PlayerID: %d\n", playerId)
				}

				env.DeleteGlobalRef(ls.Obj)
			}

			env.DeleteGlobalRef(light.Obj)
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(output, "Light iteration: %v\n", err)
	}

	if lightCount == 0 {
		fmt.Fprintln(output, "(No lights found)")
	}

	return nil
}
