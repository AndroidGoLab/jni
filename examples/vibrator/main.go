//go:build android

// Command vibrator demonstrates using the Android Vibrator system
// service, wrapped by the vibrator package. It is built as a c-shared
// library and packaged into an APK using the shared apk.mk infrastructure.
//
// The vibrator package wraps android.os.Vibrator and provides
// methods to check hardware support, trigger vibrations, and
// cancel ongoing vibrations. It requires the VIBRATE permission.
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
	"github.com/AndroidGoLab/jni/os/vibrator"
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

	vib, err := vibrator.NewVibrator(ctx)
	if err != nil {
		return fmt.Errorf("vibrator.NewVibrator: %w", err)
	}

	// Check if the device has a vibrator.
	hasVib, err := vib.HasVibrator()
	if err != nil {
		return fmt.Errorf("HasVibrator: %w", err)
	}
	fmt.Fprintf(&output, "has vibrator: %v\n", hasVib)

	if !hasVib {
		fmt.Fprintln(&output, "device does not have a vibrator")
		return nil
	}

	// The Vibrator provides Vibrate methods for triggering vibrations
	// and HasVibrator/HasAmplitudeControl for capability checking.

	// Vibrator also provides unexported methods:
	//   vibrateMs(milliseconds int64)              -- vibrate for a duration.
	//   vibratePatternMs(pattern, repeat int32)    -- vibrate with a pattern.
	//     The pattern is a JNI long array of alternating off/on durations.
	//     Set repeat to -1 for one-shot, or the index to repeat from.

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
