//go:build android

// Command recorder demonstrates using the MediaRecorder API.
//
// It creates a MediaRecorder via JNI, exercises query methods,
// and prints actual results from the API.
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
	"github.com/AndroidGoLab/jni/media/recorder"
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
	fmt.Fprintln(output, "=== MediaRecorder Demo ===")
	ui.RenderOutput()

	// Create a MediaRecorder via JNI new MediaRecorder().
	var rec recorder.MediaRecorder
	rec.VM = vm
	err := vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("android/media/MediaRecorder")
		if err != nil {
			return fmt.Errorf("find MediaRecorder: %w", err)
		}
		initMid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("get MediaRecorder.<init>: %w", err)
		}
		obj, err := env.NewObject(cls, initMid)
		if err != nil {
			return fmt.Errorf("new MediaRecorder: %w", err)
		}
		rec.Obj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		return fmt.Errorf("create recorder: %w", err)
	}
	fmt.Fprintln(output, "MediaRecorder created OK")
	ui.RenderOutput()

	// Call static method GetAudioSourceMax.
	maxSrc, err := rec.GetAudioSourceMax()
	if err != nil {
		fmt.Fprintf(output, "GetAudioSourceMax: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetAudioSourceMax: %d\n", maxSrc)
	}
	ui.RenderOutput()

	// Query GetPreferredDevice (nil when none set).
	prefDev, err := rec.GetPreferredDevice()
	if err != nil {
		fmt.Fprintf(output, "GetPreferredDevice: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetPreferredDevice: %v\n", prefDev)
	}
	ui.RenderOutput()

	// Query GetRoutedDevice (nil in initial state).
	routedDev, err := rec.GetRoutedDevice()
	if err != nil {
		fmt.Fprintf(output, "GetRoutedDevice: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetRoutedDevice: %v\n", routedDev)
	}
	ui.RenderOutput()

	// Filtered: GetRoutedDevices returns generic type (List<AudioDeviceInfo>)
	// routedDevs, err := rec.GetRoutedDevices()
	// if err != nil {
	// 	fmt.Fprintf(output, "GetRoutedDevices: err=%v\n", err)
	// } else {
	// 	fmt.Fprintf(output, "GetRoutedDevices: %v\n", routedDevs)
	// }
	ui.RenderOutput()

	// Query GetMetrics (returns PersistableBundle).
	metrics, err := rec.GetMetrics()
	if err != nil {
		fmt.Fprintf(output, "GetMetrics: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetMetrics: %v\n", metrics)
	}
	ui.RenderOutput()

	// Filtered: GetActiveMicrophones returns generic type (List<MicrophoneInfo>)
	// mics, err := rec.GetActiveMicrophones()
	// if err != nil {
	// 	fmt.Fprintf(output, "GetActiveMics: err=%v\n", err)
	// } else {
	// 	fmt.Fprintf(output, "GetActiveMics: %v\n", mics)
	// }
	ui.RenderOutput()

	// Query GetActiveRecordingConfiguration.
	recCfg, err := rec.GetActiveRecordingConfiguration()
	if err != nil {
		fmt.Fprintf(output, "GetActiveRecCfg: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetActiveRecCfg: %v\n", recCfg)
	}
	ui.RenderOutput()

	// Query GetLogSessionId.
	logSess, err := rec.GetLogSessionId()
	if err != nil {
		fmt.Fprintf(output, "GetLogSessionId: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetLogSessionId: %v\n", logSess)
	}
	ui.RenderOutput()

	// Exercise SetOrientationHint.
	if err := rec.SetOrientationHint(0); err != nil {
		fmt.Fprintf(output, "SetOrientHint(0): %v\n", err)
	} else {
		fmt.Fprintln(output, "SetOrientHint(0): OK")
	}
	ui.RenderOutput()

	// Exercise Reset (returns recorder to initial state).
	if err := rec.Reset(); err != nil {
		fmt.Fprintf(output, "Reset: %v\n", err)
	} else {
		fmt.Fprintln(output, "Reset: OK")
	}
	ui.RenderOutput()

	// Clean up: release the recorder via JNI.
	vm.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(rec.Obj)
		relMid, err := env.GetMethodID(cls, "release", "()V")
		if err == nil {
			_ = env.CallVoidMethod(rec.Obj, relMid)
		}
		env.DeleteGlobalRef(rec.Obj)
		rec.Obj = nil
		return nil
	})

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Recorder released")
	fmt.Fprintln(output, "Recorder example complete.")
	return nil
}
