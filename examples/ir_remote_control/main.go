//go:build android

// Command ir_remote_control demonstrates the ConsumerIrManager API for
// IR blasting. It checks for an IR emitter, reads supported carrier
// frequency ranges, and transmits a sample NEC-protocol TV power toggle
// signal at 38kHz.
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
	"strings"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/hardware/ir"
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

	fmt.Fprintln(output, "=== IR Remote Control ===")
	fmt.Fprintln(output, "")

	mgr, err := ir.NewConsumerIrManager(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "service not available") {
			fmt.Fprintln(output, "ConsumerIrManager not available on this device")
			fmt.Fprintln(output, "")
			fmt.Fprintln(output, "IR remote control requires a device with an IR blaster.")
			fmt.Fprintln(output, "Common devices with IR: some Samsung, Xiaomi, Huawei phones")
			fmt.Fprintln(output, "")
			fmt.Fprintln(output, "API surface available:")
			fmt.Fprintln(output, "  ConsumerIrManager.HasIrEmitter() -> bool")
			fmt.Fprintln(output, "  ConsumerIrManager.GetCarrierFrequencies() -> []CarrierFrequencyRange")
			fmt.Fprintln(output, "  ConsumerIrManager.Transmit(frequency, pattern)")
			fmt.Fprintln(output, "  CarrierFrequencyRange.GetMinFrequency() -> int")
			fmt.Fprintln(output, "  CarrierFrequencyRange.GetMaxFrequency() -> int")
			return nil
		}
		return fmt.Errorf("ir.NewConsumerIrManager: %w", err)
	}

	// Check IR emitter
	hasIR, err := mgr.HasIrEmitter()
	if err != nil {
		return fmt.Errorf("HasIrEmitter: %w", err)
	}
	fmt.Fprintf(output, "Has IR emitter: %v\n", hasIR)

	if !hasIR {
		fmt.Fprintln(output, "Device reports no IR emitter")
		return nil
	}

	// Get carrier frequency ranges
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Supported carrier frequencies:")
	freqArrayObj, err := mgr.GetCarrierFrequencies()
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else if freqArrayObj != nil {
		vm.Do(func(env *jni.Env) error {
			arr := (*jni.ObjectArray)(unsafe.Pointer(freqArrayObj))
			n := env.GetArrayLength((*jni.Array)(unsafe.Pointer(arr)))
			for i := int32(0); i < n; i++ {
				elem, err := env.GetObjectArrayElement(arr, i)
				if err != nil || elem == nil {
					continue
				}
				gRef := env.NewGlobalRef(elem)
				freqRange := ir.ConsumerIrManagerCarrierFrequencyRange{VM: vm, Obj: gRef}
				minFreq, _ := freqRange.GetMinFrequency()
				maxFreq, _ := freqRange.GetMaxFrequency()
				fmt.Fprintf(output, "  Range %d: %d Hz - %d Hz\n", i, minFreq, maxFreq)
				env.DeleteGlobalRef(gRef)
			}
			return nil
		})
	}

	// Transmit a sample NEC TV power toggle at 38kHz
	// NEC protocol: 9ms mark, 4.5ms space, then 32 bits of data
	// Pattern is alternating mark/space durations in microseconds.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Transmitting sample NEC power toggle at 38kHz...")

	// Build the int[] pattern via JNI
	// Samsung TV power: address=0x07, command=0x02
	necPattern := []int32{
		// Leader: 9000us mark, 4500us space
		9000, 4500,
		// Address byte 0x07: 11100000 (LSB first)
		560, 1690, 560, 1690, 560, 1690, 560, 560,
		560, 560, 560, 560, 560, 560, 560, 560,
		// Inverse address 0xF8: 00011111
		560, 560, 560, 560, 560, 560, 560, 1690,
		560, 1690, 560, 1690, 560, 1690, 560, 1690,
		// Command 0x02: 01000000
		560, 560, 560, 1690, 560, 560, 560, 560,
		560, 560, 560, 560, 560, 560, 560, 560,
		// Inverse command 0xFD: 10111111
		560, 1690, 560, 560, 560, 1690, 560, 1690,
		560, 1690, 560, 1690, 560, 1690, 560, 1690,
		// Stop bit
		560,
	}

	var patternObj *jni.Object
	vm.Do(func(env *jni.Env) error {
		intArr := env.NewIntArray(int32(len(necPattern)))
		if intArr == nil {
			return fmt.Errorf("NewIntArray returned nil")
		}
		env.SetIntArrayRegion(intArr, 0, int32(len(necPattern)), unsafe.Pointer(&necPattern[0]))
		patternObj = (*jni.Object)(unsafe.Pointer(intArr))
		return nil
	})

	if patternObj != nil {
		carrierFreq := int32(38000)
		if err := mgr.Transmit(carrierFreq, patternObj); err != nil {
			fmt.Fprintf(output, "Transmit error: %v\n", err)
		} else {
			fmt.Fprintf(output, "Transmitted at %d Hz, pattern length: %d\n", carrierFreq, len(necPattern))
			fmt.Fprintln(output, "IR signal sent!")
		}
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "IR remote control example complete.")
	return nil
}
