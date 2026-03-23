//go:build android

// Command audiomanager demonstrates the Android AudioManager by
// querying live audio state: stream volumes, ringer mode, audio mode,
// and boolean flags such as isMusicActive and isSpeakerphoneOn.
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
	"github.com/AndroidGoLab/jni/media/audiomanager"
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

	mgr, err := audiomanager.NewAudioManager(ctx)
	if err != nil {
		return fmt.Errorf("audiomanager.NewAudioManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== AudioManager ===")

	// Stream volumes.
	type streamInfo struct {
		name     string
		constant int32
	}
	streams := []streamInfo{
		{"voice_call", int32(audiomanager.StreamVoiceCall)},
		{"system", int32(audiomanager.StreamSystem)},
		{"ring", int32(audiomanager.StreamRing)},
		{"music", int32(audiomanager.StreamMusic)},
		{"alarm", int32(audiomanager.StreamAlarm)},
		{"notification", int32(audiomanager.StreamNotification)},
	}
	for _, s := range streams {
		vol, err := mgr.GetStreamVolume(s.constant)
		if err != nil {
			fmt.Fprintf(output, "  %s volume: error: %v\n", s.name, err)
			continue
		}
		maxVol, err := mgr.GetStreamMaxVolume(s.constant)
		if err != nil {
			fmt.Fprintf(output, "  %s volume: %d (max: error: %v)\n", s.name, vol, err)
			continue
		}
		minVol, _ := mgr.GetStreamMinVolume(s.constant)
		muted, _ := mgr.IsStreamMute(s.constant)
		fmt.Fprintf(output, "  %s: vol=%d min=%d max=%d muted=%v\n", s.name, vol, minVol, maxVol, muted)
	}

	// Ringer mode.
	ringerMode, err := mgr.GetRingerMode()
	if err != nil {
		fmt.Fprintf(output, "ringer mode: error: %v\n", err)
	} else {
		name := "unknown"
		switch int(ringerMode) {
		case audiomanager.RingerModeSilent:
			name = "silent"
		case audiomanager.RingerModeVibrate:
			name = "vibrate"
		case audiomanager.RingerModeNormal:
			name = "normal"
		}
		fmt.Fprintf(output, "ringer mode: %s (%d)\n", name, ringerMode)
	}

	// Audio mode.
	mode, err := mgr.GetMode()
	if err != nil {
		fmt.Fprintf(output, "audio mode: error: %v\n", err)
	} else {
		name := "unknown"
		switch int(mode) {
		case audiomanager.ModeNormal:
			name = "normal"
		case audiomanager.ModeRingtone:
			name = "ringtone"
		case audiomanager.ModeInCall:
			name = "in_call"
		case audiomanager.ModeInCommunication:
			name = "in_communication"
		case audiomanager.ModeCallScreening:
			name = "call_screening"
		}
		fmt.Fprintf(output, "audio mode: %s (%d)\n", name, mode)
	}

	// Boolean flags.
	type boolQuery struct {
		name string
		fn   func() (bool, error)
	}
	boolQueries := []boolQuery{
		{"music active", mgr.IsMusicActive},
		{"speakerphone on", mgr.IsSpeakerphoneOn},
		{"mic mute", mgr.IsMicrophoneMute},
		{"bluetooth A2DP on", mgr.IsBluetoothA2dpOn},
		{"bluetooth SCO on", mgr.IsBluetoothScoOn},
		{"wired headset on", mgr.IsWiredHeadsetOn},
		{"volume fixed", mgr.IsVolumeFixed},
	}
	for _, q := range boolQueries {
		val, err := q.fn()
		if err != nil {
			fmt.Fprintf(output, "%s: error: %v\n", q.name, err)
		} else {
			fmt.Fprintf(output, "%s: %v\n", q.name, val)
		}
	}

	return nil
}
