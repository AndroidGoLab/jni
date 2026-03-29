//go:build android

// Command telephony_sim_info reads SIM state, carrier name, phone type,
// network operator, and IMEI (if permitted) via TelephonyManager.
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

func phoneTypeName(t int32) string {
	switch int(t) {
	case telephony.PhoneTypeNone:
		return "NONE"
	case telephony.PhoneTypeGsm:
		return "GSM"
	case telephony.PhoneTypeCdma:
		return "CDMA"
	case telephony.PhoneTypeSip:
		return "SIP"
	default:
		return fmt.Sprintf("(%d)", t)
	}
}

func simStateName(s int32) string {
	switch int(s) {
	case telephony.SimStateUnknown:
		return "UNKNOWN"
	case telephony.SimStateAbsent:
		return "ABSENT"
	case telephony.SimStatePinRequired:
		return "PIN_REQUIRED"
	case telephony.SimStatePukRequired:
		return "PUK_REQUIRED"
	case telephony.SimStateNetworkLocked:
		return "NETWORK_LOCKED"
	case telephony.SimStateReady:
		return "READY"
	case telephony.SimStateNotReady:
		return "NOT_READY"
	case telephony.SimStatePermDisabled:
		return "PERM_DISABLED"
	case telephony.SimStateCardIoError:
		return "CARD_IO_ERROR"
	case telephony.SimStateCardRestricted:
		return "CARD_RESTRICTED"
	default:
		return fmt.Sprintf("(%d)", s)
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

	fmt.Fprintln(output, "=== SIM Info ===")
	fmt.Fprintln(output)

	simState, err := mgr.GetSimState0()
	if err != nil {
		return fmt.Errorf("getSimState: %w", err)
	}
	fmt.Fprintf(output, "SIM state: %s\n", simStateName(simState))

	phoneType, err := mgr.GetPhoneType()
	if err != nil {
		return fmt.Errorf("getPhoneType: %w", err)
	}
	fmt.Fprintf(output, "phone type: %s\n", phoneTypeName(phoneType))

	carrierName, err := mgr.GetSimOperatorName()
	if err != nil {
		return fmt.Errorf("getSimOperatorName: %w", err)
	}
	fmt.Fprintf(output, "carrier: %s\n", carrierName)

	operatorName, err := mgr.GetNetworkOperatorName()
	if err != nil {
		return fmt.Errorf("getNetworkOperatorName: %w", err)
	}
	fmt.Fprintf(output, "operator: %s\n", operatorName)

	operator, err := mgr.GetNetworkOperator()
	if err != nil {
		return fmt.Errorf("getNetworkOperator: %w", err)
	}
	fmt.Fprintf(output, "MCC+MNC: %s\n", operator)

	simCountryIso, err := mgr.GetSimCountryIso()
	if err != nil {
		return fmt.Errorf("getSimCountryIso: %w", err)
	}
	fmt.Fprintf(output, "SIM country: %s\n", simCountryIso)

	// IMEI requires READ_PRIVILEGED_PHONE_STATE on newer APIs;
	// report gracefully if unavailable.
	imei, err := mgr.GetImei0()
	if err != nil {
		fmt.Fprintf(output, "IMEI: (denied)\n")
	} else {
		fmt.Fprintf(output, "IMEI: %s\n", imei)
	}

	return nil
}
