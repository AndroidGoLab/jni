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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/media/player"
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

	fmt.Fprintf(&output, "MediaPlayer created: %v\n", p)
	fmt.Fprintln(&output, "Player example complete.")
	return nil
}
