//go:build android

// Command vibrator demonstrates the Android Vibrator system service.
// It queries vibrator capabilities and triggers a short vibration.
// Requires the VIBRATE permission in AndroidManifest.xml.
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
	"github.com/AndroidGoLab/jni/os/vibrator"
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

	vib, err := vibrator.NewVibrator(ctx)
	if err != nil {
		return fmt.Errorf("vibrator.NewVibrator: %w", err)
	}
	defer vib.Close()

	fmt.Fprintln(output, "=== Vibrator ===")

	// Check if the device has a vibrator.
	hasVib, err := vib.HasVibrator()
	if err != nil {
		return fmt.Errorf("HasVibrator: %w", err)
	}
	fmt.Fprintf(output, "has vibrator: %v\n", hasVib)

	if !hasVib {
		fmt.Fprintln(output, "device does not have a vibrator")
		return nil
	}

	// Check amplitude control support (API 26+).
	hasAmp, err := vib.HasAmplitudeControl()
	if err != nil {
		fmt.Fprintf(output, "has amplitude control: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "has amplitude control: %v\n", hasAmp)
	}

	// Check envelope effects support (API 36+).
	hasEnvelope, err := vib.AreEnvelopeEffectsSupported()
	if err != nil {
		fmt.Fprintf(output, "envelope effects: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "envelope effects: %v\n", hasEnvelope)
	}

	// Vibrator ID (API 31+).
	vibID, err := vib.GetId()
	if err != nil {
		fmt.Fprintf(output, "vibrator id: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "vibrator id: %d\n", vibID)
	}

	// Resonant frequency (API 34+).
	freq, err := vib.GetResonantFrequency()
	if err != nil {
		fmt.Fprintf(output, "resonant freq: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "resonant freq: %.1f Hz\n", freq)
	}

	// Q factor (API 34+).
	qFactor, err := vib.GetQFactor()
	if err != nil {
		fmt.Fprintf(output, "Q factor: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Q factor: %.2f\n", qFactor)
	}

	// Vibrate for 200ms using the deprecated but simple Vibrate1_3(long).
	fmt.Fprintln(output, "vibrating for 200ms...")
	if err := vib.Vibrate1_3(200); err != nil {
		fmt.Fprintf(output, "vibrate error: %v\n", err)
	} else {
		fmt.Fprintln(output, "vibration triggered!")
	}

	return nil
}
