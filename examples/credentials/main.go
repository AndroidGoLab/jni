//go:build android

// Command credentials demonstrates the Android Credential Manager API.
// It initializes the JNI bindings, attempts to create a CredentialManager
// instance via JNI, and reports whether the credentials framework is
// available on the device.
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
	"github.com/AndroidGoLab/jni/credentials"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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
	fmt.Fprintln(output, "=== Credential Manager ===")

	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// Initialize JNI class and method references for the credentials package.
	// This resolves all androidx.credentials.* classes.
	var initErr error
	vm.Do(func(env *jni.Env) error {
		initErr = credentials.Init(env)
		return nil
	})

	if initErr != nil {
		fmt.Fprintf(output, "Init: %v\n", initErr)
		fmt.Fprintln(output, "Credential Manager NOT available.")
		fmt.Fprintln(output, "(Requires androidx.credentials)")
		return nil
	}
	fmt.Fprintln(output, "Init: OK")
	fmt.Fprintln(output, "JNI classes resolved:")
	fmt.Fprintln(output, "  CredentialManager")
	fmt.Fprintln(output, "  GetCredentialRequest$Builder")
	fmt.Fprintln(output, "  GetCredentialResponse")
	fmt.Fprintln(output, "  PasswordCredential")
	fmt.Fprintln(output, "  PublicKeyCredential")

	// Create a CredentialManager via the static factory.
	// The wrapper methods are unexported, so we call
	// CredentialManager.create(context) directly via JNI.
	var mgrObj *jni.GlobalRef
	var createErr error
	vm.Do(func(env *jni.Env) error {
		cmClass, err := env.FindClass("androidx/credentials/CredentialManager")
		if err != nil {
			createErr = fmt.Errorf("find class: %w", err)
			return nil
		}
		createMid, err := env.GetStaticMethodID(cmClass, "create",
			"(Landroid/content/Context;)Landroidx/credentials/CredentialManager;")
		if err != nil {
			createErr = fmt.Errorf("get create: %w", err)
			return nil
		}
		obj, err := env.CallStaticObjectMethod(cmClass, createMid,
			jni.ObjectValue(ctx.Obj))
		if err != nil {
			createErr = fmt.Errorf("call create: %w", err)
			return nil
		}
		if obj != nil {
			mgrObj = env.NewGlobalRef(obj)
		}
		return nil
	})

	fmt.Fprintln(output)
	if createErr != nil {
		fmt.Fprintf(output, "create: %v\n", createErr)
	} else if mgrObj == nil || mgrObj.Ref() == 0 {
		fmt.Fprintln(output, "create: returned null")
	} else {
		fmt.Fprintln(output, "CredentialManager.create: OK")
		fmt.Fprintf(output, "  ref: %d\n", mgrObj.Ref())

		// Attempt clearCredentialState via raw JNI.
		// This exercises the method binding. It will fail
		// because we pass null for the request, but it
		// proves the JNI plumbing works.
		vm.Do(func(env *jni.Env) error {
			cmClass, err := env.FindClass("androidx/credentials/CredentialManager")
			if err != nil {
				fmt.Fprintf(output, "clearCredentialState: %v\n", err)
				return nil
			}
			clearMid, err := env.GetMethodID(cmClass, "clearCredentialState",
				"(Landroidx/credentials/ClearCredentialStateRequest;)V")
			if err != nil {
				fmt.Fprintf(output, "clearCredentialState: %v\n", err)
				fmt.Fprintln(output, "  (method not found)")
				return nil
			}
			err = env.CallVoidMethod(mgrObj, clearMid, jni.ObjectValue(nil))
			if err != nil {
				fmt.Fprintf(output, "clearCredentialState: %v\n", err)
				fmt.Fprintln(output, "  (expected: needs request)")
			} else {
				fmt.Fprintln(output, "clearCredentialState: OK")
			}
			return nil
		})

		// Inspect the manager object's class name.
		vm.Do(func(env *jni.Env) error {
			cls := env.GetObjectClass(mgrObj)
			getNameMid, err := env.GetMethodID(cls, "getClass", "()Ljava/lang/Class;")
			if err != nil {
				return nil
			}
			classObj, err := env.CallObjectMethod(mgrObj, getNameMid)
			if err != nil || classObj == nil {
				return nil
			}
			classCls := env.GetObjectClass(classObj)
			nameMid, err := env.GetMethodID(classCls, "getName", "()Ljava/lang/String;")
			if err != nil {
				return nil
			}
			nameObj, err := env.CallObjectMethod(classObj, nameMid)
			if err != nil || nameObj == nil {
				return nil
			}
			name := env.GoString((*jni.String)(unsafe.Pointer(nameObj)))
			fmt.Fprintf(output, "  class: %s\n", name)
			return nil
		})

		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(mgrObj)
			return nil
		})
	}

	// Show extracted data types.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Data types:")
	fmt.Fprintln(output, "  PasswordCredential:")
	fmt.Fprintln(output, "    Fields: ID, Password")
	fmt.Fprintln(output, "  PublicKeyCredential:")
	fmt.Fprintln(output, "    Fields: AuthResponseJSON")

	return nil
}
