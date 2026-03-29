//go:build android

// Command keystore_encrypt demonstrates the Android KeyStore typed wrapper
// by obtaining the KeyStoreManager system service and calling its methods,
// and by querying supplementary attestation info for each security level.
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
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== KeyStore Encrypt ===")
	fmt.Fprintln(output)

	// 1. Obtain the KeyStoreManager system service.
	mgr, err := keystore.NewKeyStoreManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "KeyStoreManager: %v\n", err)
		fmt.Fprintln(output, "(Requires API 35+, falling back to constant display)")
		fmt.Fprintln(output)

		// Even without the manager, show that constants are real values
		// by exercising them (these are compile-time constants from the wrapper).
		fmt.Fprintf(output, "PURPOSE_ENCRYPT   = %d\n", keystore.PurposeEncrypt)
		fmt.Fprintf(output, "PURPOSE_DECRYPT   = %d\n", keystore.PurposeDecrypt)
		fmt.Fprintf(output, "PURPOSE_SIGN      = %d\n", keystore.PurposeSign)
		fmt.Fprintf(output, "PURPOSE_VERIFY    = %d\n", keystore.PurposeVerify)
		fmt.Fprintf(output, "PURPOSE_WRAP_KEY  = %d\n", keystore.PurposeWrapKey)
		fmt.Fprintf(output, "PURPOSE_AGREE_KEY = %d\n", keystore.PurposeAgreeKey)
		fmt.Fprintf(output, "PURPOSE_ATTEST_KEY= %d\n", keystore.PurposeAttestKey)
		return nil
	}
	defer mgr.Close()

	fmt.Fprintln(output, "KeyStoreManager: obtained OK")

	// 2. ToString
	str, err := mgr.ToString()
	if err != nil {
		fmt.Fprintf(output, "  ToString: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  ToString: %s\n", str)
	}

	// 3. GetSupplementaryAttestationInfo for each security level.
	type secLevel struct {
		name  string
		level int32
	}
	levels := []secLevel{
		{"SOFTWARE", int32(keystore.SecurityLevelSoftware)},
		{"TRUSTED_ENVIRONMENT", int32(keystore.SecurityLevelTrustedEnvironment)},
		{"STRONGBOX", int32(keystore.SecurityLevelStrongbox)},
		{"UNKNOWN_SECURE", int32(keystore.SecurityLevelUnknownSecure)},
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Supplementary attestation info:")
	for _, lvl := range levels {
		info, err := mgr.GetSupplementaryAttestationInfo(lvl.level)
		if err != nil {
			fmt.Fprintf(output, "  %s (%d): %v\n", lvl.name, lvl.level, err)
		} else if info == nil || info.Ref() == 0 {
			fmt.Fprintf(output, "  %s (%d): (null)\n", lvl.name, lvl.level)
		} else {
			fmt.Fprintf(output, "  %s (%d): obtained (ref=%d)\n", lvl.name, lvl.level, info.Ref())
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(info)
				return nil
			})
		}
	}

	// 4. Try GetGrantedKeyFromId with a test ID (will likely fail, but exercises the call).
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Granted key lookups (test IDs):")
	testIDs := []int64{0, 1, -1}
	for _, id := range testIDs {
		key, err := mgr.GetGrantedKeyFromId(id)
		if err != nil {
			fmt.Fprintf(output, "  GetGrantedKeyFromId(%d): %v\n", id, err)
		} else if key == nil || key.Ref() == 0 {
			fmt.Fprintf(output, "  GetGrantedKeyFromId(%d): (null)\n", id)
		} else {
			fmt.Fprintf(output, "  GetGrantedKeyFromId(%d): obtained\n", id)
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(key)
				return nil
			})
		}
	}

	// 5. Try GetGrantedKeyPairFromId.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Granted key pair lookups:")
	for _, id := range testIDs {
		pair, err := mgr.GetGrantedKeyPairFromId(id)
		if err != nil {
			fmt.Fprintf(output, "  GetGrantedKeyPairFromId(%d): %v\n", id, err)
		} else if pair == nil || pair.Ref() == 0 {
			fmt.Fprintf(output, "  GetGrantedKeyPairFromId(%d): (null)\n", id)
		} else {
			fmt.Fprintf(output, "  GetGrantedKeyPairFromId(%d): obtained\n", id)
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(pair)
				return nil
			})
		}
	}

	// 6. Try GetGrantedCertificateChainFromId.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Granted certificate chain lookups:")
	for _, id := range testIDs {
		chain, err := mgr.GetGrantedCertificateChainFromId(id)
		if err != nil {
			fmt.Fprintf(output, "  GetGrantedCertificateChainFromId(%d): %v\n", id, err)
		} else if chain == nil || chain.Ref() == 0 {
			fmt.Fprintf(output, "  GetGrantedCertificateChainFromId(%d): (null)\n", id)
		} else {
			fmt.Fprintf(output, "  GetGrantedCertificateChainFromId(%d): obtained\n", id)
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(chain)
				return nil
			})
		}
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "KeyStore encrypt example complete.")
	return nil
}
