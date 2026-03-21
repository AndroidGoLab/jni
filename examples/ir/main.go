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
	"strings"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/hardware/ir"
)

func main() {}

func init() { exampleui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	exampleui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	exampleui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	exampleui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := ir.NewConsumerIrManager(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "service not available") {
			fmt.Fprintln(output, "ConsumerIrManager not available on this device")
			fmt.Fprintln(output, "")
			fmt.Fprintln(output, "Package ir provides the following API surface:")
			fmt.Fprintln(output, "  Manager type (wraps android.hardware.ConsumerIrManager)")
			fmt.Fprintln(output, "    - HasIrEmitter() (bool, error)")
			fmt.Fprintln(output, "    - Transmit(carrierFrequency int32, pattern) error")
			fmt.Fprintln(output, "    - getCarrierFrequenciesRaw() (*jni.Object, error)")
			fmt.Fprintln(output, "  frequencyRange data class (CarrierFrequencyRange)")
			fmt.Fprintln(output, "    - MinFrequency int")
			fmt.Fprintln(output, "    - MaxFrequency int")
			return nil
		}
		return fmt.Errorf("ir.NewConsumerIrManager: %w", err)
	}

	// Check whether the device has an IR emitter.
	hasIR, err := mgr.HasIrEmitter()
	if err != nil {
		return fmt.Errorf("HasIrEmitter: %w", err)
	}
	fmt.Fprintf(output, "has IR emitter: %v\n", hasIR)

	if !hasIR {
		fmt.Fprintln(output, "device does not have an IR emitter")
		return nil
	}

	// Transmit an IR signal. The carrier frequency is in Hz and
	// the pattern is an alternating series of on/off durations
	// in microseconds (as a Java int[]).
	var pattern *jni.Object // Java int[] with alternating on/off durations
	carrierFrequency := int32(38000)
	if err := mgr.Transmit(carrierFrequency, pattern); err != nil {
		return fmt.Errorf("Transmit: %w", err)
	}
	fmt.Fprintln(output, "IR signal transmitted at 38 kHz")

	// The frequencyRange data class represents a supported carrier
	// frequency range with MinFrequency and MaxFrequency fields (Hz).
	// It is populated by extracting data from the Java
	// ConsumerIrManager.CarrierFrequencyRange object returned by
	// getCarrierFrequencies().
	fmt.Fprintln(output, "frequencyRange fields: MinFrequency, MaxFrequency")

	return nil
}
