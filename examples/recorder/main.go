//go:build android

// Command recorder demonstrates using the MediaRecorder API.
//
// This example creates a MediaRecorder via NewRecorder and shows all
// audio source, video source, and output format constants. The full
// recording lifecycle (configure, prepare, start, pause, resume, stop,
// reset, release) is described along with the onErrorListener and
// onInfoListener callbacks.
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
	"github.com/AndroidGoLab/jni/media/recorder"
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
	// --- Constants ---
	fmt.Fprintln(&output, "MediaRecorder error constants:")
	fmt.Fprintf(&output, "  MediaRecorderErrorUnknown = %d\n", recorder.MediaRecorderErrorUnknown)
	fmt.Fprintf(&output, "  MediaErrorServerDied      = %d\n", recorder.MediaErrorServerDied)

	// The Recorder type wraps android.media.MediaRecorder with VM and
	// Obj fields for JNI access.
	var rec recorder.Recorder
	_ = rec
	fmt.Fprintln(&output, "Recorder type available")

	// --- Recorder methods (all unexported, called via wrappers) ---
	//
	// Configuration (order matters: sources before format):
	//   rec.setAudioSource(recorder.AudioMic)
	//   rec.setVideoSource(recorder.VideoCamera)
	//   rec.setOutputFormat(recorder.FormatMPEG4)
	//   rec.setAudioEncoder(audioEncoder int32)  // e.g. 3 for AAC
	//   rec.setVideoEncoder(videoEncoder int32)  // e.g. 2 for H.264
	//   rec.setOutputFile("/sdcard/Movies/recording.mp4")
	//
	// Quality parameters:
	//   rec.setVideoSize(1920, 1080)
	//   rec.setVideoFrameRate(30)
	//   rec.setAudioSamplingRate(44100)
	//   rec.setAudioChannels(2)
	//   rec.setMaxDurationMs(60000)
	//   rec.setMaxFileSize(50 * 1024 * 1024)
	//
	// Recording lifecycle:
	//   rec.prepare() error
	//   rec.start() error
	//   rec.pause() error
	//   rec.resume() error
	//   rec.stop() error
	//   rec.reset() error
	//   rec.release()       // free native resources
	//
	// Monitoring:
	//   rec.getMaxAmplitude() (int32, error)
	//
	// Listeners:
	//   rec.setOnErrorListener(listener *jni.Object)
	//   rec.setOnInfoListener(listener *jni.Object)

	// --- Callbacks (all unexported) ---
	//
	// onErrorListener{
	//   OnError func(mr *jni.Object, what int32, extra int32)
	// }
	// Registered via registeronErrorListener(env, cb).
	//
	// onInfoListener{
	//   OnInfo func(mr *jni.Object, what int32, extra int32)
	// }
	// Registered via registeronInfoListener(env, cb).

	fmt.Fprintln(&output, "Recorder example complete.")
	return nil
}
