//go:build android

// Command keystore demonstrates the Android KeyStore JNI bindings.
// It is built as a c-shared library and packaged into an APK.
//
// This example shows how to use the keystore package for cryptographic
// key management on Android. It demonstrates the purpose constants,
// the keyGenParamBuilder for configuring key generation parameters,
// and the various crypto wrappers (KeyStore, KeyGenerator, Cipher,
// Signature, GCMParameterSpec).
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/AndroidGoLab/jni/security/keystore"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// Key purpose constants define what operations a key may be used for.
	// These are bitfield values from KeyProperties.PURPOSE_*.
	fmt.Fprintln(&output, "Key purpose constants:")
	fmt.Fprintf(&output, "  PurposeEncrypt = %d\n", keystore.PurposeEncrypt)
	fmt.Fprintf(&output, "  PurposeDecrypt = %d\n", keystore.PurposeDecrypt)
	fmt.Fprintf(&output, "  PurposeSign    = %d\n", keystore.PurposeSign)
	fmt.Fprintf(&output, "  PurposeVerify  = %d\n", keystore.PurposeVerify)

	// Purposes can be combined with bitwise OR for keys that serve
	// multiple roles.
	encryptDecrypt := keystore.PurposeEncrypt | keystore.PurposeDecrypt
	fmt.Fprintf(&output, "  Encrypt|Decrypt = %d\n", encryptDecrypt)

	signVerify := keystore.PurposeSign | keystore.PurposeVerify
	fmt.Fprintf(&output, "  Sign|Verify     = %d\n", signVerify)

	// The keystore package provides wrappers for these Java classes:
	//
	// keyStoreJava (java.security.KeyStore):
	//   - load(param)           - load the keystore
	//   - containsAlias(alias)  - check if an alias exists
	//   - deleteEntry(alias)    - delete a key entry
	//   - aliasesRaw()          - list all aliases
	//   - getEntry(alias, param)- retrieve a key entry
	//
	// keyGenParamBuilder (android.security.keystore.KeyGenParameterSpec.Builder):
	//   - setKeySize(size)
	//   - setBlockModes(modes)
	//   - setEncryptionPaddings(paddings)
	//   - setSignaturePaddings(paddings)
	//   - setDigests(digests)
	//   - setUserAuthenticationRequired(required)
	//   - setUserAuthenticationValidityDurationSeconds(seconds)
	//   - setInvalidatedByBiometricEnrollment(invalidated)
	//   - setUnlockedDeviceRequired(required)
	//   - build()
	//
	// keyGeneratorJava (javax.crypto.KeyGenerator):
	//   - init(params)          - initialize with algorithm parameters
	//   - generateKey()         - generate a secret key
	//
	// keyPairGeneratorJava (java.security.KeyPairGenerator):
	//   - initialize(params)    - initialize with algorithm parameters
	//   - generateKeyPair()     - generate a key pair
	//
	// cipherJava (javax.crypto.Cipher):
	//   - initWithKey(opmode, key)
	//   - initWithKeyAndParams(opmode, key, params)
	//   - doFinal(input)        - encrypt or decrypt data
	//   - getIV()               - get the initialization vector
	//
	// signatureJava (java.security.Signature):
	//   - initSign(privateKey)  - initialize for signing
	//   - initVerify(publicKey) - initialize for verification
	//   - update(data)          - feed data to sign/verify
	//   - sign()                - produce the signature
	//   - verify(signature)     - verify a signature
	//
	// gcmParamSpec (javax.crypto.spec.GCMParameterSpec):
	//   - created via NewgcmParamSpec(vm)

	fmt.Fprintln(&output, "keystore bindings available for key management operations")
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
