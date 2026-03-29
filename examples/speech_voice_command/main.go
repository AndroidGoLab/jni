//go:build android

// Command speech_voice_command checks if speech recognition is available
// on the device and shows the SpeechRecognizer API surface.
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
	fmt.Fprintln(output, "=== Voice Command ===")

	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("GetAppContext: %w", err)
	}
	defer ctx.Close()

	var recognizer speech.Recognizer
	recognizer.VM = vm

	// Check recognition availability.
	available, err := recognizer.IsRecognitionAvailable(ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "IsRecognitionAvailable: %v\n", err)
	} else {
		fmt.Fprintf(output, "RecognitionAvailable: %v\n", available)
	}

	// Check on-device recognition.
	onDevice, err := recognizer.IsOnDeviceRecognitionAvailable(ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "OnDeviceRecognition: %v\n", err)
	} else {
		fmt.Fprintf(output, "OnDeviceRecognition: %v\n", onDevice)
	}

	// Show RecognizerIntent constants used for building
	// speech recognition intents.
	fmt.Fprintln(output, "\nRecognizerIntent Constants:")
	fmt.Fprintf(output, "  ACTION_RECOGNIZE_SPEECH: %s\n", speech.ActionRecognizeSpeech)
	fmt.Fprintf(output, "  ACTION_WEB_SEARCH: %s\n", speech.ActionWebSearch)
	fmt.Fprintf(output, "  ACTION_VOICE_SEARCH_HANDS_FREE: %s\n", speech.ActionVoiceSearchHandsFree)
	fmt.Fprintf(output, "  EXTRA_LANGUAGE_MODEL: %s\n", speech.ExtraLanguageModel)
	fmt.Fprintf(output, "  EXTRA_LANGUAGE: %s\n", speech.ExtraLanguage)
	fmt.Fprintf(output, "  EXTRA_PROMPT: %s\n", speech.ExtraPrompt)
	fmt.Fprintf(output, "  EXTRA_MAX_RESULTS: %s\n", speech.ExtraMaxResults)
	fmt.Fprintf(output, "  EXTRA_PARTIAL_RESULTS: %s\n", speech.ExtraPartialResults)
	fmt.Fprintf(output, "  EXTRA_PREFER_OFFLINE: %s\n", speech.ExtraPreferOffline)
	fmt.Fprintf(output, "  LANGUAGE_MODEL_FREE_FORM: %s\n", speech.LanguageModelFreeForm)
	fmt.Fprintf(output, "  LANGUAGE_MODEL_WEB_SEARCH: %s\n", speech.LanguageModelWebSearch)

	// Show error code constants.
	fmt.Fprintln(output, "\nSpeechRecognizer Error Codes:")
	fmt.Fprintf(output, "  ERROR_AUDIO: %d\n", speech.ErrorAudio)
	fmt.Fprintf(output, "  ERROR_CLIENT: %d\n", speech.ErrorClient)
	fmt.Fprintf(output, "  ERROR_NETWORK: %d\n", speech.ErrorNetwork)
	fmt.Fprintf(output, "  ERROR_NO_MATCH: %d\n", speech.ErrorNoMatch)
	fmt.Fprintf(output, "  ERROR_SERVER: %d\n", speech.ErrorServer)
	fmt.Fprintf(output, "  ERROR_SPEECH_TIMEOUT: %d\n", speech.ErrorSpeechTimeout)
	fmt.Fprintf(output, "  ERROR_INSUFFICIENT_PERMISSIONS: %d\n", speech.ErrorInsufficientPermissions)

	fmt.Fprintln(output, "\nVoice command complete.")
	return nil
}
