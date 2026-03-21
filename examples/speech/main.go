//go:build android

// Command speech demonstrates using the Android speech recognition and
// text-to-speech APIs, wrapped by the speech package.
//
// The speech package provides:
//   - Recognizer: wraps android.speech.SpeechRecognizer for speech-to-text.
//   - TTS: wraps android.speech.tts.TextToSpeech for text-to-speech.
//   - Error constants for speech recognition failures.
//   - Queue mode constants for TTS playback.
//   - Callback types for recognition events and TTS progress.
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
	"github.com/AndroidGoLab/jni/speech"
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
	// --- Text-to-Speech ---

	tts, err := speech.NewTTS(vm)
	if err != nil {
		return fmt.Errorf("speech.NewTTS: %w", err)
	}
	defer tts.Close()

	fmt.Fprintln(&output, "TTS engine created")

	// TTS provides unexported methods for speech synthesis:
	//   speakRaw(text, queueMode, params, utteranceId) -- synthesizes speech.
	//   stopRaw()                                      -- stops current speech.
	//   isSpeakingRaw()                                -- checks if TTS is active.
	//   setLanguageRaw(locale)                         -- sets the speech language.
	//   setPitchRaw(pitch)                             -- adjusts voice pitch.
	//   setSpeechRateRaw(rate)                         -- adjusts speech rate.
	//   getAvailableLanguagesRaw()                     -- lists available languages.
	//   shutdownRaw()                                  -- releases TTS resources.
	//   setOnUtteranceProgressListenerRaw(listener)    -- registers a progress listener.

	// --- Speech Recognizer ---

	// Recognizer wraps android.speech.SpeechRecognizer. It has no exported
	// constructor; the unexported createRecognizer(ctx) static method
	// creates the underlying Java object. Other unexported methods:
	//   isRecognitionAvailable(ctx)    -- checks if speech recognition is available.
	//   setRecognitionListener(listener) -- registers a recognition callback.
	//   startListeningRaw(intent)      -- starts listening for speech.
	//   stopListeningRaw()             -- stops the recognizer.
	//   cancelRaw()                    -- cancels ongoing recognition.
	//   destroyRaw()                   -- releases recognizer resources.

	// --- Error Constants ---
	// These correspond to SpeechRecognizer.ERROR_* values:
	fmt.Fprintf(&output, "ErrorNetworkTimeout:          %d\n", speech.ErrorNetworkTimeout)
	fmt.Fprintf(&output, "ErrorNetwork:                 %d\n", speech.ErrorNetwork)
	fmt.Fprintf(&output, "ErrorAudio:                   %d\n", speech.ErrorAudio)
	fmt.Fprintf(&output, "ErrorServer:                  %d\n", speech.ErrorServer)
	fmt.Fprintf(&output, "ErrorClient:                  %d\n", speech.ErrorClient)
	fmt.Fprintf(&output, "ErrorSpeechTimeout:           %d\n", speech.ErrorSpeechTimeout)
	fmt.Fprintf(&output, "ErrorNoMatch:                 %d\n", speech.ErrorNoMatch)
	fmt.Fprintf(&output, "ErrorRecognizerBusy:          %d\n", speech.ErrorRecognizerBusy)
	fmt.Fprintf(&output, "ErrorInsufficientPermissions: %d\n", speech.ErrorInsufficientPermissions)

	// --- TTS Queue Mode Constants ---
	// These correspond to TextToSpeech.QUEUE_* values:
	fmt.Fprintf(&output, "QueueFlush: %d\n", speech.QueueFlush)
	fmt.Fprintf(&output, "QueueAdd:   %d\n", speech.QueueAdd)

	// --- Callback Types (all unexported) ---
	// recognitionListener: OnResults, OnPartialResults, OnError,
	//   OnReadyForSpeech, OnBeginningOfSpeech, OnEndOfSpeech,
	//   OnRmsChanged, OnBufferReceived.
	// ttsOnInitListener: OnInit.
	// utteranceProgressListener: OnDone, OnError, OnStart.

	return nil
}
