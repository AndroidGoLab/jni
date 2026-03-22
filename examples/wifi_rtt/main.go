//go:build android

// Command wifi_rtt demonstrates using the Android Wi-Fi RTT (Round-Trip
// Time) ranging API. It obtains the WifiRttManager, checks whether RTT
// ranging is available, and reports device RTT characteristics.
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
	"strings"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/net/wifi/rtt"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(activity.vm)),
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
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

	mgr, err := rtt.NewWifiRttManager(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "service not available") {
			fmt.Fprintln(output, "WifiRttManager not available on this device")
			fmt.Fprintln(output, "(Wi-Fi RTT requires hardware support)")
			fmt.Fprintln(output, "")
			printConstants(output)
			return nil
		}
		return fmt.Errorf("WifiRttManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "WifiRttManager OK")

	// Check if RTT ranging is available.
	avail, err := mgr.IsAvailable()
	if err != nil {
		fmt.Fprintf(output, "IsAvailable: %v\n", err)
	} else {
		fmt.Fprintf(output, "RTT available: %v\n", avail)
	}

	// Query RTT characteristics (API 34+).
	chars, err := mgr.GetRttCharacteristics()
	if err != nil {
		fmt.Fprintf(output, "GetRttCharacteristics: %v\n", err)
	} else if chars == nil || chars.Ref() == 0 {
		fmt.Fprintln(output, "No RTT characteristics returned")
	} else {
		fmt.Fprintln(output, "RTT characteristics:")

		// Read boolean keys from the Bundle-like characteristics object.
		vm.Do(func(env *jni.Env) error {
			cls := env.GetObjectClass(chars)

			// Try getBoolean(String key, boolean default).
			getBoolMid, err := env.GetMethodID(cls, "getBoolean", "(Ljava/lang/String;Z)Z")
			if err != nil {
				// Fall back to toString().
				toStrMid, err := env.GetMethodID(cls, "toString", "()Ljava/lang/String;")
				if err != nil {
					return nil
				}
				strObj, err := env.CallObjectMethod(chars, toStrMid)
				if err != nil {
					return nil
				}
				s := env.GoString((*jni.String)(unsafe.Pointer(strObj)))
				fmt.Fprintf(output, "  %s\n", s)
				return nil
			}

			boolKeys := []struct {
				name string
				key  string
			}{
				{"OneSidedRTT", rtt.CharacteristicsKeyBooleanOneSidedRtt},
				{"LCI", rtt.CharacteristicsKeyBooleanLci},
				{"LCR", rtt.CharacteristicsKeyBooleanLcr},
				{"NTB Initiator", rtt.CharacteristicsKeyBooleanNtbInitiator},
				{"STA Responder", rtt.CharacteristicsKeyBooleanStaResponder},
			}
			for _, k := range boolKeys {
				jKey, _ := env.NewStringUTF(k.key)
				val, err := env.CallBooleanMethod(chars, getBoolMid,
					jni.ObjectValue(&jKey.Object), jni.BooleanValue(0))
				if err != nil {
					fmt.Fprintf(output, "  %-15s: err\n", k.name)
					continue
				}
				fmt.Fprintf(output, "  %-15s: %v\n", k.name, val != 0)
			}

			env.DeleteGlobalRef(chars)
			return nil
		})
	}

	fmt.Fprintln(output, "")
	printConstants(output)

	return nil
}

func printConstants(output *bytes.Buffer) {
	fmt.Fprintln(output, "Ranging status constants:")
	fmt.Fprintf(output, "  StatusSuccess: %d\n", rtt.StatusSuccess)
	fmt.Fprintf(output, "  StatusFail:    %d\n", rtt.StatusFail)
	fmt.Fprintf(output, "  StatusNoMC:    %d\n", rtt.StatusResponderDoesNotSupportIeee80211mc)
}
