//go:build android

// Command camera_qr_scanner demonstrates the CameraManager API for setting up
// a QR code scanner. It enumerates cameras, obtains characteristics, and
// describes the Camera2 workflow for QR scanning.
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

	fmt.Fprintln(output, "=== Camera QR Scanner Setup ===")
	fmt.Fprintln(output, "")

	// Step 1: Enumerate cameras.
	fmt.Fprintln(output, "Step 1: Enumerating cameras...")
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

	fmt.Fprintf(output, "  Found %d camera(s)\n", len(cameraIDs))
	for _, id := range cameraIDs {
		fmt.Fprintf(output, "    Camera ID: %s\n", id)
	}

	if len(cameraIDs) == 0 {
		fmt.Fprintln(output, "  No cameras available for QR scanning")
		return nil
	}

	// Step 2: Get characteristics for each camera.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Step 2: Reading camera characteristics...")
	for _, id := range cameraIDs {
		chars, err := mgr.GetCameraCharacteristics(id)
		if err != nil {
			fmt.Fprintf(output, "  Camera [%s]: error getting characteristics: %v\n", id, err)
			continue
		}
		fmt.Fprintf(output, "  Camera [%s]: characteristics obtained\n", id)
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(chars)
			return nil
		})
	}

	// Use first camera for QR scanning workflow description.
	selectedID := cameraIDs[0]
	fmt.Fprintf(output, "\n  Selected camera for QR scanning: %s\n", selectedID)

	// Step 3: QR scanning setup overview.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Step 3: QR scanning setup workflow")
	fmt.Fprintln(output, "  For QR scanning, the typical Camera2 workflow is:")
	fmt.Fprintln(output, "  1. CameraManager.openCamera(cameraId, callback, handler)")
	fmt.Fprintln(output, "  2. Create CaptureRequest.Builder for TEMPLATE_PREVIEW")
	fmt.Fprintln(output, "  3. Set up ImageReader with YUV_420_888 format")
	fmt.Fprintln(output, "  4. Add ImageReader surface as capture target")
	fmt.Fprintln(output, "  5. Create CameraCaptureSession")
	fmt.Fprintln(output, "  6. Start repeating request for preview frames")
	fmt.Fprintln(output, "  7. Process frames in ImageReader.OnImageAvailableListener")
	fmt.Fprintln(output, "  8. Feed frames to a QR decode library")
	fmt.Fprintln(output, "")
	fmt.Fprintf(output, "  Camera %s is ready for openCamera() call\n", selectedID)
	fmt.Fprintln(output, "  (openCamera requires a CameraDevice.StateCallback proxy)")

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "QR scanner camera setup complete.")
	return nil
}
