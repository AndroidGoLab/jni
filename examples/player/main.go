//go:build android

// Command player demonstrates using the MediaPlayer API.
//
// It creates a MediaPlayer via JNI, exercises lifecycle and query methods,
// and prints actual results from the API.
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/media/player"
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
	fmt.Fprintln(output, "=== MediaPlayer Demo ===")
	ui.RenderOutput()

	// Create a MediaPlayer via JNI new MediaPlayer().
	var p player.MediaPlayer
	p.VM = vm
	err := vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("android/media/MediaPlayer")
		if err != nil {
			return fmt.Errorf("find MediaPlayer: %w", err)
		}
		initMid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("get MediaPlayer.<init>: %w", err)
		}
		obj, err := env.NewObject(cls, initMid)
		if err != nil {
			return fmt.Errorf("new MediaPlayer: %w", err)
		}
		p.Obj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		return fmt.Errorf("create player: %w", err)
	}
	fmt.Fprintln(output, "MediaPlayer created OK")
	ui.RenderOutput()

	// Query GetMetrics (returns a PersistableBundle or nil).
	metrics, err := p.GetMetrics()
	if err != nil {
		fmt.Fprintf(output, "GetMetrics: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetMetrics: obj=%v\n", metrics)
	}
	ui.RenderOutput()

	// Query GetDrmInfo (returns nil when no data source is set).
	drmInfo, err := p.GetDrmInfo()
	if err != nil {
		fmt.Fprintf(output, "GetDrmInfo: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetDrmInfo: obj=%v\n", drmInfo)
	}
	ui.RenderOutput()

	// Query GetTimestamp (returns nil in idle state).
	ts, err := p.GetTimestamp()
	if err != nil {
		fmt.Fprintf(output, "GetTimestamp: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetTimestamp: obj=%v\n", ts)
	}
	ui.RenderOutput()

	// Query GetTrackInfo (returns track info array).
	trackInfo, err := p.GetTrackInfo()
	if err != nil {
		fmt.Fprintf(output, "GetTrackInfo: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetTrackInfo: obj=%v\n", trackInfo)
	}
	ui.RenderOutput()

	// Query GetPreferredDevice (nil when none set).
	prefDev, err := p.GetPreferredDevice()
	if err != nil {
		fmt.Fprintf(output, "GetPreferredDevice: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetPreferredDevice: %v\n", prefDev)
	}
	ui.RenderOutput()

	// Query GetRoutedDevice.
	routedDev, err := p.GetRoutedDevice()
	if err != nil {
		fmt.Fprintf(output, "GetRoutedDevice: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetRoutedDevice: %v\n", routedDev)
	}
	ui.RenderOutput()

	// Filtered: GetRoutedDevices returns generic type (List<AudioDeviceInfo>)
	// routedDevs, err := p.GetRoutedDevices()
	// if err != nil {
	// 	fmt.Fprintf(output, "GetRoutedDevices: err=%v\n", err)
	// } else {
	// 	fmt.Fprintf(output, "GetRoutedDevices: %v\n", routedDevs)
	// }
	ui.RenderOutput()

	// Query GetSelectedTrack for audio (track type 2 = MEDIA_TRACK_TYPE_AUDIO).
	selTrack, err := p.GetSelectedTrack(2)
	if err != nil {
		fmt.Fprintf(output, "GetSelectedTrack(audio): err=%v\n", err)
	} else {
		fmt.Fprintf(output, "GetSelectedTrack(audio): %d\n", selTrack)
	}
	ui.RenderOutput()

	// Exercise SetVolume.
	if err := p.SetVolume(0.5, 0.5); err != nil {
		fmt.Fprintf(output, "SetVolume: %v\n", err)
	} else {
		fmt.Fprintln(output, "SetVolume(0.5,0.5): OK")
	}
	ui.RenderOutput()

	// Exercise Reset (returns player to idle state).
	if err := p.Reset(); err != nil {
		fmt.Fprintf(output, "Reset: %v\n", err)
	} else {
		fmt.Fprintln(output, "Reset: OK")
	}
	ui.RenderOutput()

	// Exercise Release (frees native resources).
	if err := p.Release(); err != nil {
		fmt.Fprintf(output, "Release: %v\n", err)
	} else {
		fmt.Fprintln(output, "Release: OK")
	}
	ui.RenderOutput()

	// Delete the global JNI reference.
	vm.Do(func(env *jni.Env) error {
		if p.Obj != nil {
			env.DeleteGlobalRef(p.Obj)
			p.Obj = nil
		}
		return nil
	})

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Player example complete.")
	return nil
}
