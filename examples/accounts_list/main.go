//go:build android

// Command accounts_list uses AccountManager to list all accounts on
// the device (Google, Samsung, etc.) and prints account name and type.
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
	"github.com/AndroidGoLab/jni/accounts"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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

	mgr, err := accounts.NewAccountManager(ctx)
	if err != nil {
		return fmt.Errorf("NewAccountManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Account List ===")

	acctArray, err := mgr.GetAccounts()
	if err != nil {
		fmt.Fprintf(output, "GetAccounts: %v\n", err)
		return nil
	}

	var acctCount int32
	err = vm.Do(func(env *jni.Env) error {
		if acctArray == nil {
			return nil
		}
		acctCount = env.GetArrayLength((*jni.Array)(unsafe.Pointer(acctArray)))
		return nil
	})
	if err != nil {
		return fmt.Errorf("get array length: %w", err)
	}

	fmt.Fprintf(output, "Total accounts: %d\n\n", acctCount)

	for i := int32(0); i < acctCount; i++ {
		acct := accounts.Account{VM: vm}
		err := vm.Do(func(env *jni.Env) error {
			elem, err := env.GetObjectArrayElement((*jni.ObjectArray)(unsafe.Pointer(acctArray)), i)
			if err != nil {
				return fmt.Errorf("get element %d: %w", i, err)
			}
			acct.Obj = env.NewGlobalRef(elem)
			return nil
		})
		if err != nil {
			fmt.Fprintf(output, "  [%d] error: %v\n", i, err)
			continue
		}

		str, err := acct.ToString()
		if err != nil {
			fmt.Fprintf(output, "  [%d] ToString err: %v\n", i, err)
		} else {
			fmt.Fprintf(output, "  [%d] %s\n", i, str)
		}

		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(acct.Obj)
			return nil
		})
	}

	// List authenticator types.
	authArray, err := mgr.GetAuthenticatorTypes()
	if err != nil {
		fmt.Fprintf(output, "\nGetAuthenticatorTypes: %v\n", err)
		return nil
	}

	var authCount int32
	_ = vm.Do(func(env *jni.Env) error {
		if authArray == nil {
			return nil
		}
		authCount = env.GetArrayLength((*jni.Array)(unsafe.Pointer(authArray)))
		return nil
	})

	fmt.Fprintf(output, "\nAuthenticator types: %d\n", authCount)
	fmt.Fprintln(output, "\nAccounts list complete.")
	return nil
}
