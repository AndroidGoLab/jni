//go:build android

// Command wifi_rtt demonstrates using the Android Wi-Fi RTT (Round-Trip
// Time) ranging API, wrapped by the wifi_rtt package. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// The wifi_rtt package wraps android.net.wifi.rtt.WifiRttManager and
// provides the RangingResult data class, status constants, and a
// ranging request builder. Wi-Fi RTT enables precise indoor positioning
// by measuring the round-trip time of Wi-Fi frames. It requires
// ACCESS_FINE_LOCATION and NEARBY_WIFI_DEVICES permissions.
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
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/net/wifi/rtt"
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

	mgr, err := rtt.NewWifiRttManager(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "service not available") {
			fmt.Fprintln(output, "WifiRttManager not available on this device")
			fmt.Fprintln(output, "")
		} else {
			return fmt.Errorf("rtt.NewWifiRttManager: %w", err)
		}
	} else {
		fmt.Fprintln(output, "WifiRttManager obtained successfully")
		_ = mgr
	}

	// --- Ranging Status Constants ---
	fmt.Fprintf(output, "StatusSuccess:                            %d\n", rtt.StatusSuccess)
	fmt.Fprintf(output, "StatusFail:                               %d\n", rtt.StatusFail)
	fmt.Fprintf(output, "StatusResponderDoesNotSupportIeee80211mc: %d\n", rtt.StatusResponderDoesNotSupportIeee80211mc)

	return nil
}
