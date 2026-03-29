//go:build android

// Command camera_availability uses CameraManager to check camera
// availability. It enumerates all cameras, obtains their characteristics,
// checks for concurrent camera support, and describes the AvailabilityCallback
// registration API using typed wrappers only.
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

	fmt.Fprintln(output, "=== Camera Availability ===")
	fmt.Fprintln(output, "")

	// Enumerate all cameras.
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

	fmt.Fprintf(output, "Registered cameras: %d\n", len(cameraIDs))
	fmt.Fprintln(output, "")

	// Report each camera with characteristics check.
	fmt.Fprintln(output, "Camera list:")
	for _, id := range cameraIDs {
		chars, err := mgr.GetCameraCharacteristics(id)
		if err != nil {
			fmt.Fprintf(output, "  [%s] error: %v\n", id, err)
			continue
		}

		fmt.Fprintf(output, "  [%s] characteristics=obtained status=listed (available at enumeration time)\n", id)

		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(chars)
			return nil
		})
	}

	// Check for concurrent camera support (API 30+).
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Concurrent Camera IDs:")
	concurrentObj, err := mgr.GetConcurrentCameraIds()
	if err != nil {
		fmt.Fprintf(output, "  Not available: %v\n", err)
	} else if concurrentObj == nil {
		fmt.Fprintln(output, "  (null result)")
	} else {
		fmt.Fprintln(output, "  Concurrent camera combinations: obtained (Set<Set<String>>)")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(concurrentObj)
			return nil
		})
	}

	// Describe the availability callback API.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "--- AvailabilityCallback API ---")
	fmt.Fprintln(output, "The camera package provides:")
	fmt.Fprintln(output, "  mgr.RegisterAvailabilityCallback(callback, handler)")
	fmt.Fprintln(output, "  mgr.UnregisterAvailabilityCallback(callback)")
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "ManagerAvailabilityCallback methods:")
	fmt.Fprintln(output, "  OnCameraAvailable(cameraId)")
	fmt.Fprintln(output, "  OnCameraUnavailable(cameraId)")
	fmt.Fprintln(output, "  OnCameraAccessPrioritiesChanged()")
	fmt.Fprintln(output, "  OnPhysicalCameraAvailable(cameraId, physicalId)")
	fmt.Fprintln(output, "  OnPhysicalCameraUnavailable(cameraId, physicalId)")
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "These callbacks fire when cameras become available/unavailable,")
	fmt.Fprintln(output, "e.g., when another app opens or closes a camera.")

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Camera availability check complete.")
	return nil
}
