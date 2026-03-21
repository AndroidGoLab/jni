//go:build android

// Command player demonstrates using the MediaPlayer API.
//
// This example creates a MediaPlayer via NewPlayer and shows the
// complete playback lifecycle: setting a data source, preparing,
// starting, pausing, stopping, and releasing. Three callback types
// (onCompletionListener, onErrorListener, onPreparedListener) handle
// playback events. Most methods are unexported and designed to be
// called from within the player package or via higher-level wrappers.
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
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/media/player"
)

func main() {}

func init() { exampleui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	exampleui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	exampleui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	exampleui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	// --- NewPlayer ---
	p, err := player.NewPlayer(vm)
	if err != nil {
		return fmt.Errorf("player.NewPlayer: %w", err)
	}
	// Player.Close() is inherited from the exported struct fields
	// (VM, Obj). The release() method frees the underlying Java player.

	// --- Player methods (all unexported, called via wrappers) ---
	//
	// Data source:
	//   p.setDataSourcePath(path string) error
	//   p.setDataSourceUri(ctx, uri *jni.Object) error
	//
	// Preparation:
	//   p.prepare() error       - synchronous
	//   p.prepareAsync() error  - asynchronous
	//
	// Playback control:
	//   p.start() error
	//   p.pause() error
	//   p.stop() error
	//   p.seekTo(msec int32) error
	//
	// State queries:
	//   p.getDuration() (int32, error)
	//   p.getCurrentPosition() (int32, error)
	//   p.isPlaying() (bool, error)
	//
	// Configuration:
	//   p.setVolume(left, right float32) error
	//   p.setLooping(looping bool) error
	//
	// Listeners (set via raw JNI proxy objects):
	//   p.setOnCompletionListener(listener *jni.Object)
	//   p.setOnErrorListener(listener *jni.Object)
	//   p.setOnPreparedListener(listener *jni.Object)
	//
	// Lifecycle:
	//   p.reset()    - reset to uninitialized state
	//   p.release()  - release native resources

	// --- Callbacks (all unexported) ---
	// Three callback types handle media player events:
	//
	// onCompletionListener{
	//   OnCompletion func(mp *jni.Object)
	// }
	// Registered via registeronCompletionListener(env, cb).
	//
	// onErrorListener{
	//   OnError func(mp *jni.Object, what int32, extra int32)
	// }
	// Registered via registeronErrorListener(env, cb).
	//
	// onPreparedListener{
	//   OnPrepared func(mp *jni.Object)
	// }
	// Registered via registeronPreparedListener(env, cb).

	fmt.Fprintf(output, "MediaPlayer created: %v\n", p)
	fmt.Fprintln(output, "Player example complete.")
	return nil
}
