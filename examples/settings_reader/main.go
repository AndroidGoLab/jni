//go:build android

// Command settings_reader reads system settings: screen brightness,
// screen timeout, airplane mode, and ringtone using the typed Settings
// wrapper packages (System, Secure, Global).
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
	"github.com/AndroidGoLab/jni/provider/settings"
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

	resolverObj, err := ctx.GetContentResolver()
	if err != nil {
		return fmt.Errorf("getContentResolver: %w", err)
	}

	fmt.Fprintln(output, "=== Settings Reader ===")
	fmt.Fprintln(output)

	sys := settings.System{VM: vm}
	sec := settings.Secure{VM: vm}
	glob := settings.Global{VM: vm}

	// --- System settings ---
	fmt.Fprintln(output, "-- System --")

	brightness, err := sys.GetInt3_1(resolverObj, settings.ScreenBrightness, -1)
	if err != nil {
		fmt.Fprintf(output, "brightness: err (%v)\n", err)
	} else {
		fmt.Fprintf(output, "brightness: %d\n", brightness)
	}

	brightnessMode, err := sys.GetInt3_1(resolverObj, settings.ScreenBrightnessMode, -1)
	if err != nil {
		fmt.Fprintf(output, "brightness_mode: err (%v)\n", err)
	} else {
		modeName := "manual"
		if brightnessMode == 1 {
			modeName = "automatic"
		}
		fmt.Fprintf(output, "brightness_mode: %d (%s)\n", brightnessMode, modeName)
	}

	timeout, err := sys.GetLong3_1(resolverObj, settings.ScreenOffTimeout, -1)
	if err != nil {
		fmt.Fprintf(output, "screen_off_timeout: err (%v)\n", err)
	} else {
		fmt.Fprintf(output, "screen_off_timeout: %dms (%.0fs)\n", timeout, float64(timeout)/1000)
	}

	fontScale, err := sys.GetFloat3_1(resolverObj, settings.FontScale, -1)
	if err != nil {
		fmt.Fprintf(output, "font_scale: err (%v)\n", err)
	} else {
		fmt.Fprintf(output, "font_scale: %.2f\n", fontScale)
	}

	ringtone, err := sys.GetString(resolverObj, settings.Ringtone)
	if err != nil {
		fmt.Fprintf(output, "ringtone: err (%v)\n", err)
	} else {
		fmt.Fprintf(output, "ringtone: %s\n", ringtone)
	}

	// --- Secure settings ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "-- Secure --")

	androidID, err := sec.GetString(resolverObj, settings.AndroidId)
	if err != nil {
		fmt.Fprintf(output, "android_id: err (%v)\n", err)
	} else {
		fmt.Fprintf(output, "android_id: %s\n", androidID)
	}

	defaultIME, err := sec.GetString(resolverObj, settings.DefaultInputMethod)
	if err != nil {
		fmt.Fprintf(output, "default_input_method: err (%v)\n", err)
	} else {
		fmt.Fprintf(output, "default_input_method: %s\n", defaultIME)
	}

	// --- Global settings ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "-- Global --")

	airplane, err := glob.GetInt3_1(resolverObj, settings.AirplaneModeOn, -1)
	if err != nil {
		fmt.Fprintf(output, "airplane_mode: err (%v)\n", err)
	} else {
		state := "off"
		if airplane == 1 {
			state = "on"
		}
		fmt.Fprintf(output, "airplane_mode: %d (%s)\n", airplane, state)
	}

	deviceName, err := glob.GetString(resolverObj, settings.DeviceName)
	if err != nil {
		fmt.Fprintf(output, "device_name: err (%v)\n", err)
	} else {
		fmt.Fprintf(output, "device_name: %s\n", deviceName)
	}

	btOn, err := glob.GetInt3_1(resolverObj, settings.BluetoothOn, -1)
	if err != nil {
		fmt.Fprintf(output, "bluetooth_on: err (%v)\n", err)
	} else {
		fmt.Fprintf(output, "bluetooth_on: %d\n", btOn)
	}

	wifiOn, err := glob.GetInt3_1(resolverObj, settings.WifiOn, -1)
	if err != nil {
		fmt.Fprintf(output, "wifi_on: err (%v)\n", err)
	} else {
		fmt.Fprintf(output, "wifi_on: %d\n", wifiOn)
	}

	return nil
}
