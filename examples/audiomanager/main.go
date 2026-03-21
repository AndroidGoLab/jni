//go:build android

// Command audiomanager demonstrates the Android AudioManager for
// controlling audio routing, volume, and device enumeration. It is
// built as a c-shared library and packaged into an APK using the
// shared apk.mk infrastructure.
//
// The audiomanager package wraps android.media.AudioManager.
// Most methods are unexported (getStreamVolume, setStreamVolume, etc.)
// and are intended to be wrapped by higher-level helpers.
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
	"github.com/AndroidGoLab/jni/media/audiomanager"
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
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := audiomanager.NewAudioManager(ctx)
	if err != nil {
		return fmt.Errorf("audiomanager.NewAudioManager: %w", err)
	}
	defer mgr.Close()

	// Audio device type constants from android.media.AudioDeviceInfo.
	fmt.Fprintf(output, "device types: speaker=%d, mic=%d, wired_headset=%d, wired_headphones=%d, bluetooth_a2dp=%d, usb=%d, hdmi=%d\n",
		audiomanager.TypeBuiltinSpeaker, audiomanager.TypeBuiltinMic,
		audiomanager.TypeWiredHeadset, audiomanager.TypeWiredHeadphones,
		audiomanager.TypeBluetoothA2dp,
		audiomanager.TypeUsbDevice, audiomanager.TypeHdmi)

	// Audio focus request constants.
	fmt.Fprintf(output, "focus gain: gain=%d, transient=%d, transient_duck=%d\n",
		audiomanager.AudiofocusGain, audiomanager.AudiofocusGainTransient,
		audiomanager.AudiofocusGainTransientMayDuck)
	fmt.Fprintf(output, "focus loss: loss=%d, transient=%d, transient_duck=%d\n",
		audiomanager.AudiofocusLoss, audiomanager.AudiofocusLossTransient,
		audiomanager.AudiofocusLossTransientCanDuck)

	// Stream type constants for volume control.
	fmt.Fprintf(output, "streams: voice_call=%d, system=%d, ring=%d, music=%d, alarm=%d, notification=%d\n",
		audiomanager.StreamVoiceCall, audiomanager.StreamSystem,
		audiomanager.StreamRing, audiomanager.StreamMusic,
		audiomanager.StreamAlarm, audiomanager.StreamNotification)

	// Device filter constants for getDevices.
	fmt.Fprintf(output, "device filters: input=%d, output=%d, all=%d\n",
		audiomanager.GetDevicesInputs, audiomanager.GetDevicesOutputs, audiomanager.GetDevicesAll)

	// The Manager also provides unexported methods for:
	//   - getDevicesRaw(flags) - enumerate audio devices
	//   - getStreamVolume(streamType) / setStreamVolume(streamType, index, flags)
	//   - getStreamMaxVolume(streamType)
	//   - isSpeakerphoneOn() / setSpeakerphoneOn(on)
	//   - requestAudioFocusRaw(request) / abandonAudioFocusRequest(request)
	//   - registerAudioDeviceCallback / unregisterAudioDeviceCallback
	fmt.Fprintln(output, "AudioManager created successfully")

	return nil
}
