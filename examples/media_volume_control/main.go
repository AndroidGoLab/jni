//go:build android

// Command media_volume_control demonstrates using the AudioManager API to
// read and display all volume streams (music, ring, alarm, notification,
// system, voice call) with their current, minimum, and maximum volumes.
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
		fmt.Fprintf(output, "audiomanager.NewAudioManager: %v\n", err)
		fmt.Fprintln(output, "Volume control example complete (manager unavailable).")
		return nil
	}
	if mgr == nil || mgr.Obj == nil || mgr.Obj.Ref() == 0 {
		fmt.Fprintln(output, "AudioManager: null")
		fmt.Fprintln(output, "Volume control example complete (manager null).")
		return nil
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Volume Control Demo ===")
	ui.RenderOutput()

	// Define all stream types to query.
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

	fmt.Fprintln(output, "Stream volumes:")
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
		minVol, _ := mgr.GetStreamMinVolume(s.constant)
		muted, _ := mgr.IsStreamMute(s.constant)
		fmt.Fprintf(output, "  %s: %d/%d (min=%d muted=%v)\n",
			s.name, vol, maxVol, minVol, muted)
	}
	ui.RenderOutput()

	// Ringer mode.
	fmt.Fprintln(output, "")
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

	// Audio mode.
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

	// Volume fixed flag.
	volFixed, err := mgr.IsVolumeFixed()
	if err != nil {
		fmt.Fprintf(output, "Volume fixed: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "Volume fixed: %v\n", volFixed)
	}
	ui.RenderOutput()

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Volume control example complete.")
	return nil
}
