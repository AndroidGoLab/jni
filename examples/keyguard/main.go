//go:build android

// Command keyguard demonstrates the KeyguardManager JNI bindings. It is built
// as a c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the KeyguardManager system service and queries
// the device lock state. It demonstrates all exported boolean query
// methods and shows how to set up a keyguardDismissCallback for
// receiving keyguard dismiss events.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/os/keyguard"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	vm := jni.VMFromPtr(unsafe.Pointer(cvm))
	if err := run(vm); err != nil {
		fmt.Fprintf(&output, "ERROR: %v\n", err)
	}
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}

func run(vm *jni.VM) error {
	ctx, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := keyguard.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("keyguard.NewManager: %v", err)
	}
	defer mgr.Close()

	// Check whether the keyguard (lock screen) is currently displayed.
	locked, err := mgr.IsKeyguardLocked()
	if err != nil {
		return fmt.Errorf("IsKeyguardLocked: %v", err)
	}
	fmt.Fprintf(&output, "keyguard locked: %v\n", locked)

	// Check whether the keyguard is secured by a PIN, pattern, or password.
	secure, err := mgr.IsKeyguardSecure()
	if err != nil {
		return fmt.Errorf("IsKeyguardSecure: %v", err)
	}
	fmt.Fprintf(&output, "keyguard secure: %v\n", secure)

	// Check whether the device is currently locked (for the current user).
	deviceLocked, err := mgr.IsDeviceLocked()
	if err != nil {
		return fmt.Errorf("IsDeviceLocked: %v", err)
	}
	fmt.Fprintf(&output, "device locked: %v\n", deviceLocked)

	// Check whether the device has a secure lock screen.
	deviceSecure, err := mgr.IsDeviceSecure()
	if err != nil {
		return fmt.Errorf("IsDeviceSecure: %v", err)
	}
	fmt.Fprintf(&output, "device secure: %v\n", deviceSecure)

	// The keyguardDismissCallback is used with requestDismissKeyguard
	// to receive notifications about the keyguard dismiss outcome.
	// It supports three callback functions:
	//   OnDismissSucceeded  - called when the keyguard is dismissed
	//   OnDismissError      - called when an error occurs
	//   OnDismissCancelled  - called when the user cancels the dismiss
	//
	// The keyguardLockedStateListener is used to listen for changes
	// in the keyguard locked state via:
	//   OnLockedStateChanged(isLocked bool)
	fmt.Fprintln(&output, "keyguard state query complete")
	return nil
}

// getAppContext obtains an Android Context via ActivityThread.currentApplication().
func getAppContext(vm *jni.VM) (*app.Context, error) {
	var ctx app.Context
	ctx.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}

		atClass, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread: %w", err)
		}

		curAppMid, err := env.GetStaticMethodID(atClass, "currentApplication", "()Landroid/app/Application;")
		if err != nil {
			return fmt.Errorf("get currentApplication: %w", err)
		}
		appObj, err := env.CallStaticObjectMethod(atClass, curAppMid)
		if err != nil {
			return fmt.Errorf("call currentApplication: %w", err)
		}
		if appObj == nil || appObj.Ref() == 0 {
			return fmt.Errorf("currentApplication returned null")
		}

		ctx.Obj = env.NewGlobalRef(appObj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}
