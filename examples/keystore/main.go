//go:build android

// Command keystore demonstrates the Android KeyStore JNI bindings.
// It is built as a c-shared library and packaged into an APK.
//
// The keystore package wraps java.security.KeyStore and related
// Android keystore classes. All methods are unexported and intended
// to be called via higher-level wrappers.
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
	"unsafe"
	"bytes"
	"fmt"

	_ "github.com/AndroidGoLab/jni/security/keystore"
	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
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
	// The keystore package provides wrappers for key management:
	//
	// keyStoreJava (java.security.KeyStore):
	//   - load, containsAlias, deleteEntry, aliasesRaw, getEntry
	//
	// keyGenParamBuilder (KeyGenParameterSpec.Builder):
	//   - setKeySize, setBlockModes, setEncryptionPaddings, etc.
	//
	// keyGeneratorJava, keyPairGeneratorJava, cipherJava, signatureJava
	//
	// All methods are unexported and intended for higher-level wrappers.
	fmt.Fprintln(output, "keystore bindings available for key management operations")
	return nil
}
