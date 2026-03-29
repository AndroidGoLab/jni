//go:build android

// Command credential_login demonstrates the Android Credential Manager API.
// It attempts to obtain the CredentialManager system service, checks
// availability, and shows the credential management API surface.
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
	"github.com/AndroidGoLab/jni/credentials"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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
	fmt.Fprintln(output, "=== Credential Login ===")
	fmt.Fprintln(output)

	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// Initialize credentials JNI bindings.
	var initErr error
	vm.Do(func(env *jni.Env) error {
		initErr = credentials.Init(env)
		return nil
	})

	if initErr != nil {
		fmt.Fprintf(output, "credentials.Init: %v\n", initErr)
		fmt.Fprintln(output, "Credential Manager NOT available.")
		fmt.Fprintln(output, "(Requires API 34+)")
	} else {
		fmt.Fprintln(output, "credentials.Init: OK")
	}

	// Try to get the CredentialManager system service.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "System service check:")

	mgr, err := credentials.NewCredentialManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "  CredentialManager: %v\n", err)
		fmt.Fprintln(output, "  (service not available on this device)")
	} else {
		fmt.Fprintln(output, "  CredentialManager: available")
		defer mgr.Close()

		// Show the object reference.
		fmt.Fprintf(output, "  ref: %d\n", mgr.Obj.Ref())

		// Show available API methods.
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Available methods:")
		fmt.Fprintln(output, "  IsEnabledCredentialProviderService()")
		fmt.Fprintln(output, "  RegisterCredentialDescription()")
		fmt.Fprintln(output, "  UnregisterCredentialDescription()")
	}

	// Show request builder API surface.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Request types:")
	fmt.Fprintln(output, "  GetCredentialRequest.Builder")
	fmt.Fprintln(output, "  CreateCredentialRequest.Builder")
	fmt.Fprintln(output, "  ClearCredentialStateRequest")

	// Show response types.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Response types:")
	fmt.Fprintln(output, "  GetCredentialResponse")
	fmt.Fprintln(output, "  CreateCredentialResponse")
	fmt.Fprintln(output, "  PrepareGetCredentialResponse")

	// Show credential types.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Credential types:")
	fmt.Fprintln(output, "  Credential (base)")
	fmt.Fprintln(output, "  CredentialOption")
	fmt.Fprintln(output, "  CredentialDescription")

	return nil
}
