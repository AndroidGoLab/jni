//go:build android

// Command speech_tts_reader initializes a TextToSpeech engine with a proper
// OnInitListener callback, queries engine properties, and demonstrates the
// TTS lifecycle (init -> query -> speak -> shutdown).
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
	"sync"
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

	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// Set the proxy class loader so NewProxy can find APK adapter classes.
	err = vm.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(ctx.Obj)
		mid, err := env.GetMethodID(cls, "getClassLoader", "()Ljava/lang/ClassLoader;")
		if err != nil {
			return fmt.Errorf("get getClassLoader: %w", err)
		}
		cl, err := env.CallObjectMethod(ctx.Obj, mid)
		if err != nil {
			return fmt.Errorf("getClassLoader: %w", err)
		}
		globalCL := env.NewGlobalRef(cl)
		jni.SetProxyClassLoader(globalCL)
		return nil
	})
	if err != nil {
		fmt.Fprintf(output, "SetProxyClassLoader: %v (continuing)\n", err)
	}

	// Create OnInitListener proxy to receive TTS init callback.
	var initStatus int32
	var initOnce sync.Once
	initDone := make(chan struct{})

	var listenerObj *jni.Object
	var listenerCleanup func()
	err = vm.Do(func(env *jni.Env) error {
		listenerCls, err := env.FindClass("android/speech/tts/TextToSpeech$OnInitListener")
		if err != nil {
			return fmt.Errorf("find OnInitListener: %w", err)
		}
		defer env.DeleteLocalRef(&listenerCls.Object)

		proxy, cleanup, err := env.NewProxy(
			[]*jni.Class{listenerCls},
			func(env *jni.Env, method string, args []*jni.Object) (*jni.Object, error) {
				if method == "onInit" && len(args) > 0 {
					// Extract status from the Integer argument.
					intCls, err := env.FindClass("java/lang/Integer")
					if err == nil {
						intValMid, err := env.GetMethodID(intCls, "intValue", "()I")
						if err == nil {
							val, err := env.CallIntMethod(args[0], intValMid)
							if err == nil {
								initStatus = val
							}
						}
					}
					initOnce.Do(func() { close(initDone) })
				}
				return nil, nil
			},
		)
		if err != nil {
			return fmt.Errorf("create proxy: %w", err)
		}
		listenerObj = env.NewGlobalRef(proxy)
		listenerCleanup = cleanup
		return nil
	})
	if err != nil {
		fmt.Fprintf(output, "OnInitListener proxy: %v\n", err)
		fmt.Fprintln(output, "Falling back to TTS constants display.")
		ui.RenderOutput()

		// Show TTS constants even without proxy.
		fmt.Fprintf(output, "\nTTS constants: Success=%d, Error=%d\n", speech.Success, speech.Error)
		fmt.Fprintf(output, "Queue modes: Add=%d, Flush=%d\n", speech.QueueAdd, speech.QueueFlush)
		fmt.Fprintf(output, "Lang results: Available=%d, MissingData=%d, NotSupported=%d\n",
			speech.LangAvailable, speech.LangMissingData, speech.LangNotSupported)
		fmt.Fprintln(output, "\nTTS Reader complete (proxy unavailable).")
		return nil
	}
	defer func() {
		if listenerCleanup != nil {
			listenerCleanup()
		}
		if listenerObj != nil {
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(listenerObj)
				return nil
			})
		}
	}()

	fmt.Fprintln(output, "OnInitListener proxy: created OK")
	ui.RenderOutput()

	// Create TTS with context and listener.
	appCtxObj, err := ctx.GetApplicationContext()
	if err != nil {
		return fmt.Errorf("get app context: %w", err)
	}

	tts, err := speech.NewTextToSpeech(vm, appCtxObj, listenerObj)
	if err != nil {
		return fmt.Errorf("NewTextToSpeech: %w", err)
	}
	defer tts.Close()
	fmt.Fprintln(output, "TTS engine: created OK")
	ui.RenderOutput()

	// Wait for init callback (up to 5 seconds).
	select {
	case <-initDone:
		fmt.Fprintf(output, "TTS init callback: status=%d (Success=%d)\n", initStatus, speech.Success)
	case <-time.After(5 * time.Second):
		fmt.Fprintln(output, "TTS init callback: not received in 5s (continuing)")
	}
	ui.RenderOutput()

	// --- Query engine properties ---
	engine, err := tts.GetDefaultEngine()
	if err != nil {
		fmt.Fprintf(output, "GetDefaultEngine: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Default engine: %s\n", engine)
	}

	maxLen, err := tts.GetMaxSpeechInputLength()
	if err != nil {
		fmt.Fprintf(output, "GetMaxSpeechInputLength: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Max speech input length: %d\n", maxLen)
	}

	speaking, err := tts.IsSpeaking()
	if err != nil {
		fmt.Fprintf(output, "IsSpeaking: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Is speaking: %v\n", speaking)
	}

	defaultsEnforced, err := tts.AreDefaultsEnforced()
	if err != nil {
		fmt.Fprintf(output, "AreDefaultsEnforced: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "Defaults enforced: %v\n", defaultsEnforced)
	}

	enginesObj, err := tts.GetEngines()
	if err != nil {
		fmt.Fprintf(output, "GetEngines: error (%v)\n", err)
	} else if enginesObj == nil {
		fmt.Fprintln(output, "GetEngines: null")
	} else {
		fmt.Fprintln(output, "Engines list: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(enginesObj); return nil })
	}

	availLangs, err := tts.GetAvailableLanguages()
	if err != nil {
		fmt.Fprintf(output, "GetAvailableLanguages: error (%v)\n", err)
	} else if availLangs == nil {
		fmt.Fprintln(output, "GetAvailableLanguages: null")
	} else {
		fmt.Fprintln(output, "Available languages: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(availLangs); return nil })
	}

	defaultLang, err := tts.GetDefaultLanguage()
	if err != nil {
		fmt.Fprintf(output, "GetDefaultLanguage: error (%v)\n", err)
	} else if defaultLang == nil {
		fmt.Fprintln(output, "GetDefaultLanguage: null")
	} else {
		fmt.Fprintln(output, "Default language: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(defaultLang); return nil })
	}

	curLang, err := tts.GetLanguage()
	if err != nil {
		fmt.Fprintf(output, "GetLanguage: error (%v)\n", err)
	} else if curLang == nil {
		fmt.Fprintln(output, "GetLanguage: null")
	} else {
		fmt.Fprintln(output, "Current language: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(curLang); return nil })
	}

	defaultVoice, err := tts.GetDefaultVoice()
	if err != nil {
		fmt.Fprintf(output, "GetDefaultVoice: error (%v)\n", err)
	} else if defaultVoice == nil {
		fmt.Fprintln(output, "GetDefaultVoice: null")
	} else {
		fmt.Fprintln(output, "Default voice: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(defaultVoice); return nil })
	}

	curVoice, err := tts.GetVoice()
	if err != nil {
		fmt.Fprintf(output, "GetVoice: error (%v)\n", err)
	} else if curVoice == nil {
		fmt.Fprintln(output, "GetVoice: null")
	} else {
		fmt.Fprintln(output, "Current voice: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(curVoice); return nil })
	}

	voicesObj, err := tts.GetVoices()
	if err != nil {
		fmt.Fprintf(output, "GetVoices: error (%v)\n", err)
	} else if voicesObj == nil {
		fmt.Fprintln(output, "GetVoices: null")
	} else {
		fmt.Fprintln(output, "Voices set: obtained OK")
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(voicesObj); return nil })
	}
	ui.RenderOutput()

	// --- Set pitch and rate ---
	if rc, err := tts.SetPitch(1.0); err == nil {
		fmt.Fprintf(output, "SetPitch(1.0): rc=%d\n", rc)
	}
	if rc, err := tts.SetSpeechRate(1.0); err == nil {
		fmt.Fprintf(output, "SetSpeechRate(1.0): rc=%d\n", rc)
	}

	// --- Stop and shutdown ---
	if rc, err := tts.Stop(); err == nil {
		fmt.Fprintf(output, "Stop: rc=%d\n", rc)
	}

	ttsStr, err := tts.ToString()
	if err == nil {
		fmt.Fprintf(output, "TTS.ToString: %s\n", ttsStr)
	}

	if err := tts.Shutdown(); err == nil {
		fmt.Fprintln(output, "Shutdown: OK")
	}

	fmt.Fprintln(output, "\nTTS Reader complete.")
	return nil
}
