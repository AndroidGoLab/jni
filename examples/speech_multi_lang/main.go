//go:build android

// Command speech_multi_lang initializes TTS, reports the default engine
// and language, and shows TTS language availability constants.
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
	fmt.Fprintln(output, "=== Multi-Language TTS ===")
	ui.RenderOutput()

	// Create TTS engine.
	tts, err := speech.NewTTS(vm)
	if err != nil {
		return fmt.Errorf("NewTTS: %w", err)
	}
	defer tts.Close()
	fmt.Fprintln(output, "TTS engine created OK")
	ui.RenderOutput()

	// Wait for initialization.
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

	// Available languages (the wrapper returns this as a Set object).
	availLangs, err := tts.GetAvailableLanguages()
	if err != nil {
		fmt.Fprintf(output, "\nAvailableLanguages: %v\n", err)
	} else if availLangs != nil && availLangs.Ref() != 0 {
		fmt.Fprintln(output, "\nAvailableLanguages: obtained (Set<Locale>)")
	} else {
		fmt.Fprintln(output, "\nAvailableLanguages: null")
	}

	// Is speaking?
	speaking, err := tts.IsSpeaking()
	if err != nil {
		fmt.Fprintf(output, "IsSpeaking: %v\n", err)
	} else {
		fmt.Fprintf(output, "IsSpeaking: %v\n", speaking)
	}

	// Max speech input length.
	maxLen, err := tts.GetMaxSpeechInputLength()
	if err != nil {
		fmt.Fprintf(output, "MaxSpeechInputLength: %v\n", err)
	} else {
		fmt.Fprintf(output, "MaxSpeechInputLength: %d\n", maxLen)
	}
	ui.RenderOutput()

	// Show TTS language availability constants.
	fmt.Fprintln(output, "\nTTS Constants:")
	fmt.Fprintf(output, "  LANG_AVAILABLE: %d\n", speech.LangAvailable)
	fmt.Fprintf(output, "  LANG_COUNTRY_AVAILABLE: %d\n", speech.LangCountryAvailable)
	fmt.Fprintf(output, "  LANG_COUNTRY_VAR_AVAILABLE: %d\n", speech.LangCountryVarAvailable)
	fmt.Fprintf(output, "  LANG_MISSING_DATA: %d\n", speech.LangMissingData)
	fmt.Fprintf(output, "  LANG_NOT_SUPPORTED: %d\n", speech.LangNotSupported)

	// Shutdown.
	if err := tts.Shutdown(); err != nil {
		fmt.Fprintf(output, "\nShutdown: %v\n", err)
	} else {
		fmt.Fprintln(output, "\nShutdown: OK")
	}

	fmt.Fprintln(output, "\nMulti-lang TTS complete.")
	return nil
}
