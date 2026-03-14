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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/app"
	"github.com/xaionaro-go/jni/media/audiomanager"
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
	ctx, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := audiomanager.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("audiomanager.NewManager: %v", err)
	}
	defer mgr.Close()

	// Audio device type constants from android.media.AudioDeviceInfo.
	fmt.Fprintf(&output, "device types: speaker=%d, mic=%d, wired_headset=%d, wired_headphones=%d, bluetooth=%d, bluetooth_le=%d, usb=%d, hdmi=%d\n",
		audiomanager.DeviceBuiltinSpeaker, audiomanager.DeviceBuiltinMic,
		audiomanager.DeviceWiredHeadset, audiomanager.DeviceWiredHeadphones,
		audiomanager.DeviceBluetooth, audiomanager.DeviceBluetoothLE,
		audiomanager.DeviceUSB, audiomanager.DeviceHDMI)

	// Audio focus request constants.
	fmt.Fprintf(&output, "focus gain: gain=%d, transient=%d, transient_duck=%d\n",
		audiomanager.FocusGain, audiomanager.FocusGainTransient,
		audiomanager.FocusGainTransientDuck)
	fmt.Fprintf(&output, "focus loss: loss=%d, transient=%d, transient_duck=%d\n",
		audiomanager.FocusLoss, audiomanager.FocusLossTransient,
		audiomanager.FocusLossTransientDuck)

	// Stream type constants for volume control.
	fmt.Fprintf(&output, "streams: voice_call=%d, system=%d, ring=%d, music=%d, alarm=%d, notification=%d\n",
		audiomanager.StreamVoiceCall, audiomanager.StreamSystem,
		audiomanager.StreamRing, audiomanager.StreamMusic,
		audiomanager.StreamAlarm, audiomanager.StreamNotification)

	// Device filter constants for getDevices.
	fmt.Fprintf(&output, "device filters: input=%d, output=%d, all=%d\n",
		audiomanager.DevicesInput, audiomanager.DevicesOutput, audiomanager.DevicesAll)

	// The Manager also provides unexported methods for:
	//   - getDevicesRaw(flags) - enumerate audio devices
	//   - getStreamVolume(streamType) / setStreamVolume(streamType, index, flags)
	//   - getStreamMaxVolume(streamType)
	//   - isSpeakerphoneOn() / setSpeakerphoneOn(on)
	//   - requestAudioFocusRaw(request) / abandonAudioFocusRequest(request)
	//   - registerAudioDeviceCallback / unregisterAudioDeviceCallback
	fmt.Fprintln(&output, "AudioManager created successfully")

	return nil
}

// getAppContext obtains an Android Context via ActivityThread.currentApplication().
func getAppContext(vm *jni.VM) (*app.Context, error) {
	var ctx app.Context
	ctx.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}

		atClass, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread: %w", err)
		}

		curAppMid, err := env.GetStaticMethodID(atClass, "currentApplication", "()Landroid/app/Application;")
		if err != nil {
			return fmt.Errorf("get currentApplication: %w", err)
		}
		appObj, err := env.CallStaticObjectMethod(atClass, curAppMid)
		if err != nil {
			return fmt.Errorf("call currentApplication: %w", err)
		}
		if appObj == nil || appObj.Ref() == 0 {
			return fmt.Errorf("currentApplication returned null")
		}

		ctx.Obj = env.NewGlobalRef(appObj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}
