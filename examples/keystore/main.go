//go:build android

// Command keystore demonstrates the Android KeyStore API.
// It loads the AndroidKeyStore, lists existing key aliases,
// and checks for a test alias. The generated keystore package
// types are unexported, so this example uses raw JNI for
// KeyStore.getInstance() and iterates aliases via Enumeration.
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
	_ "github.com/AndroidGoLab/jni/security/keystore"
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
	fmt.Fprintln(output, "=== Android KeyStore ===")
	fmt.Fprintln(output)

	var aliases []string
	err := vm.Do(func(env *jni.Env) error {
		// KeyStore.getInstance("AndroidKeyStore")
		ksCls, err := env.FindClass("java/security/KeyStore")
		if err != nil {
			return fmt.Errorf("find KeyStore: %w", err)
		}

		getInstMid, err := env.GetStaticMethodID(
			ksCls, "getInstance",
			"(Ljava/lang/String;)Ljava/security/KeyStore;",
		)
		if err != nil {
			return fmt.Errorf("get getInstance: %w", err)
		}

		jType, err := env.NewStringUTF("AndroidKeyStore")
		if err != nil {
			return fmt.Errorf("NewStringUTF: %w", err)
		}

		ksObj, err := env.CallStaticObjectMethod(
			ksCls, getInstMid,
			jni.ObjectValue(&jType.Object),
		)
		if err != nil {
			return fmt.Errorf("getInstance: %w", err)
		}

		// ks.load(null) -- required before use.
		loadMid, err := env.GetMethodID(
			ksCls, "load",
			"(Ljava/security/KeyStore$LoadStoreParameter;)V",
		)
		if err != nil {
			return fmt.Errorf("get load: %w", err)
		}
		if err := env.CallVoidMethod(ksObj, loadMid, jni.ObjectValue((*jni.Object)(nil))); err != nil {
			return fmt.Errorf("load: %w", err)
		}
		fmt.Fprintln(output, "KeyStore loaded OK")

		// ks.size()
		sizeMid, err := env.GetMethodID(ksCls, "size", "()I")
		if err != nil {
			return fmt.Errorf("get size: %w", err)
		}
		size, err := env.CallIntMethod(ksObj, sizeMid)
		if err != nil {
			return fmt.Errorf("size: %w", err)
		}
		fmt.Fprintf(output, "keys: %d\n", size)
		fmt.Fprintln(output)

		// Iterate aliases via Enumeration.
		aliasesMid, err := env.GetMethodID(ksCls, "aliases", "()Ljava/util/Enumeration;")
		if err != nil {
			return fmt.Errorf("get aliases: %w", err)
		}
		enumObj, err := env.CallObjectMethod(ksObj, aliasesMid)
		if err != nil {
			return fmt.Errorf("aliases: %w", err)
		}

		enumCls := env.GetObjectClass(enumObj)
		hasMoreMid, err := env.GetMethodID(enumCls, "hasMoreElements", "()Z")
		if err != nil {
			return fmt.Errorf("get hasMoreElements: %w", err)
		}
		nextMid, err := env.GetMethodID(enumCls, "nextElement", "()Ljava/lang/Object;")
		if err != nil {
			return fmt.Errorf("get nextElement: %w", err)
		}

		for {
			hasMore, err := env.CallBooleanMethod(enumObj, hasMoreMid)
			if err != nil {
				return fmt.Errorf("hasMoreElements: %w", err)
			}
			if hasMore == 0 {
				break
			}

			elemObj, err := env.CallObjectMethod(enumObj, nextMid)
			if err != nil {
				return fmt.Errorf("nextElement: %w", err)
			}
			alias := env.GoString((*jni.String)(unsafe.Pointer(elemObj)))
			aliases = append(aliases, alias)
		}

		// Check for a specific test alias.
		containsMid, err := env.GetMethodID(ksCls, "containsAlias", "(Ljava/lang/String;)Z")
		if err != nil {
			return fmt.Errorf("get containsAlias: %w", err)
		}
		jTestAlias, _ := env.NewStringUTF("go-jni-test-key")
		hasTest, err := env.CallBooleanMethod(ksObj, containsMid, jni.ObjectValue(&jTestAlias.Object))
		if err != nil {
			return fmt.Errorf("containsAlias: %w", err)
		}
		if hasTest != 0 {
			fmt.Fprintln(output, "test key: present")
		} else {
			fmt.Fprintln(output, "test key: absent")
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("keystore: %w", err)
	}

	if len(aliases) == 0 {
		fmt.Fprintln(output, "(no aliases)")
	} else {
		fmt.Fprintln(output, "aliases:")
		for _, a := range aliases {
			fmt.Fprintf(output, "  %s\n", a)
		}
	}

	return nil
}
