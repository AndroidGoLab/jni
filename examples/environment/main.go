//go:build android

// Command environment demonstrates the Android Environment API.
// It queries directory paths, storage state, and storage properties
// using the generated Environment bindings.
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
	"github.com/AndroidGoLab/jni/os/environment"
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

// fileToPath extracts the absolute path string from a java.io.File object.
func fileToPath(vm *jni.VM, fileObj *jni.Object) (string, error) {
	var path string
	err := vm.Do(func(env *jni.Env) error {
		if fileObj == nil {
			return fmt.Errorf("nil File object")
		}
		fileCls, err := env.FindClass("java/io/File")
		if err != nil {
			return fmt.Errorf("find File class: %w", err)
		}
		mid, err := env.GetMethodID(fileCls, "getAbsolutePath", "()Ljava/lang/String;")
		if err != nil {
			return fmt.Errorf("get getAbsolutePath: %w", err)
		}
		pathObj, err := env.CallObjectMethod(fileObj, mid)
		if err != nil {
			return fmt.Errorf("getAbsolutePath: %w", err)
		}
		path = env.GoString((*jni.String)(unsafe.Pointer(pathObj)))
		return nil
	})
	return path, err
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	fmt.Fprintln(output, "=== Environment ===")

	// Environment is a static-methods-only class.
	// We create a zero-value struct with just the VM pointer.
	env := environment.Environment{VM: vm}

	// --- Directories ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Directories:")

	if fileObj, err := env.GetRootDirectory(); err != nil {
		fmt.Fprintf(output, "  Root: %v\n", err)
	} else if path, err := fileToPath(vm, fileObj); err != nil {
		fmt.Fprintf(output, "  Root: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Root: %s\n", path)
	}

	if fileObj, err := env.GetDataDirectory(); err != nil {
		fmt.Fprintf(output, "  Data: %v\n", err)
	} else if path, err := fileToPath(vm, fileObj); err != nil {
		fmt.Fprintf(output, "  Data: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Data: %s\n", path)
	}

	if fileObj, err := env.GetDownloadCacheDirectory(); err != nil {
		fmt.Fprintf(output, "  Cache: %v\n", err)
	} else if path, err := fileToPath(vm, fileObj); err != nil {
		fmt.Fprintf(output, "  Cache: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Cache: %s\n", path)
	}

	if fileObj, err := env.GetExternalStorageDirectory(); err != nil {
		fmt.Fprintf(output, "  ExtStorage: %v\n", err)
	} else if path, err := fileToPath(vm, fileObj); err != nil {
		fmt.Fprintf(output, "  ExtStorage: %v\n", err)
	} else {
		fmt.Fprintf(output, "  ExtStorage: %s\n", path)
	}

	if fileObj, err := env.GetStorageDirectory(); err != nil {
		fmt.Fprintf(output, "  Storage: %v\n", err)
	} else if path, err := fileToPath(vm, fileObj); err != nil {
		fmt.Fprintf(output, "  Storage: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Storage: %s\n", path)
	}

	// --- External Storage State ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "External storage state:")

	if state, err := env.GetExternalStorageState0(); err != nil {
		fmt.Fprintf(output, "  State: %v\n", err)
	} else {
		fmt.Fprintf(output, "  State: %s\n", state)
	}

	// --- State Constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "State constants:")
	fmt.Fprintf(output, "  Mounted:     %s\n", environment.MediaMounted)
	fmt.Fprintf(output, "  ReadOnly:    %s\n", environment.MediaMountedReadOnly)
	fmt.Fprintf(output, "  Removed:     %s\n", environment.MediaRemoved)
	fmt.Fprintf(output, "  Unmounted:   %s\n", environment.MediaUnmounted)
	fmt.Fprintf(output, "  BadRemoval:  %s\n", environment.MediaBadRemoval)
	fmt.Fprintf(output, "  Checking:    %s\n", environment.MediaChecking)

	// --- Storage Properties ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Storage properties:")

	if emulated, err := env.IsExternalStorageEmulated0(); err != nil {
		fmt.Fprintf(output, "  Emulated: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Emulated: %v\n", emulated)
	}

	if removable, err := env.IsExternalStorageRemovable0(); err != nil {
		fmt.Fprintf(output, "  Removable: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Removable: %v\n", removable)
	}

	if manager, err := env.IsExternalStorageManager0(); err != nil {
		fmt.Fprintf(output, "  Manager: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Manager: %v\n", manager)
	}

	if legacy, err := env.IsExternalStorageLegacy0(); err != nil {
		fmt.Fprintf(output, "  Legacy: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Legacy: %v\n", legacy)
	}

	return nil
}
