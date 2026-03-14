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

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/media/recorder"
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
	fmt.Fprintln(&output, "Audio source constants:")
	fmt.Fprintf(&output, "  AudioMic         = %d\n", recorder.AudioMic)
	fmt.Fprintf(&output, "  AudioVoiceCall   = %d\n", recorder.AudioVoiceCall)
	fmt.Fprintf(&output, "  AudioCamcorder   = %d\n", recorder.AudioCamcorder)
	fmt.Fprintf(&output, "  AudioVoiceComm   = %d\n", recorder.AudioVoiceComm)
	fmt.Fprintf(&output, "  AudioUnprocessed = %d\n", recorder.AudioUnprocessed)

	fmt.Fprintln(&output, "Video source constants:")
	fmt.Fprintf(&output, "  VideoCamera  = %d\n", recorder.VideoCamera)
	fmt.Fprintf(&output, "  VideoSurface = %d\n", recorder.VideoSurface)

	fmt.Fprintln(&output, "Output format constants:")
	fmt.Fprintf(&output, "  FormatMPEG4    = %d\n", recorder.FormatMPEG4)
	fmt.Fprintf(&output, "  FormatThreeGPP = %d\n", recorder.FormatThreeGPP)
	fmt.Fprintf(&output, "  FormatWebM     = %d\n", recorder.FormatWebM)
	fmt.Fprintf(&output, "  FormatAAC_ADTS = %d\n", recorder.FormatAAC_ADTS)
	fmt.Fprintf(&output, "  FormatOGG      = %d\n", recorder.FormatOGG)

	// --- NewRecorder ---
	rec, err := recorder.NewRecorder(vm)
	if err != nil {
		return fmt.Errorf("recorder.NewRecorder: %w", err)
	}
	fmt.Fprintf(&output, "MediaRecorder created: %v\n", rec)

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
