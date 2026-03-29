//go:build android

// Command media_audio_record demonstrates using the MediaRecorder API to
// record audio. It configures audio source, output format, encoder, and
// output file, then records for 3 seconds and verifies the file was created.
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
	"os"
	"time"
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

	fmt.Fprintln(output, "=== Audio Record Demo ===")
	ui.RenderOutput()

	// Create a MediaRecorder using the typed constructor.
	rec, err := recorder.NewMediaRecorder(vm, ctx.Obj)
	if err != nil {
		return fmt.Errorf("create recorder: %w", err)
	}
	fmt.Fprintln(output, "MediaRecorder created OK")
	ui.RenderOutput()

	// Output file path in app cache directory.
	outPath := "/sdcard/jni_audio_test.3gp"

	// Configure audio recording.
	if err := rec.SetAudioSource(recorder.AudioSourceMIC); err != nil {
		return fmt.Errorf("SetAudioSource: %w", err)
	}
	fmt.Fprintln(output, "AudioSource: MIC")
	ui.RenderOutput()

	if err := rec.SetOutputFormat(recorder.OutputFormatThreeGPP); err != nil {
		return fmt.Errorf("SetOutputFormat: %w", err)
	}
	fmt.Fprintln(output, "OutputFormat: 3GPP")
	ui.RenderOutput()

	if err := rec.SetAudioEncoder(recorder.AudioEncoderAMRNB); err != nil {
		return fmt.Errorf("SetAudioEncoder: %w", err)
	}
	fmt.Fprintln(output, "AudioEncoder: AMR_NB")
	ui.RenderOutput()

	if err := rec.SetOutputFile1_2(outPath); err != nil {
		return fmt.Errorf("SetOutputFile: %w", err)
	}
	fmt.Fprintf(output, "OutputFile: %s\n", outPath)
	ui.RenderOutput()

	// Prepare the recorder.
	if err := rec.Prepare(); err != nil {
		return fmt.Errorf("Prepare: %w", err)
	}
	fmt.Fprintln(output, "Prepare: OK")
	ui.RenderOutput()

	// Start recording.
	if err := rec.Start(); err != nil {
		return fmt.Errorf("Start: %w", err)
	}
	fmt.Fprintln(output, "Recording started...")
	ui.RenderOutput()

	// Record for 3 seconds, sampling amplitude.
	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		amp, err := rec.GetMaxAmplitude()
		if err != nil {
			fmt.Fprintf(output, "  t=%ds amp=err(%v)\n", i+1, err)
		} else {
			fmt.Fprintf(output, "  t=%ds amp=%d\n", i+1, amp)
		}
		ui.RenderOutput()
	}

	// Stop recording.
	if err := rec.Stop(); err != nil {
		fmt.Fprintf(output, "Stop: %v\n", err)
	} else {
		fmt.Fprintln(output, "Recording stopped")
	}
	ui.RenderOutput()

	// Release the recorder.
	if err := rec.Release(); err != nil {
		fmt.Fprintf(output, "Release: %v\n", err)
	} else {
		fmt.Fprintln(output, "Recorder released")
	}
	ui.RenderOutput()

	// Clean up global ref.
	vm.Do(func(env *jni.Env) error {
		if rec.Obj != nil {
			env.DeleteGlobalRef(rec.Obj)
			rec.Obj = nil
		}
		return nil
	})

	// Verify the output file was created.
	info, err := os.Stat(outPath)
	if err != nil {
		fmt.Fprintf(output, "File check: %v\n", err)
	} else {
		fmt.Fprintf(output, "File created: %s (%d bytes)\n", outPath, info.Size())
	}
	ui.RenderOutput()

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Audio record example complete.")
	return nil
}
