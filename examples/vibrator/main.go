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
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/os/vibrator"
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

	vib, err := vibrator.NewVibrator(ctx)
	if err != nil {
		return fmt.Errorf("vibrator.NewVibrator: %w", err)
	}

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

	// The Vibrator provides Vibrate methods for triggering vibrations
	// and HasVibrator/HasAmplitudeControl for capability checking.

	// Vibrator also provides unexported methods:
	//   vibrateMs(milliseconds int64)              -- vibrate for a duration.
	//   vibratePatternMs(pattern, repeat int32)    -- vibrate with a pattern.
	//     The pattern is a JNI long array of alternating off/on durations.
	//     Set repeat to -1 for one-shot, or the index to repeat from.

	return nil
}
