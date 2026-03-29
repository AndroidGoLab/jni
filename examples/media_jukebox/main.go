//go:build android

// Command media_jukebox demonstrates using the MediaPlayer API to
// create a player, show play/pause lifecycle, query duration, and
// exercise key lifecycle methods using typed wrappers.
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
	"github.com/AndroidGoLab/jni/media/player"
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
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Media Jukebox Demo ===")
	ui.RenderOutput()

	// Create a MediaPlayer using the typed constructor.
	p, err := player.NewMediaPlayer(vm, ctx.Obj)
	if err != nil {
		return fmt.Errorf("create player: %w", err)
	}
	fmt.Fprintln(output, "MediaPlayer created OK")
	ui.RenderOutput()

	// --- Lifecycle: Idle state ---
	// In idle state (no data source), many queries return defaults.

	// Query duration (returns -1 or error in idle state).
	dur, err := p.GetDuration()
	if err != nil {
		fmt.Fprintf(output, "Duration (idle): err=%v\n", err)
	} else {
		fmt.Fprintf(output, "Duration (idle): %d ms\n", dur)
	}
	ui.RenderOutput()

	// Query current position.
	pos, err := p.GetCurrentPosition()
	if err != nil {
		fmt.Fprintf(output, "Position (idle): err=%v\n", err)
	} else {
		fmt.Fprintf(output, "Position (idle): %d ms\n", pos)
	}
	ui.RenderOutput()

	// Query isPlaying.
	playing, err := p.IsPlaying()
	if err != nil {
		fmt.Fprintf(output, "IsPlaying (idle): err=%v\n", err)
	} else {
		fmt.Fprintf(output, "IsPlaying (idle): %v\n", playing)
	}
	ui.RenderOutput()

	// Query isLooping.
	looping, err := p.IsLooping()
	if err != nil {
		fmt.Fprintf(output, "IsLooping (idle): err=%v\n", err)
	} else {
		fmt.Fprintf(output, "IsLooping (idle): %v\n", looping)
	}
	ui.RenderOutput()

	// Query audio session ID.
	sessionId, err := p.GetAudioSessionId()
	if err != nil {
		fmt.Fprintf(output, "AudioSessionId: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "AudioSessionId: %d\n", sessionId)
	}
	ui.RenderOutput()

	// Query video dimensions (0 in idle).
	vw, err := p.GetVideoWidth()
	if err != nil {
		fmt.Fprintf(output, "VideoWidth: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "VideoWidth: %d\n", vw)
	}
	ui.RenderOutput()

	vh, err := p.GetVideoHeight()
	if err != nil {
		fmt.Fprintf(output, "VideoHeight: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "VideoHeight: %d\n", vh)
	}
	ui.RenderOutput()

	// --- Lifecycle: SetVolume ---
	if err := p.SetVolume(0.8, 0.8); err != nil {
		fmt.Fprintf(output, "SetVolume: %v\n", err)
	} else {
		fmt.Fprintln(output, "SetVolume(0.8, 0.8): OK")
	}
	ui.RenderOutput()

	// --- Lifecycle: SetLooping ---
	if err := p.SetLooping(true); err != nil {
		fmt.Fprintf(output, "SetLooping(true): %v\n", err)
	} else {
		fmt.Fprintln(output, "SetLooping(true): OK")
	}
	ui.RenderOutput()

	looping, err = p.IsLooping()
	if err != nil {
		fmt.Fprintf(output, "IsLooping: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "IsLooping: %v\n", looping)
	}
	ui.RenderOutput()

	// --- Lifecycle: Try setting a data source and preparing ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Setting data source...")
	ui.RenderOutput()

	// Use a content URI for a default notification sound.
	dataSourceURI := "content://settings/system/notification_sound"
	if err := p.SetDataSource1_5(dataSourceURI); err != nil {
		fmt.Fprintf(output, "SetDataSource: %v\n", err)
		fmt.Fprintln(output, "(expected on some devices)")
	} else {
		fmt.Fprintf(output, "DataSource: %s\n", dataSourceURI)

		// Prepare synchronously.
		if err := p.Prepare(); err != nil {
			fmt.Fprintf(output, "Prepare: %v\n", err)
		} else {
			fmt.Fprintln(output, "Prepare: OK")

			dur, err = p.GetDuration()
			if err != nil {
				fmt.Fprintf(output, "Duration: err=%v\n", err)
			} else {
				fmt.Fprintf(output, "Duration: %d ms\n", dur)
			}

			// Start briefly.
			if err := p.Start(); err != nil {
				fmt.Fprintf(output, "Start: %v\n", err)
			} else {
				fmt.Fprintln(output, "Start: OK (playing)")

				playing, _ = p.IsPlaying()
				fmt.Fprintf(output, "IsPlaying: %v\n", playing)

				pos, _ = p.GetCurrentPosition()
				fmt.Fprintf(output, "Position: %d ms\n", pos)

				// Pause.
				if err := p.Pause(); err != nil {
					fmt.Fprintf(output, "Pause: %v\n", err)
				} else {
					fmt.Fprintln(output, "Pause: OK")

					playing, _ = p.IsPlaying()
					fmt.Fprintf(output, "IsPlaying: %v\n", playing)
				}
			}
		}
	}
	ui.RenderOutput()

	// --- Lifecycle: Reset ---
	if err := p.Reset(); err != nil {
		fmt.Fprintf(output, "Reset: %v\n", err)
	} else {
		fmt.Fprintln(output, "Reset: OK (back to idle)")
	}
	ui.RenderOutput()

	// --- Lifecycle: Release ---
	if err := p.Release(); err != nil {
		fmt.Fprintf(output, "Release: %v\n", err)
	} else {
		fmt.Fprintln(output, "Release: OK")
	}
	ui.RenderOutput()

	// Clean up global ref.
	vm.Do(func(env *jni.Env) error {
		if p.Obj != nil {
			env.DeleteGlobalRef(p.Obj)
			p.Obj = nil
		}
		return nil
	})

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Jukebox example complete.")
	return nil
}
