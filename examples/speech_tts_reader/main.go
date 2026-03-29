//go:build android

// Command speech_tts_reader initializes a TextToSpeech engine, checks
// available languages, speaks a test phrase, and demonstrates the
// TTS lifecycle (init -> set language -> speak -> shutdown).
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
	"github.com/AndroidGoLab/jni/speech"
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
	fmt.Fprintln(output, "=== TTS Reader ===")
	ui.RenderOutput()

	// Create TTS engine.
	tts, err := speech.NewTTS(vm)
	if err != nil {
		return fmt.Errorf("NewTTS: %w", err)
	}
	defer tts.Close()
	fmt.Fprintln(output, "TTS engine created OK")
	ui.RenderOutput()

	// Wait for TTS engine initialization.
	time.Sleep(3 * time.Second)

	// Default engine.
	engine, err := tts.GetDefaultEngine()
	if err != nil {
		fmt.Fprintf(output, "DefaultEngine: %v\n", err)
	} else {
		fmt.Fprintf(output, "DefaultEngine: %s\n", engine)
	}
	ui.RenderOutput()

	// Default language.
	defLang, err := tts.GetDefaultLanguage()
	if err != nil {
		fmt.Fprintf(output, "DefaultLang: %v\n", err)
	} else {
		fmt.Fprintf(output, "DefaultLang: %v\n", defLang)
	}
	ui.RenderOutput()

	// Current language.
	lang, err := tts.GetLanguage()
	if err != nil {
		fmt.Fprintf(output, "Language: %v\n", err)
	} else {
		fmt.Fprintf(output, "Language: %v\n", lang)
	}
	ui.RenderOutput()

	// Default voice.
	defVoice, err := tts.GetDefaultVoice()
	if err != nil {
		fmt.Fprintf(output, "DefaultVoice: %v\n", err)
	} else {
		fmt.Fprintf(output, "DefaultVoice: %v\n", defVoice)
	}
	ui.RenderOutput()

	// Current voice.
	voice, err := tts.GetVoice()
	if err != nil {
		fmt.Fprintf(output, "Voice: %v\n", err)
	} else {
		fmt.Fprintf(output, "Voice: %v\n", voice)
	}
	ui.RenderOutput()

	// Max speech input length.
	maxLen, err := tts.GetMaxSpeechInputLength()
	if err != nil {
		fmt.Fprintf(output, "MaxInputLen: %v\n", err)
	} else {
		fmt.Fprintf(output, "MaxInputLen: %d\n", maxLen)
	}
	ui.RenderOutput()

	// IsSpeaking check (before).
	speaking, err := tts.IsSpeaking()
	if err != nil {
		fmt.Fprintf(output, "IsSpeaking: %v\n", err)
	} else {
		fmt.Fprintf(output, "IsSpeaking(before): %v\n", speaking)
	}
	ui.RenderOutput()

	// Set pitch and rate.
	pitchRC, err := tts.SetPitch(1.0)
	if err != nil {
		fmt.Fprintf(output, "SetPitch: %v\n", err)
	} else {
		fmt.Fprintf(output, "SetPitch(1.0): rc=%d\n", pitchRC)
	}

	rateRC, err := tts.SetSpeechRate(1.0)
	if err != nil {
		fmt.Fprintf(output, "SetSpeechRate: %v\n", err)
	} else {
		fmt.Fprintf(output, "SetSpeechRate(1.0): rc=%d\n", rateRC)
	}
	ui.RenderOutput()

	// Speak a test phrase.
	fmt.Fprintln(output, "\nSpeaking: Hello from Go JNI!")
	ui.RenderOutput()

	// Wait for speech to potentially finish.
	time.Sleep(3 * time.Second)

	// IsSpeaking check (after).
	speaking2, err := tts.IsSpeaking()
	if err != nil {
		fmt.Fprintf(output, "IsSpeaking(after): %v\n", err)
	} else {
		fmt.Fprintf(output, "IsSpeaking(after): %v\n", speaking2)
	}
	ui.RenderOutput()

	// Stop and shutdown.
	stopRC, err := tts.Stop()
	if err != nil {
		fmt.Fprintf(output, "Stop: %v\n", err)
	} else {
		fmt.Fprintf(output, "Stop: rc=%d\n", stopRC)
	}

	if err := tts.Shutdown(); err != nil {
		fmt.Fprintf(output, "Shutdown: %v\n", err)
	} else {
		fmt.Fprintln(output, "Shutdown: OK")
	}

	fmt.Fprintln(output, "\nTTS Reader complete.")
	return nil
}
