//go:build android

// Command keystore demonstrates the Android KeyStore JNI bindings.
// It is built as a c-shared library and packaged into an APK.
//
// The keystore package wraps java.security.KeyStore and related
// Android keystore classes. All methods are unexported and intended
// to be called via higher-level wrappers.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	_ "github.com/AndroidGoLab/jni/security/keystore"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
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
	fmt.Fprintln(&output, "keystore bindings available for key management operations")
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
