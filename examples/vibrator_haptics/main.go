//go:build android

// Command vibrator_haptics demonstrates advanced vibrator haptics.
// It queries vibrator capabilities, triggers different vibration patterns
// (single pulse, repeated pattern), and reports hardware characteristics.
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
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/os/vibrator"
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

	vib, err := vibrator.NewVibrator(ctx)
	if err != nil {
		return fmt.Errorf("vibrator.NewVibrator: %w", err)
	}
	defer vib.Close()

	fmt.Fprintln(output, "=== Vibrator Haptics ===")
	fmt.Fprintln(output, "")

	// Check basic capabilities
	hasVib, err := vib.HasVibrator()
	if err != nil {
		return fmt.Errorf("HasVibrator: %w", err)
	}
	fmt.Fprintf(output, "Has vibrator: %v\n", hasVib)

	if !hasVib {
		fmt.Fprintln(output, "Device does not have a vibrator")
		return nil
	}

	// Amplitude control (API 26+)
	hasAmp, err := vib.HasAmplitudeControl()
	if err != nil {
		fmt.Fprintf(output, "Amplitude control: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Amplitude control: %v\n", hasAmp)
	}

	// Vibrator ID (API 31+)
	vibID, err := vib.GetId()
	if err != nil {
		fmt.Fprintf(output, "Vibrator ID: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Vibrator ID: %d\n", vibID)
	}

	// Resonant frequency (API 34+)
	freq, err := vib.GetResonantFrequency()
	if err != nil {
		fmt.Fprintf(output, "Resonant frequency: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Resonant frequency: %.1f Hz\n", freq)
	}

	// Q factor (API 34+)
	qFactor, err := vib.GetQFactor()
	if err != nil {
		fmt.Fprintf(output, "Q factor: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Q factor: %.2f\n", qFactor)
	}

	// Envelope effects support (API 36+)
	hasEnv, err := vib.AreEnvelopeEffectsSupported()
	if err != nil {
		fmt.Fprintf(output, "Envelope effects: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Envelope effects: %v\n", hasEnv)
	}

	// Pattern 1: Single short pulse (100ms)
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "--- Pattern 1: Short pulse (100ms) ---")
	if err := vib.Vibrate1_3(100); err != nil {
		fmt.Fprintf(output, "Error: %v\n", err)
	} else {
		fmt.Fprintln(output, "Vibrating...")
	}
	time.Sleep(200 * time.Millisecond)

	// Pattern 2: Medium pulse (300ms)
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "--- Pattern 2: Medium pulse (300ms) ---")
	if err := vib.Vibrate1_3(300); err != nil {
		fmt.Fprintf(output, "Error: %v\n", err)
	} else {
		fmt.Fprintln(output, "Vibrating...")
	}
	time.Sleep(400 * time.Millisecond)

	// Pattern 3: Three quick pulses using the deprecated vibrate(long[], int) API
	// Pattern: [delay, vibrate, sleep, vibrate, sleep, vibrate]
	// -1 for repeat means don't repeat
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "--- Pattern 3: Three quick pulses ---")
	pattern := []int64{0, 100, 100, 100, 100, 100} // ms: delay, on, off, on, off, on
	var patternObj *jni.Object
	vm.Do(func(env *jni.Env) error {
		longArr := env.NewLongArray(int32(len(pattern)))
		if longArr == nil {
			return fmt.Errorf("NewLongArray returned nil")
		}
		env.SetLongArrayRegion(longArr, 0, int32(len(pattern)), unsafe.Pointer(&pattern[0]))
		patternObj = (*jni.Object)(unsafe.Pointer(longArr))
		return nil
	})

	if patternObj != nil {
		if err := vib.Vibrate2_5(patternObj, -1); err != nil {
			fmt.Fprintf(output, "Error: %v\n", err)
		} else {
			fmt.Fprintln(output, "Vibrating pattern: [0, 100, 100, 100, 100, 100] ms")
			fmt.Fprintln(output, "  (delay=0, on=100, off=100, on=100, off=100, on=100)")
		}
	}
	time.Sleep(600 * time.Millisecond)

	// Pattern 4: SOS pattern (... --- ...)
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "--- Pattern 4: SOS Morse code ---")
	// S = ... (3 short), O = --- (3 long), S = ... (3 short)
	// short = 100ms, long = 300ms, gap between = 100ms, letter gap = 300ms
	sosPattern := []int64{
		0,
		100, 100, 100, 100, 100, // S: dot dot dot
		300,                      // letter gap
		300, 100, 300, 100, 300,  // O: dash dash dash
		300,                      // letter gap
		100, 100, 100, 100, 100,  // S: dot dot dot
	}
	var sosObj *jni.Object
	vm.Do(func(env *jni.Env) error {
		longArr := env.NewLongArray(int32(len(sosPattern)))
		if longArr == nil {
			return fmt.Errorf("NewLongArray returned nil")
		}
		env.SetLongArrayRegion(longArr, 0, int32(len(sosPattern)), unsafe.Pointer(&sosPattern[0]))
		sosObj = (*jni.Object)(unsafe.Pointer(longArr))
		return nil
	})

	if sosObj != nil {
		if err := vib.Vibrate2_5(sosObj, -1); err != nil {
			fmt.Fprintf(output, "Error: %v\n", err)
		} else {
			fmt.Fprintln(output, "Vibrating SOS: ... --- ...")
		}
	}
	time.Sleep(3 * time.Second)

	// Cancel any ongoing vibration
	if err := vib.Cancel(); err != nil {
		fmt.Fprintf(output, "Cancel error: %v\n", err)
	} else {
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "Vibration cancelled.")
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Vibrator haptics example complete.")
	return nil
}
