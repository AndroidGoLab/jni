//go:build android

// Command storage demonstrates the Android StorageManager API.
// It queries the primary storage volume and displays its properties.
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/os/storage"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
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

	mgr, err := storage.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Storage Info ===")
	fmt.Fprintln(output)

	// Query the primary storage volume.
	volObj, err := mgr.GetPrimaryStorageVolume()
	if err != nil {
		return fmt.Errorf("getPrimaryStorageVolume: %w", err)
	}

	vol := storage.Volume{VM: vm, Obj: volObj}

	desc, err := vol.GetDescription(ctx.Obj)
	if err != nil {
		return fmt.Errorf("getDescription: %w", err)
	}
	fmt.Fprintf(output, "desc: %s\n", desc)

	state, err := vol.GetState()
	if err != nil {
		return fmt.Errorf("getState: %w", err)
	}
	fmt.Fprintf(output, "state: %s\n", state)

	isPrimary, err := vol.IsPrimary()
	if err != nil {
		return fmt.Errorf("isPrimary: %w", err)
	}
	fmt.Fprintf(output, "primary: %v\n", isPrimary)

	isEmulated, err := vol.IsEmulated()
	if err != nil {
		return fmt.Errorf("isEmulated: %w", err)
	}
	fmt.Fprintf(output, "emulated: %v\n", isEmulated)

	isRemovable, err := vol.IsRemovable()
	if err != nil {
		return fmt.Errorf("isRemovable: %w", err)
	}
	fmt.Fprintf(output, "removable: %v\n", isRemovable)

	uuid, err := vol.GetUuid()
	if err != nil {
		return fmt.Errorf("getUuid: %w", err)
	}
	if uuid == "" {
		uuid = "(internal)"
	}
	fmt.Fprintf(output, "uuid: %s\n", uuid)

	mediaName, err := vol.GetMediaStoreVolumeName()
	if err != nil {
		return fmt.Errorf("getMediaStoreVolumeName: %w", err)
	}
	fmt.Fprintf(output, "media: %s\n", mediaName)

	// Check filesystem encryption support.
	checkpointOK, err := mgr.IsCheckpointSupported()
	if err != nil {
		fmt.Fprintf(output, "checkpoint: err\n")
	} else {
		fmt.Fprintf(output, "checkpoint: %v\n", checkpointOK)
	}

	return nil
}
