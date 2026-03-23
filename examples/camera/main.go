//go:build android

// Command camera demonstrates using the Android Camera2 CameraManager API.
// It lists camera IDs and queries their characteristics.
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
	"github.com/AndroidGoLab/jni/hardware/camera"
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

// cameraFacingName maps CameraCharacteristics.LENS_FACING values.
func cameraFacingName(facing int32) string {
	switch facing {
	case 0:
		return "FRONT"
	case 1:
		return "BACK"
	case 2:
		return "EXTERNAL"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", facing)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := camera.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("camera.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Camera Devices ===")

	// GetCameraIdList returns a String[] array.
	idArray, err := mgr.GetCameraIdList()
	if err != nil {
		return fmt.Errorf("GetCameraIdList: %w", err)
	}

	var cameraIDs []string
	err = vm.Do(func(env *jni.Env) error {
		if idArray == nil {
			return nil
		}
		arr := (*jni.Array)(unsafe.Pointer(idArray))
		count := env.GetArrayLength(arr)
		objArr := (*jni.ObjectArray)(unsafe.Pointer(idArray))

		for i := int32(0); i < count; i++ {
			elem, err := env.GetObjectArrayElement(objArr, i)
			if err != nil {
				continue
			}
			cameraIDs = append(cameraIDs, env.GoString((*jni.String)(unsafe.Pointer(elem))))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("read camera IDs: %w", err)
	}

	fmt.Fprintf(output, "Camera count: %d\n", len(cameraIDs))

	// For each camera, get characteristics and read LENS_FACING.
	for _, id := range cameraIDs {
		chars, err := mgr.GetCameraCharacteristics(id)
		if err != nil {
			fmt.Fprintf(output, "  [%s] chars: %v\n", id, err)
			continue
		}

		// Read LENS_FACING via CameraCharacteristics.get(Key).
		var facing int32
		var facingFound bool
		_ = vm.Do(func(env *jni.Env) error {
			charsCls := env.GetObjectClass(chars)
			getMid, err := env.GetMethodID(charsCls, "get", "(Landroid/hardware/camera2/CameraCharacteristics$Key;)Ljava/lang/Object;")
			if err != nil {
				return err
			}

			// LENS_FACING is a static field on CameraCharacteristics.
			lensFacingFid, err := env.GetStaticFieldID(charsCls, "LENS_FACING", "Landroid/hardware/camera2/CameraCharacteristics$Key;")
			if err != nil {
				return err
			}
			lensFacingKey := env.GetStaticObjectField(charsCls, lensFacingFid)
			if lensFacingKey == nil {
				return nil
			}

			facingObj, err := env.CallObjectMethod(chars, getMid, jni.ObjectValue(lensFacingKey))
			if err != nil || facingObj == nil {
				return err
			}

			// Unbox Integer to int.
			intCls := env.GetObjectClass(facingObj)
			intValueMid, err := env.GetMethodID(intCls, "intValue", "()I")
			if err != nil {
				return err
			}
			facing, err = env.CallIntMethod(facingObj, intValueMid)
			if err != nil {
				return err
			}
			facingFound = true
			return nil
		})

		if facingFound {
			fmt.Fprintf(output, "  [%s] facing: %s\n", id, cameraFacingName(facing))
		} else {
			fmt.Fprintf(output, "  [%s] (no facing info)\n", id)
		}
	}

	return nil
}
