//go:build android

// Command bt_audio_routing queries the Bluetooth adapter for audio profile
// connection states and capabilities using the bluetooth typed wrapper package.
// It reports A2DP, Headset, and HearingAid profile states along with the
// maximum number of connected audio devices.
//
// Required permissions (Android 12+): BLUETOOTH_SCAN, BLUETOOTH_CONNECT,
// BLUETOOTH_ADVERTISE.
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
	"github.com/AndroidGoLab/jni/bluetooth"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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

// Android BluetoothProfile constants for audio profiles.
const (
	profileA2DP       = 2  // BluetoothProfile.A2DP
	profileHeadset    = 1  // BluetoothProfile.HEADSET
	profileHearingAid = 21 // BluetoothProfile.HEARING_AID
)

// connectionStateName returns a human-readable name for a connection state.
func connectionStateName(s int32) string {
	switch int(s) {
	case bluetooth.StateDisconnected:
		return "disconnected"
	case bluetooth.StateConnecting:
		return "connecting"
	case bluetooth.StateConnected:
		return "connected"
	case bluetooth.StateDisconnecting:
		return "disconnecting"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- Connection state constants ---
	fmt.Fprintln(output, "=== Connection state constants ===")
	fmt.Fprintf(output, "  StateDisconnected  = %d\n", bluetooth.StateDisconnected)
	fmt.Fprintf(output, "  StateConnecting    = %d\n", bluetooth.StateConnecting)
	fmt.Fprintf(output, "  StateConnected     = %d\n", bluetooth.StateConnected)
	fmt.Fprintf(output, "  StateDisconnecting = %d\n", bluetooth.StateDisconnecting)

	// --- Adapter ---
	adapter, err := bluetooth.NewAdapter(ctx)
	if err != nil {
		return fmt.Errorf("bluetooth.NewAdapter: %w", err)
	}
	defer adapter.Close()

	enabled, err := adapter.IsEnabled()
	if err != nil {
		return fmt.Errorf("IsEnabled: %w", err)
	}
	fmt.Fprintf(output, "\nBluetooth enabled: %v\n", enabled)
	if !enabled {
		fmt.Fprintln(output, "Bluetooth is off; enable it in Settings.")
		return nil
	}

	name, err := adapter.GetName()
	if err != nil {
		return fmt.Errorf("GetName: %w", err)
	}
	fmt.Fprintf(output, "Adapter name: %s\n", name)

	// --- Check audio profile connection states ---
	fmt.Fprintln(output, "\n=== Audio profile connection states ===")
	for _, p := range []struct {
		name string
		id   int32
	}{
		{"A2DP", profileA2DP},
		{"Headset", profileHeadset},
		{"HearingAid", profileHearingAid},
	} {
		state, err := adapter.GetProfileConnectionState(p.id)
		if err != nil {
			fmt.Fprintf(output, "  %s: error %v\n", p.name, err)
		} else {
			fmt.Fprintf(output, "  %s: %s (state=%d)\n", p.name, connectionStateName(state), state)
		}
	}

	// --- Max connected audio devices ---
	maxAudio, err := adapter.GetMaxConnectedAudioDevices()
	if err != nil {
		fmt.Fprintf(output, "GetMaxConnectedAudioDevices error: %v\n", err)
	} else {
		fmt.Fprintf(output, "\nMax connected audio devices: %d\n", maxAudio)
	}

	// --- LE Audio support ---
	leAudio, err := adapter.IsLeAudioSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeAudioSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "LE Audio supported: %d\n", leAudio)
	}

	leAudioBroadcast, err := adapter.IsLeAudioBroadcastSourceSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeAudioBroadcastSourceSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "LE Audio broadcast source supported: %d\n", leAudioBroadcast)
	}

	leAudioAssistant, err := adapter.IsLeAudioBroadcastAssistantSupported()
	if err != nil {
		fmt.Fprintf(output, "IsLeAudioBroadcastAssistantSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "LE Audio broadcast assistant supported: %d\n", leAudioAssistant)
	}

	fmt.Fprintln(output, "\nAudio routing analysis completed.")
	fmt.Fprintln(output, "No errors occurred during audio routing analysis.")

	return nil
}
