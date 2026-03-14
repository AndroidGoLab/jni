//go:build android

// Command telephony demonstrates using the Android TelephonyManager
// system service, wrapped by the telephony package. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// The telephony package wraps android.telephony.TelephonyManager and
// provides methods for querying cellular network information such as
// operator name, phone type, SIM state, roaming status, and data
// connection state. Some methods require READ_PHONE_STATE permission.
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
	"github.com/xaionaro-go/jni/telephony"
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

	mgr, err := telephony.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("telephony.NewManager: %w", err)
	}

	// Manager provides unexported methods for telephony queries:
	//   getNetworkOperatorName() -- human-readable operator name (e.g. "T-Mobile").
	//   getNetworkOperator()     -- MCC+MNC string (e.g. "310260").
	//   getPhoneType()           -- phone radio type (GSM=1, CDMA=2, SIP=3).
	//   getSimState()            -- SIM card state (READY=5, ABSENT=1, etc.).
	//   isNetworkRoaming()       -- whether the device is currently roaming.
	//   getDataState()           -- mobile data connection state.
	//   getDataNetworkType()     -- data network type (LTE=13, NR=20, etc.).
	//
	// These are intended to be wrapped by higher-level helpers.

	fmt.Fprintln(&output, "TelephonyManager obtained successfully")
	fmt.Fprintln(&output, "Unexported methods: getNetworkOperatorName, getNetworkOperator, getPhoneType, getSimState, isNetworkRoaming, getDataState, getDataNetworkType")

	_ = mgr
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
