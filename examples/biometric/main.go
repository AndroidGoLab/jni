//go:build android

// Command biometric demonstrates Android biometric authentication
// using BiometricManager. Requires API 28+. It is built as a c-shared
// library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// The biometric package wraps android.hardware.biometrics.BiometricManager
// and related types (BiometricPrompt, BiometricPrompt.Builder,
// CancellationSignal). Most prompt-related types and methods are
// unexported and intended to be wrapped by higher-level helpers.
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
	"github.com/AndroidGoLab/jni/hardware/biometric"
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

	mgr, err := biometric.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("biometric.NewManager: %v", err)
	}
	defer mgr.Close()

	// Error code constants from BiometricPrompt.
	fmt.Fprintf(&output, "errors: hw_unavailable=%d, canceled=%d, lockout=%d, lockout_permanent=%d, user_canceled=%d, no_biometrics=%d\n",
		biometric.BiometricErrorHwUnavailable, biometric.BiometricErrorCanceled,
		biometric.BiometricErrorLockout, biometric.BiometricErrorLockoutPermanent,
		biometric.BiometricErrorUserCanceled, biometric.BiometricErrorNoBiometrics)

	// Availability check result constants from BiometricManager.canAuthenticate.
	fmt.Fprintf(&output, "availability: success=%d, no_hardware=%d, hw_unavailable=%d, none_enrolled=%d\n",
		biometric.BiometricSuccess, biometric.BiometricErrorNoHardware,
		biometric.BiometricErrorHwUnavailable, biometric.BiometricErrorNoneEnrolled)

	// The authentication flow uses unexported types:
	//   1. biometricPromptBuilder - set title, subtitle, description,
	//      negative button, and allowed authenticators, then build()
	//   2. biometricPrompt - call authenticate() with a cancellation signal,
	//      executor, and callback
	//   3. cancellationSignal - cancel() to abort authentication
	//
	// The canAuthenticateRaw(authenticators) method on Manager checks
	// whether biometric authentication is available on the device.
	fmt.Fprintln(&output, "BiometricManager created successfully")

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
