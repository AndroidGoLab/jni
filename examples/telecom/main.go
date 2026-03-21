//go:build android

// Command telecom demonstrates using the Android TelecomManager
// system service, wrapped by the telecom package.
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
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/telecom"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
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

	mgr, err := telecom.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("telecom.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== TelecomManager ===")

	dialer, err := mgr.GetDefaultDialerPackage()
	if err != nil {
		fmt.Fprintf(output, "DefaultDialer: %v\n", err)
	} else {
		fmt.Fprintf(output, "Default dialer: %s\n", dialer)
	}

	sysDial, err := mgr.GetSystemDialerPackage()
	if err != nil {
		fmt.Fprintf(output, "SystemDialer: %v\n", err)
	} else {
		fmt.Fprintf(output, "System dialer: %s\n", sysDial)
	}

	inCall, err := mgr.IsInCall()
	if err != nil {
		fmt.Fprintf(output, "IsInCall: %v\n", err)
	} else {
		fmt.Fprintf(output, "In call: %v\n", inCall)
	}

	inManaged, err := mgr.IsInManagedCall()
	if err != nil {
		fmt.Fprintf(output, "IsInManagedCall: %v\n", err)
	} else {
		fmt.Fprintf(output, "In managed call: %v\n", inManaged)
	}

	tty, err := mgr.IsTtySupported()
	if err != nil {
		fmt.Fprintf(output, "IsTtySupported: %v\n", err)
	} else {
		fmt.Fprintf(output, "TTY supported: %v\n", tty)
	}

	// Query call-capable phone accounts (returns java.util.List).
	phoneAccts, err := mgr.GetCallCapablePhoneAccounts()
	if err != nil {
		fmt.Fprintf(output, "PhoneAccounts: %v\n", err)
	} else {
		var listSize int32
		_ = vm.Do(func(env *jni.Env) error {
			if phoneAccts == nil {
				return nil
			}
			listClass, err := env.FindClass("java/util/List")
			if err != nil {
				return err
			}
			sizeMid, err := env.GetMethodID(listClass, "size", "()I")
			if err != nil {
				return err
			}
			listSize, err = env.CallIntMethod(phoneAccts, sizeMid)
			return err
		})
		fmt.Fprintf(output, "Call-capable accounts: %d\n", listSize)
	}

	return nil
}
