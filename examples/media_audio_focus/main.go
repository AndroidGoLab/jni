//go:build android

// Command media_audio_focus demonstrates audio focus concepts using the
// AudioManager API. It displays stream volumes, ringer mode, music active
// status, and other audio state information.
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

	fmt.Fprintln(output, "=== Audio Focus Demo ===")
	ui.RenderOutput()

	// --- Music active status ---
	musicActive, err := mgr.IsMusicActive()
	if err != nil {
		fmt.Fprintf(output, "IsMusicActive: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Music active: %v\n", musicActive)
	}
	ui.RenderOutput()

	// --- Ringer mode ---
	ringerMode, err := mgr.GetRingerMode()
	if err != nil {
		fmt.Fprintf(output, "Ringer mode: error: %v\n", err)
	} else {
		name := "unknown"
		switch int(ringerMode) {
		case audiomanager.RingerModeSilent:
			name = "SILENT"
		case audiomanager.RingerModeVibrate:
			name = "VIBRATE"
		case audiomanager.RingerModeNormal:
			name = "NORMAL"
		}
		fmt.Fprintf(output, "Ringer mode: %s (%d)\n", name, ringerMode)
	}
	ui.RenderOutput()

	// --- Audio mode ---
	mode, err := mgr.GetMode()
	if err != nil {
		fmt.Fprintf(output, "Audio mode: error: %v\n", err)
	} else {
		name := "unknown"
		switch int(mode) {
		case audiomanager.ModeNormal:
			name = "NORMAL"
		case audiomanager.ModeRingtone:
			name = "RINGTONE"
		case audiomanager.ModeInCall:
			name = "IN_CALL"
		case audiomanager.ModeInCommunication:
			name = "IN_COMMUNICATION"
		case audiomanager.ModeCallScreening:
			name = "CALL_SCREENING"
		}
		fmt.Fprintf(output, "Audio mode: %s (%d)\n", name, mode)
	}
	ui.RenderOutput()

	// --- Stream volumes ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Stream volumes:")
	type streamInfo struct {
		name     string
		constant int32
	}
	streams := []streamInfo{
		{"VOICE_CALL", int32(audiomanager.StreamVoiceCall)},
		{"SYSTEM", int32(audiomanager.StreamSystem)},
		{"RING", int32(audiomanager.StreamRing)},
		{"MUSIC", int32(audiomanager.StreamMusic)},
		{"ALARM", int32(audiomanager.StreamAlarm)},
		{"NOTIFICATION", int32(audiomanager.StreamNotification)},
	}
	for _, s := range streams {
		vol, err := mgr.GetStreamVolume(s.constant)
		if err != nil {
			fmt.Fprintf(output, "  %s: error: %v\n", s.name, err)
			continue
		}
		maxVol, err := mgr.GetStreamMaxVolume(s.constant)
		if err != nil {
			fmt.Fprintf(output, "  %s: vol=%d max=err\n", s.name, vol)
			continue
		}
		fmt.Fprintf(output, "  %s: %d/%d\n", s.name, vol, maxVol)
	}
	ui.RenderOutput()

	// --- Audio state flags ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Audio state:")
	type boolQuery struct {
		name string
		fn   func() (bool, error)
	}
	queries := []boolQuery{
		{"speakerphone on", mgr.IsSpeakerphoneOn},
		{"mic mute", mgr.IsMicrophoneMute},
		{"bluetooth A2DP on", mgr.IsBluetoothA2dpOn},
		{"bluetooth SCO on", mgr.IsBluetoothScoOn},
		{"wired headset on", mgr.IsWiredHeadsetOn},
		{"volume fixed", mgr.IsVolumeFixed},
	}
	for _, q := range queries {
		val, err := q.fn()
		if err != nil {
			fmt.Fprintf(output, "  %s: error: %v\n", q.name, err)
		} else {
			fmt.Fprintf(output, "  %s: %v\n", q.name, val)
		}
	}
	ui.RenderOutput()

	// --- Audio focus constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Audio focus constants:")
	fmt.Fprintf(output, "  GAIN: %d\n", audiomanager.AudiofocusGain)
	fmt.Fprintf(output, "  GAIN_TRANSIENT: %d\n", audiomanager.AudiofocusGainTransient)
	fmt.Fprintf(output, "  GAIN_TRANSIENT_MAY_DUCK: %d\n", audiomanager.AudiofocusGainTransientMayDuck)
	fmt.Fprintf(output, "  LOSS: %d\n", audiomanager.AudiofocusLoss)
	fmt.Fprintf(output, "  LOSS_TRANSIENT: %d\n", audiomanager.AudiofocusLossTransient)
	fmt.Fprintf(output, "  LOSS_TRANSIENT_CAN_DUCK: %d\n", audiomanager.AudiofocusLossTransientCanDuck)
	fmt.Fprintf(output, "  REQUEST_GRANTED: %d\n", audiomanager.AudiofocusRequestGranted)
	fmt.Fprintf(output, "  REQUEST_FAILED: %d\n", audiomanager.AudiofocusRequestFailed)
	ui.RenderOutput()

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Audio focus example complete.")
	return nil
}
