//go:build android

// Command env_path_resolver demonstrates the Environment API to query
// standard Android paths, external storage state, and media state
// constants using typed wrappers.
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
	"github.com/AndroidGoLab/jni/os/environment"
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
	fmt.Fprintln(output, "=== Environment Path Resolver ===")

	// Environment is a static-methods-only class.
	env := environment.Environment{VM: vm}

	// --- Standard directories (returned as File objects) ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Standard directories (File objects):")

	type dirQuery struct {
		name string
		fn   func() (*jni.Object, error)
	}
	dirs := []dirQuery{
		{"Root", env.GetRootDirectory},
		{"Data", env.GetDataDirectory},
		{"DownloadCache", env.GetDownloadCacheDirectory},
		{"ExternalStorage", env.GetExternalStorageDirectory},
		{"Storage", env.GetStorageDirectory},
	}

	for _, d := range dirs {
		fileObj, err := d.fn()
		if err != nil {
			fmt.Fprintf(output, "  %-16s error: %v\n", d.name+":", err)
			continue
		}
		if fileObj != nil {
			fmt.Fprintf(output, "  %-16s OK\n", d.name+":")
		} else {
			fmt.Fprintf(output, "  %-16s null\n", d.name+":")
		}
	}

	// --- Public directories ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Public directories (File objects):")

	publicDirs := []string{
		"Music", "Podcasts", "Ringtones", "Alarms",
		"Notifications", "Pictures", "Movies", "Download",
		"DCIM", "Documents",
	}
	for _, dirType := range publicDirs {
		fileObj, err := env.GetExternalStoragePublicDirectory(dirType)
		if err != nil {
			fmt.Fprintf(output, "  %-14s error: %v\n", dirType+":", err)
			continue
		}
		if fileObj != nil {
			fmt.Fprintf(output, "  %-14s OK\n", dirType+":")
		} else {
			fmt.Fprintf(output, "  %-14s null\n", dirType+":")
		}
	}

	// --- External storage state ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "External storage:")

	state, err := env.GetExternalStorageState0()
	if err != nil {
		fmt.Fprintf(output, "  state: %v\n", err)
	} else {
		fmt.Fprintf(output, "  state: %s\n", state)
	}

	emulated, err := env.IsExternalStorageEmulated0()
	if err != nil {
		fmt.Fprintf(output, "  emulated: %v\n", err)
	} else {
		fmt.Fprintf(output, "  emulated: %v\n", emulated)
	}

	removable, err := env.IsExternalStorageRemovable0()
	if err != nil {
		fmt.Fprintf(output, "  removable: %v\n", err)
	} else {
		fmt.Fprintf(output, "  removable: %v\n", removable)
	}

	manager, err := env.IsExternalStorageManager0()
	if err != nil {
		fmt.Fprintf(output, "  manager: %v\n", err)
	} else {
		fmt.Fprintf(output, "  manager: %v\n", manager)
	}

	legacy, err := env.IsExternalStorageLegacy0()
	if err != nil {
		fmt.Fprintf(output, "  legacy: %v\n", err)
	} else {
		fmt.Fprintf(output, "  legacy: %v\n", legacy)
	}

	// --- Media state constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Media state constants:")
	fmt.Fprintf(output, "  MOUNTED       = %s\n", environment.MediaMounted)
	fmt.Fprintf(output, "  MOUNTED_RO    = %s\n", environment.MediaMountedReadOnly)
	fmt.Fprintf(output, "  REMOVED       = %s\n", environment.MediaRemoved)
	fmt.Fprintf(output, "  UNMOUNTED     = %s\n", environment.MediaUnmounted)
	fmt.Fprintf(output, "  BAD_REMOVAL   = %s\n", environment.MediaBadRemoval)
	fmt.Fprintf(output, "  CHECKING      = %s\n", environment.MediaChecking)
	fmt.Fprintf(output, "  EJECTING      = %s\n", environment.MediaEjecting)
	fmt.Fprintf(output, "  UNKNOWN       = %s\n", environment.MediaUnknown)

	return nil
}
