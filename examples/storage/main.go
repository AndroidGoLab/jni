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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/os/storage"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	vm := jni.VMFromPtr(unsafe.Pointer(cvm))
	if err := run(vm); err != nil {
		fmt.Fprintf(&output, "ERROR: %v\n", err)
	}
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}

func run(vm *jni.VM) error {
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

	fmt.Fprintln(&output, "StorageManager obtained successfully")
	fmt.Fprintln(&output, "Unexported methods: getStorageVolumes, getPrimaryStorageVolume, getAllocatableBytes, allocateBytes, getCacheSizeBytes, getCacheQuotaBytes")

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
