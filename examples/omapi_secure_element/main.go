//go:build android

// Command omapi_secure_element demonstrates using the OMAPI to check
// for secure element readers and list their properties.
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
	"github.com/AndroidGoLab/jni/se/omapi"
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
	fmt.Fprintln(output, "=== OMAPI Secure Element ===")
	fmt.Fprintln(output)

	// Create an SEService.
	svc, err := omapi.NewService(vm)
	if err != nil {
		return fmt.Errorf("omapi.NewService: %w", err)
	}
	defer svc.Close()
	defer svc.Shutdown()

	// Check connectivity.
	connected, err := svc.IsConnected()
	if err != nil {
		fmt.Fprintf(output, "isConnected: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "SE service connected: %v\n", connected)
	}

	// Get version.
	version, err := svc.GetVersion()
	if err != nil {
		fmt.Fprintf(output, "version: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "OMAPI version: %s\n", version)
	}

	// Get readers array.
	readersObj, err := svc.GetReaders()
	if err != nil {
		fmt.Fprintf(output, "getReaders: error: %v\n", err)
		fmt.Fprintln(output)
		fmt.Fprintln(output, "OMAPI secure element check complete.")
		return nil
	}

	// Iterate readers using JNI array access.
	fmt.Fprintln(output)
	if readersObj == nil {
		fmt.Fprintln(output, "readers: null")
	} else {
		vm.Do(func(env *jni.Env) error {
			arr := (*jni.Array)(unsafe.Pointer(readersObj))
			objArr := (*jni.ObjectArray)(unsafe.Pointer(readersObj))
			arrLen := env.GetArrayLength(arr)
			fmt.Fprintf(output, "readers found: %d\n", arrLen)

			for i := int32(0); i < arrLen; i++ {
				elemObj, err := env.GetObjectArrayElement(objArr, i)
				if err != nil || elemObj == nil {
					continue
				}
				reader := omapi.Reader{VM: vm, Obj: env.NewGlobalRef(elemObj)}

				name, err := reader.GetName()
				if err != nil {
					fmt.Fprintf(output, "  [%d] name: error: %v\n", i, err)
				} else {
					fmt.Fprintf(output, "  [%d] name: %s\n", i, name)
				}

				present, err := reader.IsSecureElementPresent()
				if err != nil {
					fmt.Fprintf(output, "  [%d] SE present: error: %v\n", i, err)
				} else {
					fmt.Fprintf(output, "  [%d] SE present: %v\n", i, present)
				}

				vm.Do(func(env *jni.Env) error {
					env.DeleteGlobalRef(reader.Obj)
					return nil
				})
			}
			return nil
		})
	}

	// Show API surface.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "SE API hierarchy:")
	fmt.Fprintln(output, "  SEService")
	fmt.Fprintln(output, "    -> Reader[]")
	fmt.Fprintln(output, "      -> Session")
	fmt.Fprintln(output, "        -> Channel (basic/logical)")

	fmt.Fprintln(output)
	fmt.Fprintln(output, "OMAPI secure element check complete.")
	return nil
}
