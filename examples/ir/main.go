//go:build android

// Command ir demonstrates the ConsumerIrManager JNI bindings. It is
// built as a c-shared library and packaged into an APK using the
// shared apk.mk infrastructure.
//
// This example obtains the ConsumerIrManager system service, checks
// whether the device has an IR emitter, and transmits an IR signal
// at a specified carrier frequency. The frequencyRange data class
// represents the supported carrier frequency range of the IR emitter.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"strings"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/hardware/ir"
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

	mgr, err := ir.NewManager(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "service not available") {
			fmt.Fprintln(&output, "ConsumerIrManager not available on this device")
			fmt.Fprintln(&output, "")
			fmt.Fprintln(&output, "Package ir provides the following API surface:")
			fmt.Fprintln(&output, "  Manager type (wraps android.hardware.ConsumerIrManager)")
			fmt.Fprintln(&output, "    - HasIrEmitter() (bool, error)")
			fmt.Fprintln(&output, "    - Transmit(carrierFrequency int32, pattern) error")
			fmt.Fprintln(&output, "    - getCarrierFrequenciesRaw() (*jni.Object, error)")
			fmt.Fprintln(&output, "  frequencyRange data class (CarrierFrequencyRange)")
			fmt.Fprintln(&output, "    - MinFrequency int")
			fmt.Fprintln(&output, "    - MaxFrequency int")
			return nil
		}
		return fmt.Errorf("ir.NewManager: %v", err)
	}

	// Check whether the device has an IR emitter.
	hasIR, err := mgr.HasIrEmitter()
	if err != nil {
		return fmt.Errorf("HasIrEmitter: %v", err)
	}
	fmt.Fprintf(&output, "has IR emitter: %v\n", hasIR)

	if !hasIR {
		fmt.Fprintln(&output, "device does not have an IR emitter")
		return nil
	}

	// Transmit an IR signal. The carrier frequency is in Hz and
	// the pattern is an alternating series of on/off durations
	// in microseconds (as a Java int[]).
	var pattern *jni.Object // Java int[] with alternating on/off durations
	carrierFrequency := int32(38000)
	if err := mgr.Transmit(carrierFrequency, pattern); err != nil {
		return fmt.Errorf("Transmit: %v", err)
	}
	fmt.Fprintln(&output, "IR signal transmitted at 38 kHz")

	// The frequencyRange data class represents a supported carrier
	// frequency range with MinFrequency and MaxFrequency fields (Hz).
	// It is populated by extracting data from the Java
	// ConsumerIrManager.CarrierFrequencyRange object returned by
	// getCarrierFrequencies().
	fmt.Fprintln(&output, "frequencyRange fields: MinFrequency, MaxFrequency")

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
