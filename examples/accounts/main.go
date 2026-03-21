//go:build android

// Command accounts demonstrates querying Android device accounts
// using the accounts package, which wraps android.accounts.AccountManager.
// It is built as a c-shared library and packaged into an APK using
// the shared apk.mk infrastructure.
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
	"github.com/AndroidGoLab/jni/accounts"
	"github.com/AndroidGoLab/jni/app"
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
	// Account wraps android.accounts.Account. Its fields (VM, Obj) hold
	// references to the Java object. Name and Type are accessed via JNI
	// methods (DescribeContents, Equals, HashCode, etc.).
	var acct accounts.Account
	_ = acct
	fmt.Fprintln(output, "Account type available with DescribeContents, Equals, HashCode methods")

	fmt.Fprintln(output, "AccountManager raw methods: getManagerRaw, getAccountsRaw, getAccountsByTypeRaw, getAuthTokenRaw, invalidateAuthTokenRaw")

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
