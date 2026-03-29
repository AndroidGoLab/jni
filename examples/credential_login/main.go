//go:build android

// Command credential_login demonstrates the Android Credential Manager API.
// It obtains the CredentialManager system service and calls every available
// method: ToString, IsEnabledCredentialProviderService, RegisterCredentialDescription,
// UnregisterCredentialDescription, plus exercises credential type constants.
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

	// 1. Initialize credentials JNI bindings.
	var initErr error
	vm.Do(func(env *jni.Env) error {
		initErr = credentials.Init(env)
		return nil
	})

	if initErr != nil {
		fmt.Fprintf(output, "credentials.Init: %v\n", initErr)
		fmt.Fprintln(output, "Credential Manager NOT available (requires API 34+).")
		return nil
	}
	fmt.Fprintln(output, "credentials.Init: OK")

	// 2. Get the CredentialManager system service.
	mgr, err := credentials.NewCredentialManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "CredentialManager: %v\n", err)
		fmt.Fprintln(output, "(service not available on this device)")
		return nil
	}
	defer mgr.Close()
	fmt.Fprintln(output, "CredentialManager: obtained OK")

	// 3. ToString.
	str, err := mgr.ToString()
	if err != nil {
		fmt.Fprintf(output, "  ToString: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  ToString: %s\n", str)
	}

	// 4. IsEnabledCredentialProviderService - pass nil ComponentName to test.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "IsEnabledCredentialProviderService:")
	enabled, err := mgr.IsEnabledCredentialProviderService(nil)
	if err != nil {
		fmt.Fprintf(output, "  (nil component): error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  (nil component): %v\n", enabled)
	}

	// 5. RegisterCredentialDescription - pass nil to test availability.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "RegisterCredentialDescription:")
	err = mgr.RegisterCredentialDescription(nil)
	if err != nil {
		fmt.Fprintf(output, "  (nil): %v\n", err)
	} else {
		fmt.Fprintln(output, "  (nil): OK")
	}

	// 6. UnregisterCredentialDescription - pass nil to test availability.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "UnregisterCredentialDescription:")
	err = mgr.UnregisterCredentialDescription(nil)
	if err != nil {
		fmt.Fprintf(output, "  (nil): %v\n", err)
	} else {
		fmt.Fprintln(output, "  (nil): OK")
	}

	// 7. Create a Credential object and query its properties.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Credential object test:")
	cred, err := credentials.NewCredential(vm, credentials.TypePasswordCredential, nil)
	if err != nil {
		fmt.Fprintf(output, "  NewCredential: %v\n", err)
	} else {
		credType, err := cred.GetType()
		if err != nil {
			fmt.Fprintf(output, "  GetType: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  GetType: %s\n", credType)
		}

		dc, err := cred.DescribeContents()
		if err != nil {
			fmt.Fprintf(output, "  DescribeContents: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  DescribeContents: %d\n", dc)
		}

		data, err := cred.GetData()
		if err != nil {
			fmt.Fprintf(output, "  GetData: error: %v\n", err)
		} else if data == nil || data.Ref() == 0 {
			fmt.Fprintln(output, "  GetData: (null)")
		} else {
			fmt.Fprintf(output, "  GetData: obtained (ref=%d)\n", data.Ref())
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(data)
				return nil
			})
		}

		credStr, err := cred.ToString()
		if err != nil {
			fmt.Fprintf(output, "  ToString: error: %v\n", err)
		} else {
			fmt.Fprintf(output, "  ToString: %s\n", credStr)
		}
	}

	// 8. Show credential type constants.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Credential type constants:")
	fmt.Fprintf(output, "  TypeUnknown            = %s\n", credentials.TypeUnknown)
	fmt.Fprintf(output, "  TypePasswordCredential = %s\n", credentials.TypePasswordCredential)
	fmt.Fprintf(output, "  TypeInterrupted        = %s\n", credentials.TypeInterrupted)
	fmt.Fprintf(output, "  TypeNoCreateOptions    = %s\n", credentials.TypeNoCreateOptions)
	fmt.Fprintf(output, "  TypeUserCanceled       = %s\n", credentials.TypeUserCanceled)
	fmt.Fprintf(output, "  TypeNoCredential       = %s\n", credentials.TypeNoCredential)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Credential login example complete.")
	return nil
}
