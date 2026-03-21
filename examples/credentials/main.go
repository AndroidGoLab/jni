//go:build android

// Command credentials demonstrates the Android Credential Manager API data types.
// It is built as a c-shared library and packaged into an APK.
//
// This example shows the exported data class types PasswordCredential and
// PublicKeyCredential, which are used to represent credential data extracted
// from JNI objects. The Manager type and its methods (createManagerRaw,
// getCredentialRaw, createCredentialRaw, clearCredentialStateRaw) are
// unexported and used internally by higher-level wrappers.
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

	"github.com/AndroidGoLab/jni/credentials"
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
	// PasswordCredential holds data extracted from a
	// androidx.credentials.PasswordCredential JNI object.
	// It has two exported fields: ID and Password.
	pwCred := credentials.PasswordCredential{
		ID:       "user@example.com",
		Password: "s3cret",
	}
	fmt.Fprintf(output, "PasswordCredential ID: %s\n", pwCred.ID)
	fmt.Fprintf(output, "PasswordCredential Password: %s\n", pwCred.Password)

	// PublicKeyCredential holds data extracted from a
	// androidx.credentials.publickeycredential.PublicKeyCredential JNI object.
	// It has one exported field: AuthResponseJSON.
	pkCred := credentials.PublicKeyCredential{
		AuthResponseJSON: `{"type":"public-key","id":"abc123","response":{"clientDataJSON":"...","authenticatorData":"..."}}`,
	}
	fmt.Fprintf(output, "PublicKeyCredential AuthResponseJSON: %s\n", pkCred.AuthResponseJSON)

	// The following types and methods are unexported (package-internal):
	//
	// Manager wraps androidx.credentials.CredentialManager with methods:
	//   Manager.createManagerRaw(ctx *jni.Object) (*jni.Object, error)
	//     Static factory method to create a CredentialManager instance.
	//
	//   Manager.getCredentialRaw(ctx, request *jni.Object) (*jni.Object, error)
	//     Retrieves a credential using the given request.
	//
	//   Manager.createCredentialRaw(ctx, request *jni.Object) (*jni.Object, error)
	//     Creates/saves a credential using the given request.
	//
	//   Manager.clearCredentialStateRaw(request *jni.Object) error
	//     Clears the credential state.
	//
	// getCredentialRequestBuilder wraps GetCredentialRequest.Builder:
	//   addCredentialOption(option *jni.Object) *jni.Object
	//   build() *jni.Object
	//
	// getCredentialResponse wraps GetCredentialResponse (empty struct).
	//
	// extractPasswordCredential(env, obj) extracts PasswordCredential fields.
	// extractPublicKeyCredential(env, obj) extracts PublicKeyCredential fields.
	return nil
}
