//go:build android

// Command accounts demonstrates querying Android device accounts
// using the accounts package, which wraps android.accounts.AccountManager.
// It is built as a c-shared library and packaged into an APK using
// the shared apk.mk infrastructure.
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
	"github.com/AndroidGoLab/jni/accounts"
	"github.com/AndroidGoLab/jni/app"
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
	_, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}

	// The accounts package wraps android.accounts.AccountManager.
	// The Manager type provides access to device accounts through
	// unexported methods (getAccountsRaw, getAccountsByTypeRaw,
	// getAuthTokenRaw, invalidateAuthTokenRaw) which are intended
	// to be wrapped by higher-level helpers.
	//
	// The Manager struct has exported fields:
	//   VM  *jni.VM
	//   Obj *jni.GlobalRef
	//
	// Manager is obtained via the static factory AccountManager.get(Context),
	// exposed as the unexported getManagerRaw method.

	// Account is a data class with exported fields extracted from
	// android.accounts.Account Java objects.
	var acct accounts.Account
	fmt.Fprintf(&output, "Account fields: Name=%q, Type=%q\n", acct.Name, acct.Type)

	fmt.Fprintln(&output, "AccountManager raw methods: getManagerRaw, getAccountsRaw, getAccountsByTypeRaw, getAuthTokenRaw, invalidateAuthTokenRaw")

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
