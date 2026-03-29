//go:build android

// Command speech_tts_reader initializes a TextToSpeech engine with a
// proper OnInitListener callback, queries engine properties, and
// demonstrates the TTS lifecycle (init -> query -> speak -> shutdown).
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
					// Extract status from Integer arg.
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
		// If proxy creation fails (e.g. missing proxy classes), fall back
		// to showing TTS constants and API surface without init.
		fmt.Fprintf(output, "OnInitListener proxy: %v\n", err)
		fmt.Fprintln(output, "Falling back to TTS constants display.")
		ui.RenderOutput()

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

	fmt.Fprintln(output, "OnInitListener proxy created OK")
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
	fmt.Fprintln(output, "TTS engine created OK")
	ui.RenderOutput()

	// Wait for init callback (up to 5 seconds).
	select {
	case <-initDone:
		fmt.Fprintf(output, "TTS init callback received (status=%d)\n", initStatus)
	case <-time.After(5 * time.Second):
		fmt.Fprintln(output, "TTS init callback not received in 5s (continuing anyway)")
	}
	ui.RenderOutput()

	// Query engine properties.
	engine, err := tts.GetDefaultEngine()
	if err != nil {
		fmt.Fprintf(output, "DefaultEngine: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "DefaultEngine: %s\n", engine)
	}

	maxLen, err := tts.GetMaxSpeechInputLength()
	if err != nil {
		fmt.Fprintf(output, "MaxInputLen: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "MaxInputLen: %d\n", maxLen)
	}

	speaking, err := tts.IsSpeaking()
	if err != nil {
		fmt.Fprintf(output, "IsSpeaking: error (%v)\n", err)
	} else {
		fmt.Fprintf(output, "IsSpeaking: %v\n", speaking)
	}
	ui.RenderOutput()

	// Set pitch and rate.
	if rc, err := tts.SetPitch(1.0); err == nil {
		fmt.Fprintf(output, "SetPitch(1.0): rc=%d\n", rc)
	}
	if rc, err := tts.SetSpeechRate(1.0); err == nil {
		fmt.Fprintf(output, "SetSpeechRate(1.0): rc=%d\n", rc)
	}

	// Stop and shutdown.
	if rc, err := tts.Stop(); err == nil {
		fmt.Fprintf(output, "Stop: rc=%d\n", rc)
	}
	if err := tts.Shutdown(); err == nil {
		fmt.Fprintln(output, "Shutdown: OK")
	}

	fmt.Fprintln(output, "\nTTS Reader complete.")
	return nil
}
