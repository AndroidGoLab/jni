//go:build android

// Command wifi demonstrates using the Android WifiManager system
// service, wrapped by the wifi package. It is built as a c-shared
// library and packaged into an APK using the shared apk.mk infrastructure.
//
// The wifi package wraps android.net.wifi.WifiManager and provides
// the ScanResult and ConnectionInfo data classes for inspecting
// Wi-Fi networks. It requires ACCESS_FINE_LOCATION and
// ACCESS_WIFI_STATE permissions.
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
	"github.com/AndroidGoLab/jni/net/wifi"
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

	mgr, err := wifi.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("wifi.NewManager: %w", err)
	}
	defer mgr.Close()

	// Check if Wi-Fi is enabled.
	enabled, err := mgr.IsWifiEnabled()
	if err != nil {
		return fmt.Errorf("IsWifiEnabled: %w", err)
	}
	fmt.Fprintf(output, "Wi-Fi enabled: %v\n", enabled)

	// Manager provides methods for Wi-Fi management:
	//   IsWifiEnabled, Is5GHzBandSupported, IsScanAlwaysAvailable,
	//   GetConnectionInfo (returns raw JNI object), GetScanResults, etc.

	// ScanResult and Info are JNI wrapper types with VM and Obj fields.
	// ScanResult wraps android.net.wifi.ScanResult with methods like
	//   DescribeContents, Equals, etc.
	// Info wraps android.net.wifi.WifiInfo with methods like
	//   GetBSSID, GetFrequency, GetSSID, GetRssi, GetLinkSpeed, etc.
	var scan wifi.ScanResult
	_ = scan
	var info wifi.Info
	_ = info
	fmt.Fprintln(output, "ScanResult and Info types available")

	return nil
}
