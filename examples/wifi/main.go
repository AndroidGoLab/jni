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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/net/wifi"
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
	fmt.Fprintf(&output, "Wi-Fi enabled: %v\n", enabled)

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
	fmt.Fprintln(&output, "ScanResult and Info types available")

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
