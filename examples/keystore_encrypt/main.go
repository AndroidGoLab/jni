//go:build android

// Command keystore_encrypt demonstrates the Android KeyStore typed wrapper
// constants for key management parameters: purposes, algorithms, block modes,
// encryption padding, and digests.
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
	"github.com/AndroidGoLab/jni/security/keystore"
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
	_ = vm // VM is available but not needed for constant display.

	fmt.Fprintln(output, "=== KeyStore Encrypt ===")
	fmt.Fprintln(output)

	// --- KeyGenParameterSpec purpose constants ---
	fmt.Fprintln(output, "KeyGenParameterSpec constants:")
	fmt.Fprintf(output, "  PURPOSE_ENCRYPT   = %d\n", keystore.PurposeEncrypt)
	fmt.Fprintf(output, "  PURPOSE_DECRYPT   = %d\n", keystore.PurposeDecrypt)
	fmt.Fprintf(output, "  PURPOSE_SIGN      = %d\n", keystore.PurposeSign)
	fmt.Fprintf(output, "  PURPOSE_VERIFY    = %d\n", keystore.PurposeVerify)
	fmt.Fprintf(output, "  PURPOSE_WRAP_KEY  = %d\n", keystore.PurposeWrapKey)
	fmt.Fprintf(output, "  PURPOSE_AGREE_KEY = %d\n", keystore.PurposeAgreeKey)
	fmt.Fprintf(output, "  PURPOSE_ATTEST_KEY= %d\n", keystore.PurposeAttestKey)

	// --- Key algorithm constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Key algorithms:")
	fmt.Fprintf(output, "  AES         = %s\n", keystore.KeyAlgorithmAes)
	fmt.Fprintf(output, "  RSA         = %s\n", keystore.KeyAlgorithmRsa)
	fmt.Fprintf(output, "  EC          = %s\n", keystore.KeyAlgorithmEc)
	fmt.Fprintf(output, "  3DES        = %s\n", keystore.KeyAlgorithm3des)
	fmt.Fprintf(output, "  HMAC_SHA1   = %s\n", keystore.KeyAlgorithmHmacSha1)
	fmt.Fprintf(output, "  HMAC_SHA224 = %s\n", keystore.KeyAlgorithmHmacSha224)
	fmt.Fprintf(output, "  HMAC_SHA256 = %s\n", keystore.KeyAlgorithmHmacSha256)
	fmt.Fprintf(output, "  HMAC_SHA384 = %s\n", keystore.KeyAlgorithmHmacSha384)
	fmt.Fprintf(output, "  HMAC_SHA512 = %s\n", keystore.KeyAlgorithmHmacSha512)

	// --- Block mode constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Block modes:")
	fmt.Fprintf(output, "  CBC = %s\n", keystore.BlockModeCbc)
	fmt.Fprintf(output, "  GCM = %s\n", keystore.BlockModeGcm)
	fmt.Fprintf(output, "  CTR = %s\n", keystore.BlockModeCtr)
	fmt.Fprintf(output, "  ECB = %s\n", keystore.BlockModeEcb)

	// --- Encryption padding constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Encryption padding:")
	fmt.Fprintf(output, "  NONE      = %s\n", keystore.EncryptionPaddingNone)
	fmt.Fprintf(output, "  PKCS7     = %s\n", keystore.EncryptionPaddingPkcs7)
	fmt.Fprintf(output, "  RSA_OAEP  = %s\n", keystore.EncryptionPaddingRsaOaep)
	fmt.Fprintf(output, "  RSA_PKCS1 = %s\n", keystore.EncryptionPaddingRsaPkcs1)

	// --- Digest constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Digests:")
	fmt.Fprintf(output, "  NONE   = %s\n", keystore.DigestNone)
	fmt.Fprintf(output, "  MD5    = %s\n", keystore.DigestMd5)
	fmt.Fprintf(output, "  SHA1   = %s\n", keystore.DigestSha1)
	fmt.Fprintf(output, "  SHA224 = %s\n", keystore.DigestSha224)
	fmt.Fprintf(output, "  SHA256 = %s\n", keystore.DigestSha256)
	fmt.Fprintf(output, "  SHA384 = %s\n", keystore.DigestSha384)
	fmt.Fprintf(output, "  SHA512 = %s\n", keystore.DigestSha512)

	// --- Signature padding constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Signature padding:")
	fmt.Fprintf(output, "  RSA_PKCS1 = %s\n", keystore.SignaturePaddingRsaPkcs1)
	fmt.Fprintf(output, "  RSA_PSS   = %s\n", keystore.SignaturePaddingRsaPss)

	// --- Security level constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Security levels:")
	fmt.Fprintf(output, "  SOFTWARE            = %d\n", keystore.SecurityLevelSoftware)
	fmt.Fprintf(output, "  TRUSTED_ENVIRONMENT = %d\n", keystore.SecurityLevelTrustedEnvironment)
	fmt.Fprintf(output, "  STRONGBOX           = %d\n", keystore.SecurityLevelStrongbox)
	fmt.Fprintf(output, "  UNKNOWN             = %d\n", keystore.SecurityLevelUnknown)
	fmt.Fprintf(output, "  UNKNOWN_SECURE      = %d\n", keystore.SecurityLevelUnknownSecure)

	// --- Origin constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Key origins:")
	fmt.Fprintf(output, "  GENERATED        = %d\n", keystore.OriginGenerated)
	fmt.Fprintf(output, "  IMPORTED         = %d\n", keystore.OriginImported)
	fmt.Fprintf(output, "  SECURELY_IMPORTED= %d\n", keystore.OriginSecurelyImported)
	fmt.Fprintf(output, "  UNKNOWN          = %d\n", keystore.OriginUnknown)

	return nil
}
