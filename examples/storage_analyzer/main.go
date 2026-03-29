//go:build android

// Command storage_analyzer demonstrates the StorageManager API to query
// storage volumes and their properties using typed wrappers.
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
	"github.com/AndroidGoLab/jni/os/storage"
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

	mgr, err := storage.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Storage Analyzer ===")
	fmt.Fprintln(output)

	// --- Primary storage volume ---
	fmt.Fprintln(output, "Primary volume:")
	volObj, err := mgr.GetPrimaryStorageVolume()
	if err != nil {
		fmt.Fprintf(output, "  error: %v\n", err)
	} else {
		vol := storage.Volume{VM: vm, Obj: volObj}

		desc, err := vol.GetDescription(ctx.Obj)
		if err != nil {
			fmt.Fprintln(output, "  description: error")
		} else {
			fmt.Fprintf(output, "  description: %s\n", desc)
		}

		state, err := vol.GetState()
		if err != nil {
			fmt.Fprintln(output, "  state: error")
		} else {
			fmt.Fprintf(output, "  state: %s\n", state)
		}

		isPrimary, _ := vol.IsPrimary()
		fmt.Fprintf(output, "  primary: %v\n", isPrimary)

		isEmulated, _ := vol.IsEmulated()
		fmt.Fprintf(output, "  emulated: %v\n", isEmulated)

		isRemovable, _ := vol.IsRemovable()
		fmt.Fprintf(output, "  removable: %v\n", isRemovable)

		uuid, _ := vol.GetUuid()
		if uuid == "" {
			uuid = "(internal)"
		}
		fmt.Fprintf(output, "  uuid: %s\n", uuid)

		mediaName, _ := vol.GetMediaStoreVolumeName()
		fmt.Fprintf(output, "  mediaStore: %s\n", mediaName)

		volStr, _ := vol.ToString()
		fmt.Fprintf(output, "  toString: %s\n", volStr)
	}

	// --- Checkpoint support ---
	fmt.Fprintln(output)
	checkpointOK, err := mgr.IsCheckpointSupported()
	if err != nil {
		fmt.Fprintf(output, "checkpoint supported: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "checkpoint supported: %v\n", checkpointOK)
	}

	return nil
}
