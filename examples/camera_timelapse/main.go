//go:build android

// Command camera_timelapse demonstrates camera setup for timelapse photography.
// It enumerates cameras, obtains characteristics, and shows timelapse
// configuration concepts using typed wrappers only.
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
	"time"
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

	fmt.Fprintln(output, "=== Camera Timelapse Setup ===")
	fmt.Fprintln(output, "")

	// List cameras.
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

	fmt.Fprintf(output, "Available cameras: %d\n", len(cameraIDs))
	if len(cameraIDs) == 0 {
		fmt.Fprintln(output, "No cameras available for timelapse")
		return nil
	}

	// Select first camera for timelapse demo.
	selectedID := cameraIDs[0]
	fmt.Fprintf(output, "Selected camera: %s\n", selectedID)

	// Read characteristics (opaque object; no exported wrapper for details).
	chars, err := mgr.GetCameraCharacteristics(selectedID)
	if err != nil {
		fmt.Fprintf(output, "GetCameraCharacteristics: %v\n", err)
	} else if chars != nil {
		fmt.Fprintln(output, "Characteristics: obtained")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(chars)
			return nil
		})
	}

	// Demonstrate timelapse configuration.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "--- Timelapse Configuration ---")

	type timelapseConfig struct {
		name     string
		interval time.Duration
		count    int
		total    time.Duration
	}

	configs := []timelapseConfig{
		{"Fast (construction)", 5 * time.Second, 720, 1 * time.Hour},
		{"Medium (clouds)", 10 * time.Second, 360, 1 * time.Hour},
		{"Slow (plant growth)", 5 * time.Minute, 288, 24 * time.Hour},
		{"Ultra-slow (seasons)", 1 * time.Hour, 168, 7 * 24 * time.Hour},
	}

	for _, cfg := range configs {
		fmt.Fprintf(output, "  %s:\n", cfg.name)
		fmt.Fprintf(output, "    Interval: %v\n", cfg.interval)
		fmt.Fprintf(output, "    Frames over %v: %d\n", cfg.total, cfg.count)
		fps := 30
		videoLen := time.Duration(cfg.count/fps) * time.Second
		fmt.Fprintf(output, "    Result at %dfps: %v video\n", fps, videoLen)
	}

	// Show the Camera2 timelapse workflow.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "--- Timelapse Capture Workflow ---")
	fmt.Fprintln(output, "  1. CameraManager.openCamera() with StateCallback")
	fmt.Fprintln(output, "  2. Create CaptureRequest for TEMPLATE_STILL_CAPTURE")
	fmt.Fprintln(output, "  3. Set up ImageReader for JPEG output")
	fmt.Fprintln(output, "  4. Create CameraCaptureSession")
	fmt.Fprintln(output, "  5. Use Go ticker for capture intervals:")
	fmt.Fprintln(output, "     ticker := time.NewTicker(interval)")
	fmt.Fprintln(output, "     for range ticker.C {")
	fmt.Fprintln(output, "         session.capture(request, callback, handler)")
	fmt.Fprintln(output, "     }")
	fmt.Fprintln(output, "  6. Save each frame to storage")
	fmt.Fprintln(output, "  7. Combine frames with ffmpeg or similar")

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Timelapse setup complete.")
	return nil
}
