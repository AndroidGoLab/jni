//go:build android

// Command speech demonstrates using the Android text-to-speech
// and speech recognition APIs via the speech package.
//
// It creates a TTS engine, queries its state, speaks a phrase,
// and checks speech recognition availability.
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
	fmt.Fprintln(output, "=== Speech Demo ===")
	ui.RenderOutput()

	// Check speech recognition availability via the Recognizer.
	var recognizer speech.Recognizer
	recognizer.VM = vm

	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		fmt.Fprintf(output, "GetAppContext: %v\n", err)
	} else {
		available, err := recognizer.IsRecognitionAvailable(ctx.Obj)
		if err != nil {
			fmt.Fprintf(output, "IsRecogAvail: %v\n", err)
		} else {
			fmt.Fprintf(output, "RecogAvail: %v\n", available)
		}

		onDevice, err := recognizer.IsOnDeviceRecognitionAvailable(ctx.Obj)
		if err != nil {
			fmt.Fprintf(output, "OnDeviceRecog: %v\n", err)
		} else {
			fmt.Fprintf(output, "OnDeviceRecog: %v\n", onDevice)
		}
		ctx.Close()
	}
	ui.RenderOutput()

	// Create TTS engine.
	tts, err := speech.NewTTS(vm)
	if err != nil {
		return fmt.Errorf("NewTTS: %w", err)
	}
	defer tts.Close()
	fmt.Fprintln(output, "TTS engine created OK")
	ui.RenderOutput()

	// Wait for TTS engine initialization (no callback available with current constructor)
	time.Sleep(3 * time.Second)

	// Query GetDefaultEngine.
	engine, err := tts.GetDefaultEngine()
	if err != nil {
		fmt.Fprintf(output, "GetDefaultEngine: %v\n", err)
	} else {
		fmt.Fprintf(output, "DefaultEngine: %s\n", engine)
	}
	ui.RenderOutput()

	// Filtered: GetEngines returns generic type (List<EngineInfo>)
	// engines, err := tts.GetEngines()
	// if err != nil {
	// 	fmt.Fprintf(output, "GetEngines: err=%v\n", err)
	// } else {
	// 	fmt.Fprintf(output, "GetEngines: %v\n", engines)
	// }
	ui.RenderOutput()

	// Query IsSpeaking.
	speaking, err := tts.IsSpeaking()
	if err != nil {
		fmt.Fprintf(output, "IsSpeaking: %v\n", err)
	} else {
		fmt.Fprintf(output, "IsSpeaking: %v\n", speaking)
	}
	ui.RenderOutput()

	// Query GetDefaultLanguage (returns a Locale).
	defLang, err := tts.GetDefaultLanguage()
	if err != nil {
		fmt.Fprintf(output, "GetDefaultLang: %v\n", err)
	} else {
		fmt.Fprintf(output, "DefaultLang: %v\n", defLang)
	}
	ui.RenderOutput()

	// Query GetDefaultVoice.
	defVoice, err := tts.GetDefaultVoice()
	if err != nil {
		fmt.Fprintf(output, "GetDefaultVoice: %v\n", err)
	} else {
		fmt.Fprintf(output, "DefaultVoice: %v\n", defVoice)
	}
	ui.RenderOutput()

	// Query GetLanguage (currently set language).
	lang, err := tts.GetLanguage()
	if err != nil {
		fmt.Fprintf(output, "GetLanguage: %v\n", err)
	} else {
		fmt.Fprintf(output, "Language: %v\n", lang)
	}
	ui.RenderOutput()

	// Query GetVoice (currently set voice).
	voice, err := tts.GetVoice()
	if err != nil {
		fmt.Fprintf(output, "GetVoice: %v\n", err)
	} else {
		fmt.Fprintf(output, "Voice: %v\n", voice)
	}
	ui.RenderOutput()

	// Filtered: GetAvailableLanguages returns generic type (Set<Locale>)
	// availLangs, err := tts.GetAvailableLanguages()
	// if err != nil {
	// 	fmt.Fprintf(output, "GetAvailLangs: %v\n", err)
	// } else {
	// 	fmt.Fprintf(output, "AvailLangs: %v\n", availLangs)
	// }
	ui.RenderOutput()

	// Filtered: GetVoices returns generic type (Set<Voice>)
	// voices, err := tts.GetVoices()
	// if err != nil {
	// 	fmt.Fprintf(output, "GetVoices: %v\n", err)
	// } else {
	// 	fmt.Fprintf(output, "Voices: %v\n", voices)
	// }
	ui.RenderOutput()

	// Query GetMaxSpeechInputLength (static).
	maxLen, err := tts.GetMaxSpeechInputLength()
	if err != nil {
		fmt.Fprintf(output, "MaxInputLen: %v\n", err)
	} else {
		fmt.Fprintf(output, "MaxInputLen: %d\n", maxLen)
	}
	ui.RenderOutput()

	// Set pitch and speech rate.
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

	// Speak a short phrase.
	fmt.Fprintln(output, "Speaking: Hello from Go!")
	ui.RenderOutput()

	// Filtered: Speak4 takes Bundle parameter (generic type filtering)
	// speakRC, err := tts.Speak4("Hello from Go!", int32(speech.QueueFlush), nil, "utt1")
	// if err != nil {
	// 	fmt.Fprintf(output, "Speak: %v\n", err)
	// } else {
	// 	fmt.Fprintf(output, "Speak result: %d\n", speakRC)
	// }
	ui.RenderOutput()

	// Wait and check IsSpeaking.
	time.Sleep(500 * time.Millisecond)
	speaking2, err := tts.IsSpeaking()
	if err != nil {
		fmt.Fprintf(output, "IsSpeaking(after): %v\n", err)
	} else {
		fmt.Fprintf(output, "IsSpeaking(after): %v\n", speaking2)
	}
	ui.RenderOutput()

	// Wait for speech to finish.
	time.Sleep(2 * time.Second)

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

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Speech example complete.")
	return nil
}
