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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/xaionaro-go/jni/credentials"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// PasswordCredential holds data extracted from a
	// androidx.credentials.PasswordCredential JNI object.
	// It has two exported fields: ID and Password.
	pwCred := credentials.PasswordCredential{
		ID:       "user@example.com",
		Password: "s3cret",
	}
	fmt.Fprintf(&output, "PasswordCredential ID: %s\n", pwCred.ID)
	fmt.Fprintf(&output, "PasswordCredential Password: %s\n", pwCred.Password)

	// PublicKeyCredential holds data extracted from a
	// androidx.credentials.publickeycredential.PublicKeyCredential JNI object.
	// It has one exported field: AuthResponseJSON.
	pkCred := credentials.PublicKeyCredential{
		AuthResponseJSON: `{"type":"public-key","id":"abc123","response":{"clientDataJSON":"...","authenticatorData":"..."}}`,
	}
	fmt.Fprintf(&output, "PublicKeyCredential AuthResponseJSON: %s\n", pkCred.AuthResponseJSON)

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
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
