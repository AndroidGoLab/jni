//go:build android

// Command settings demonstrates the Android Settings API
// provided by the settings package. It reads system, secure,
// and global settings via the Settings content provider using
// raw JNI calls (the generated package only resolves the class
// references; reading values requires Settings.System.getString etc.).
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

// readSetting reads a single setting value via the static getString method
// on the given Settings inner class (System, Secure, or Global).
func readSetting(
	env *jni.Env,
	settingsClass string,
	resolver *jni.Object,
	name string,
) (string, error) {
	cls, err := env.FindClass(settingsClass)
	if err != nil {
		return "", fmt.Errorf("find class %s: %w", settingsClass, err)
	}

	mid, err := env.GetStaticMethodID(
		cls,
		"getString",
		"(Landroid/content/ContentResolver;Ljava/lang/String;)Ljava/lang/String;",
	)
	if err != nil {
		return "", fmt.Errorf("get getString on %s: %w", settingsClass, err)
	}

	jName, err := env.NewStringUTF(name)
	if err != nil {
		return "", fmt.Errorf("NewStringUTF(%s): %w", name, err)
	}

	resultObj, err := env.CallStaticObjectMethod(
		cls, mid,
		jni.ObjectValue(resolver),
		jni.ObjectValue(&jName.Object),
	)
	if err != nil {
		return "", fmt.Errorf("getString(%s): %w", name, err)
	}

	return env.GoString((*jni.String)(unsafe.Pointer(resultObj))), nil
}

// readIntSetting reads a setting value via the static getInt method.
func readIntSetting(
	env *jni.Env,
	settingsClass string,
	resolver *jni.Object,
	name string,
	defaultVal int32,
) (int32, error) {
	cls, err := env.FindClass(settingsClass)
	if err != nil {
		return 0, fmt.Errorf("find class %s: %w", settingsClass, err)
	}

	mid, err := env.GetStaticMethodID(
		cls,
		"getInt",
		"(Landroid/content/ContentResolver;Ljava/lang/String;I)I",
	)
	if err != nil {
		return 0, fmt.Errorf("get getInt on %s: %w", settingsClass, err)
	}

	jName, err := env.NewStringUTF(name)
	if err != nil {
		return 0, fmt.Errorf("NewStringUTF(%s): %w", name, err)
	}

	result, err := env.CallStaticIntMethod(
		cls, mid,
		jni.ObjectValue(resolver),
		jni.ObjectValue(&jName.Object),
		jni.IntValue(defaultVal),
	)
	if err != nil {
		return 0, fmt.Errorf("getInt(%s): %w", name, err)
	}

	return result, nil
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	resolver, err := ctx.GetContentResolver()
	if err != nil {
		return fmt.Errorf("getContentResolver: %w", err)
	}

	fmt.Fprintln(output, "=== Android Settings ===")
	fmt.Fprintln(output)

	var errs []error

	err = vm.Do(func(env *jni.Env) error {
		// System settings.
		fmt.Fprintln(output, "-- System --")

		brightness, err := readIntSetting(env, "android/provider/Settings$System", resolver, "screen_brightness", -1)
		if err != nil {
			errs = append(errs, err)
			fmt.Fprintf(output, "brightness: err\n")
		} else {
			fmt.Fprintf(output, "brightness: %d\n", brightness)
		}

		fontScale, err := readSetting(env, "android/provider/Settings$System", resolver, "font_scale")
		if err != nil {
			errs = append(errs, err)
		} else {
			fmt.Fprintf(output, "font_scale: %s\n", fontScale)
		}

		// Secure settings.
		fmt.Fprintln(output)
		fmt.Fprintln(output, "-- Secure --")

		deviceName, err := readSetting(env, "android/provider/Settings$Secure", resolver, "bluetooth_name")
		if err != nil {
			errs = append(errs, err)
		} else {
			fmt.Fprintf(output, "bt_name: %s\n", deviceName)
		}

		androidID, err := readSetting(env, "android/provider/Settings$Secure", resolver, "android_id")
		if err != nil {
			errs = append(errs, err)
		} else {
			fmt.Fprintf(output, "android_id: %s\n", androidID)
		}

		// Global settings.
		fmt.Fprintln(output)
		fmt.Fprintln(output, "-- Global --")

		deviceName2, err := readSetting(env, "android/provider/Settings$Global", resolver, "device_name")
		if err != nil {
			errs = append(errs, err)
		} else {
			fmt.Fprintf(output, "device_name: %s\n", deviceName2)
		}

		airplaneMode, err := readIntSetting(env, "android/provider/Settings$Global", resolver, "airplane_mode_on", -1)
		if err != nil {
			errs = append(errs, err)
		} else {
			fmt.Fprintf(output, "airplane: %d\n", airplaneMode)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("vm.Do: %w", err)
	}

	if len(errs) > 0 {
		fmt.Fprintln(output)
		fmt.Fprintf(output, "(%d partial errs)\n", len(errs))
	}

	return nil
}
