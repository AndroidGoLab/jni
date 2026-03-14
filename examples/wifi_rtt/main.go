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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"strings"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/app"
	"github.com/xaionaro-go/jni/net/wifi/rtt"
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

	mgr, err := rtt.NewManager(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "service not available") {
			fmt.Fprintln(&output, "WifiRttManager not available on this device")
			fmt.Fprintln(&output, "")
		} else {
			return fmt.Errorf("rtt.NewManager: %w", err)
		}
	} else {
		fmt.Fprintln(&output, "WifiRttManager obtained successfully")
		_ = mgr
	}

	// --- Ranging Status Constants ---
	fmt.Fprintf(&output, "StatusSuccess:             %d\n", rtt.StatusSuccess)
	fmt.Fprintf(&output, "StatusFail:                %d\n", rtt.StatusFail)
	fmt.Fprintf(&output, "StatusResponderNotCapable: %d\n", rtt.StatusResponderNotCapable)

	// --- RangingResult Data Class ---
	var result rtt.RangingResult
	fmt.Fprintf(&output, "RangingResult.DistanceMM:       %d\n", result.DistanceMM)
	fmt.Fprintf(&output, "RangingResult.DistanceStdDevMM: %d\n", result.DistanceStdDevMM)
	fmt.Fprintf(&output, "RangingResult.RSSI:             %d\n", result.RSSI)
	fmt.Fprintf(&output, "RangingResult.NumAttempted:      %d\n", result.NumAttempted)
	fmt.Fprintf(&output, "RangingResult.NumSuccessful:     %d\n", result.NumSuccessful)
	fmt.Fprintf(&output, "RangingResult.Status:            %d\n", result.Status)

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
