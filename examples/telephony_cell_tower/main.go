//go:build android

// Command telephony_cell_tower reads the network type and parses
// MCC/MNC from the network operator string. It also displays
// data network type information.
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
	"github.com/AndroidGoLab/jni/telephony"
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

func networkTypeName(t int32) string {
	switch int(t) {
	case telephony.NetworkTypeUnknown:
		return "UNKNOWN"
	case telephony.NetworkTypeGprs:
		return "GPRS"
	case telephony.NetworkTypeEdge:
		return "EDGE"
	case telephony.NetworkTypeUmts:
		return "UMTS"
	case telephony.NetworkTypeCdma:
		return "CDMA"
	case telephony.NetworkTypeEvdo0:
		return "EVDO_0"
	case telephony.NetworkTypeEvdoA:
		return "EVDO_A"
	case telephony.NetworkType1xrtt:
		return "1xRTT"
	case telephony.NetworkTypeHsdpa:
		return "HSDPA"
	case telephony.NetworkTypeHsupa:
		return "HSUPA"
	case telephony.NetworkTypeHspa:
		return "HSPA"
	case telephony.NetworkTypeIden:
		return "iDEN"
	case telephony.NetworkTypeEvdoB:
		return "EVDO_B"
	case telephony.NetworkTypeLte:
		return "LTE"
	case telephony.NetworkTypeEhrpd:
		return "eHRPD"
	case telephony.NetworkTypeHspap:
		return "HSPAP"
	case telephony.NetworkTypeGsm:
		return "GSM"
	case telephony.NetworkTypeTdScdma:
		return "TD_SCDMA"
	case telephony.NetworkTypeIwlan:
		return "IWLAN"
	case telephony.NetworkTypeNr:
		return "NR (5G)"
	default:
		return fmt.Sprintf("(%d)", t)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := telephony.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("telephony.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Cell Tower Info ===")
	fmt.Fprintln(output)

	// Parse MCC and MNC from the network operator string (e.g. "310260").
	operator, err := mgr.GetNetworkOperator()
	if err != nil {
		return fmt.Errorf("getNetworkOperator: %w", err)
	}
	if len(operator) >= 5 {
		mcc := operator[:3]
		mnc := operator[3:]
		fmt.Fprintf(output, "MCC: %s\n", mcc)
		fmt.Fprintf(output, "MNC: %s\n", mnc)
	} else if operator != "" {
		fmt.Fprintf(output, "operator code: %s\n", operator)
	} else {
		fmt.Fprintln(output, "operator code: (empty)")
	}

	operatorName, err := mgr.GetNetworkOperatorName()
	if err != nil {
		return fmt.Errorf("getNetworkOperatorName: %w", err)
	}
	fmt.Fprintf(output, "operator name: %s\n", operatorName)

	countryIso, err := mgr.GetNetworkCountryIso0()
	if err != nil {
		return fmt.Errorf("getNetworkCountryIso: %w", err)
	}
	fmt.Fprintf(output, "country ISO: %s\n", countryIso)

	dataNetType, err := mgr.GetDataNetworkType()
	if err != nil {
		fmt.Fprintf(output, "data network: (unavail)\n")
	} else {
		fmt.Fprintf(output, "data network: %s\n", networkTypeName(dataNetType))
	}

	voiceNetType, err := mgr.GetVoiceNetworkType()
	if err != nil {
		fmt.Fprintf(output, "voice network: (unavail)\n")
	} else {
		fmt.Fprintf(output, "voice network: %s\n", networkTypeName(voiceNetType))
	}

	hasIcc, err := mgr.HasIccCard()
	if err != nil {
		fmt.Fprintf(output, "ICC card: (unavail)\n")
	} else {
		fmt.Fprintf(output, "ICC card: %v\n", hasIcc)
	}

	return nil
}
