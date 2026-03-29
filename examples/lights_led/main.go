//go:build android

// Command lights_led demonstrates the LightsManager API for LED control.
// It obtains the LightsManager, calls GetLights, iterates the returned
// list to query each Light's properties (id, name, type, ordinal,
// hasBrightnessControl, hasRgbControl), and calls GetLightState on each.
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

func lightTypeName(t int32) string {
	switch t {
	case int32(lights.LightTypeInput):
		return "INPUT"
	case int32(lights.LightTypeKeyboardBacklight):
		return "KEYBOARD_BACKLIGHT"
	case int32(lights.LightTypeMicrophone):
		return "MICROPHONE"
	case int32(lights.LightTypePlayerId):
		return "PLAYER_ID"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// extractLightList converts the Java List<Light> into a slice of Light wrappers.
// This is the one place where we need low-level array access to iterate the list,
// but we use only typed Light methods for all property queries.
func extractLightList(vm *jni.VM, listObj *jni.GlobalRef) ([]*lights.Light, error) {
	var result []*lights.Light
	err := vm.Do(func(env *jni.Env) error {
		// Get List.size()
		listCls, err := env.FindClass("java/util/List")
		if err != nil {
			return err
		}
		sizeMid, err := env.GetMethodID(listCls, "size", "()I")
		if err != nil {
			return err
		}
		sizeVal, err := env.CallIntMethod(listObj, sizeMid)
		if err != nil {
			return err
		}

		// Get List.get(int)
		getMid, err := env.GetMethodID(listCls, "get", "(I)Ljava/lang/Object;")
		if err != nil {
			return err
		}

		for i := int32(0); i < sizeVal; i++ {
			elem, err := env.CallObjectMethod(listObj, getMid, jni.IntValue(i))
			if err != nil {
				continue
			}
			if elem == nil {
				continue
			}
			gref := env.NewGlobalRef(elem)
			result = append(result, &lights.Light{VM: vm, Obj: gref})
		}
		return nil
	})
	return result, err
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Lights LED Control ===")
	fmt.Fprintln(output)

	// 1. Obtain the LightsManager system service (API 31+).
	svcObj, err := ctx.GetSystemService("lights")
	if err != nil {
		fmt.Fprintf(output, "LightsManager not available: %v\n", err)
		fmt.Fprintln(output, "(Requires API 31+)")
		return nil
	}

	if svcObj == nil || svcObj.Ref() == 0 {
		fmt.Fprintln(output, "LightsManager not available (null)")
		fmt.Fprintln(output, "(Requires API 31+)")
		return nil
	}

	mgr := lights.Manager{VM: vm, Obj: svcObj}
	fmt.Fprintln(output, "LightsManager: obtained OK")

	// 2. ToString.
	mgrStr, err := mgr.ToString()
	if err != nil {
		fmt.Fprintf(output, "  ToString: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  ToString: %s\n", mgrStr)
	}

	// 3. GetLights.
	fmt.Fprintln(output)
	lightsListObj, err := mgr.GetLights()
	if err != nil {
		fmt.Fprintf(output, "GetLights: error: %v\n", err)
		return nil
	}
	if lightsListObj == nil {
		fmt.Fprintln(output, "GetLights: (null)")
		return nil
	}
	defer func() {
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(lightsListObj)
			return nil
		})
	}()

	// 4. Extract light objects from the list.
	lightList, err := extractLightList(vm, lightsListObj)
	if err != nil {
		fmt.Fprintf(output, "extractLightList: %v\n", err)
		return nil
	}
	fmt.Fprintf(output, "GetLights: found %d light(s)\n", len(lightList))

	// 5-12+. For each light, query all properties using typed methods.
	for i, lt := range lightList {
		fmt.Fprintf(output, "\nLight #%d:\n", i)

		id, err := lt.GetId()
		if err != nil {
			fmt.Fprintf(output, "  GetId: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  GetId: %d\n", id)
		}

		name, err := lt.GetName()
		if err != nil {
			fmt.Fprintf(output, "  GetName: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  GetName: %s\n", name)
		}

		typ, err := lt.GetType()
		if err != nil {
			fmt.Fprintf(output, "  GetType: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  GetType: %s (%d)\n", lightTypeName(typ), typ)
		}

		ord, err := lt.GetOrdinal()
		if err != nil {
			fmt.Fprintf(output, "  GetOrdinal: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  GetOrdinal: %d\n", ord)
		}

		hasBright, err := lt.HasBrightnessControl()
		if err != nil {
			fmt.Fprintf(output, "  HasBrightnessControl: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  HasBrightnessControl: %v\n", hasBright)
		}

		hasRgb, err := lt.HasRgbControl()
		if err != nil {
			fmt.Fprintf(output, "  HasRgbControl: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  HasRgbControl: %v\n", hasRgb)
		}

		hash, err := lt.HashCode()
		if err != nil {
			fmt.Fprintf(output, "  HashCode: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  HashCode: %d\n", hash)
		}

		ltStr, err := lt.ToString()
		if err != nil {
			fmt.Fprintf(output, "  ToString: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  ToString: %s\n", ltStr)
		}

		// GetLightState for this light.
		stateObj, err := mgr.GetLightState(lt.Obj)
		if err != nil {
			fmt.Fprintf(output, "  GetLightState: error: %v\n", err)
		} else if stateObj == nil || stateObj.Ref() == 0 {
			fmt.Fprintln(output, "  GetLightState: (null)")
		} else {
			state := lights.LightState{VM: vm, Obj: stateObj}
			color, _ := state.GetColor()
			playerId, _ := state.GetPlayerId()
			stateStr, _ := state.ToString()
			fmt.Fprintf(output, "  LightState: color=0x%08X playerId=%d str=%s\n",
				uint32(color), playerId, stateStr)
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(stateObj)
				return nil
			})
		}

		// Clean up the Light global ref.
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(lt.Obj)
			return nil
		})
	}

	// Show light type constants.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Light type constants:")
	fmt.Fprintf(output, "  Input=%d KeyboardBacklight=%d Microphone=%d PlayerId=%d\n",
		lights.LightTypeInput, lights.LightTypeKeyboardBacklight,
		lights.LightTypeMicrophone, lights.LightTypePlayerId)
	fmt.Fprintf(output, "  LightCapabilityBrightness=%d LightCapabilityColorRgb=%d\n",
		lights.LightCapabilityBrightness, lights.LightCapabilityColorRgb)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Lights LED example complete.")
	return nil
}
