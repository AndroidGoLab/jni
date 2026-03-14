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

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/app"
	"github.com/xaionaro-go/jni/net/wifi"
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
	enabled, err := mgr.IsEnabled()
	if err != nil {
		return fmt.Errorf("IsEnabled: %w", err)
	}
	fmt.Fprintf(&output, "Wi-Fi enabled: %v\n", enabled)

	// Manager also provides unexported methods:
	//   getConnectionInfoRaw() -- returns current Wi-Fi connection info as JNI object.
	//   getScanResultsRaw()    -- returns scan results as a JNI list object.
	// These are intended to be wrapped by higher-level helpers that
	// extract data into the ScanResult and ConnectionInfo structs.

	// --- ScanResult Data Class ---
	// ScanResult holds data from android.net.wifi.ScanResult:
	var scan wifi.ScanResult
	fmt.Fprintf(&output, "ScanResult.SSID:         %q\n", scan.SSID)
	fmt.Fprintf(&output, "ScanResult.BSSID:        %q\n", scan.BSSID)
	fmt.Fprintf(&output, "ScanResult.RSSI:         %d\n", scan.RSSI)
	fmt.Fprintf(&output, "ScanResult.Frequency:    %d\n", scan.Frequency)
	fmt.Fprintf(&output, "ScanResult.Capabilities: %q\n", scan.Capabilities)

	// --- ConnectionInfo Data Class ---
	// ConnectionInfo holds data from android.net.wifi.WifiInfo:
	var info wifi.ConnectionInfo
	fmt.Fprintf(&output, "ConnectionInfo.SSID:      %q\n", info.SSID)
	fmt.Fprintf(&output, "ConnectionInfo.BSSID:     %q\n", info.BSSID)
	fmt.Fprintf(&output, "ConnectionInfo.RSSI:      %d\n", info.RSSI)
	fmt.Fprintf(&output, "ConnectionInfo.LinkSpeed: %d\n", info.LinkSpeed)
	fmt.Fprintf(&output, "ConnectionInfo.Frequency: %d\n", info.Frequency)
	fmt.Fprintf(&output, "ConnectionInfo.IPAddress: %d\n", info.IPAddress)

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
