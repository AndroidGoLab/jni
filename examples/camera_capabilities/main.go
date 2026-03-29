//go:build android

// Command camera_capabilities enumerates all cameras via CameraManager,
// reports the camera ID list, and shows what the typed camera API provides.
// CameraCharacteristics details require raw JNI (the wrapper is not exported),
// so this example demonstrates the CameraManager typed API surface.
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

// extractCameraIDs converts the Java String[] from GetCameraIdList into Go strings.
func extractCameraIDs(vm *jni.VM, idArray *jni.Object) ([]string, error) {
	var ids []string
	err := vm.Do(func(env *jni.Env) error {
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
			ids = append(ids, env.GoString((*jni.String)(unsafe.Pointer(elem))))
		}
		return nil
	})
	return ids, err
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

	fmt.Fprintln(output, "=== Camera Capabilities ===")

	idArray, err := mgr.GetCameraIdList()
	if err != nil {
		return fmt.Errorf("GetCameraIdList: %w", err)
	}
	defer func() {
		if idArray != nil {
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(idArray)
				return nil
			})
		}
	}()

	cameraIDs, err := extractCameraIDs(vm, idArray)
	if err != nil {
		return fmt.Errorf("read camera IDs: %w", err)
	}

	fmt.Fprintf(output, "Found %d camera(s)\n\n", len(cameraIDs))

	for _, id := range cameraIDs {
		fmt.Fprintf(output, "Camera [%s]:\n", id)

		// GetCameraCharacteristics returns an opaque object; we show it was obtained.
		chars, err := mgr.GetCameraCharacteristics(id)
		if err != nil {
			fmt.Fprintf(output, "  Characteristics: error (%v)\n", err)
		} else if chars != nil && chars.Ref() != 0 {
			fmt.Fprintf(output, "  Characteristics: obtained\n")
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(chars)
				return nil
			})
		}

		// Check torch strength (API 33+).
		torchLevel, err := mgr.GetTorchStrengthLevel(id)
		if err != nil {
			fmt.Fprintf(output, "  Torch strength: not available (%v)\n", err)
		} else {
			fmt.Fprintf(output, "  Torch strength level: %d\n", torchLevel)
		}

		// Check device setup support (API 35+).
		setupSupported, err := mgr.IsCameraDeviceSetupSupported(id)
		if err != nil {
			fmt.Fprintf(output, "  Device setup: not available\n")
		} else {
			fmt.Fprintf(output, "  Device setup supported: %v\n", setupSupported)
		}

		fmt.Fprintln(output)
	}

	// Check concurrent camera support (API 30+).
	concurrentObj, err := mgr.GetConcurrentCameraIds()
	if err != nil {
		fmt.Fprintf(output, "Concurrent camera IDs: not available (%v)\n", err)
	} else if concurrentObj != nil && concurrentObj.Ref() != 0 {
		fmt.Fprintln(output, "Concurrent camera IDs: obtained (Set<Set<String>>)")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(concurrentObj)
			return nil
		})
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Camera capabilities enumeration complete.")
	return nil
}
