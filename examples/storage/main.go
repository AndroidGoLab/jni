//go:build android

// Command storage demonstrates using the Android StorageManager
// system service, wrapped by the storage package. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// The storage package wraps android.os.storage.StorageManager and
// provides the Volume data class for inspecting storage volumes.
// It supports querying storage volumes, checking allocatable bytes,
// and managing cache quotas.
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
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/os/storage"
)

func main() {}

func init() { exampleui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	exampleui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	exampleui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	exampleui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := storage.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewManager: %w", err)
	}

	// Manager provides unexported methods for storage management:
	//   getStorageVolumes()                 -- returns all storage volumes.
	//   getPrimaryStorageVolume()           -- returns the primary storage volume.
	//   getAllocatableBytes(storageUuid)     -- bytes available for allocation.
	//   allocateBytes(storageUuid, bytes)   -- allocates storage space.
	//   getCacheSizeBytes(storageUuid)      -- current cache size.
	//   getCacheQuotaBytes(storageUuid)     -- cache quota for the app.

	// The storageVolume type (unexported) wraps android.os.storage.StorageVolume
	// with JNI methods for querying volume properties.

	fmt.Fprintln(output, "StorageManager obtained successfully")
	fmt.Fprintln(output, "Unexported methods: getStorageVolumes, getPrimaryStorageVolume, getAllocatableBytes, allocateBytes, getCacheSizeBytes, getCacheQuotaBytes")

	_ = mgr
	return nil
}

// getAppContext obtains an Android Context via ActivityThread.currentApplication().
func getAppContext(vm *jni.VM) (*app.Context, error) {
	var ctx app.Context
	ctx.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}

		atClass, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread: %w", err)
		}

		curAppMid, err := env.GetStaticMethodID(atClass, "currentApplication", "()Landroid/app/Application;")
		if err != nil {
			return fmt.Errorf("get currentApplication: %w", err)
		}
		appObj, err := env.CallStaticObjectMethod(atClass, curAppMid)
		if err != nil {
			return fmt.Errorf("call currentApplication: %w", err)
		}
		if appObj == nil || appObj.Ref() == 0 {
			return fmt.Errorf("currentApplication returned null")
		}

		ctx.Obj = env.NewGlobalRef(appObj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}
