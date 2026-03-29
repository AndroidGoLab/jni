//go:build android

// Command telephony_dual_sim queries subscription info for dual-SIM
// devices using TelephonyManager.createForSubscriptionId.
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

func simStateName(s int32) string {
	switch int(s) {
	case telephony.SimStateUnknown:
		return "UNKNOWN"
	case telephony.SimStateAbsent:
		return "ABSENT"
	case telephony.SimStateReady:
		return "READY"
	case telephony.SimStateNotReady:
		return "NOT_READY"
	case telephony.SimStatePinRequired:
		return "PIN_REQUIRED"
	case telephony.SimStatePukRequired:
		return "PUK_REQUIRED"
	case telephony.SimStateNetworkLocked:
		return "NETWORK_LOCKED"
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

	fmt.Fprintln(output, "=== Dual SIM Info ===")
	fmt.Fprintln(output)

	phoneCount, err := mgr.GetPhoneCount()
	if err != nil {
		return fmt.Errorf("getPhoneCount: %w", err)
	}
	fmt.Fprintf(output, "phone count: %d\n", phoneCount)

	activeModems, err := mgr.GetActiveModemCount()
	if err != nil {
		fmt.Fprintf(output, "active modems: (unavail)\n")
	} else {
		fmt.Fprintf(output, "active modems: %d\n", activeModems)
	}

	multiSim, err := mgr.IsMultiSimSupported()
	if err != nil {
		fmt.Fprintf(output, "multi-SIM: (unavail)\n")
	} else {
		switch int(multiSim) {
		case telephony.MultisimAllowed:
			fmt.Fprintln(output, "multi-SIM: ALLOWED")
		case telephony.MultisimNotSupportedByHardware:
			fmt.Fprintln(output, "multi-SIM: NOT_SUPPORTED_HW")
		case telephony.MultisimNotSupportedByCarrier:
			fmt.Fprintln(output, "multi-SIM: NOT_SUPPORTED_CARRIER")
		default:
			fmt.Fprintf(output, "multi-SIM: (%d)\n", multiSim)
		}
	}

	fmt.Fprintln(output)

	// Query each SIM slot.
	var slots int32 = phoneCount
	if slots < 1 {
		slots = 1
	}
	for i := int32(0); i < slots; i++ {
		fmt.Fprintf(output, "--- slot %d ---\n", i)

		simState, err := mgr.GetSimState1_1(i)
		if err != nil {
			fmt.Fprintf(output, "  state: (error)\n")
		} else {
			fmt.Fprintf(output, "  state: %s\n", simStateName(simState))
		}

		// Create a subscription-scoped manager for this slot.
		subObj, err := mgr.CreateForSubscriptionId(i)
		if err != nil {
			fmt.Fprintf(output, "  sub manager: (error)\n")
			continue
		}
		if subObj == nil || subObj.Ref() == 0 {
			fmt.Fprintf(output, "  sub manager: (null)\n")
			continue
		}

		subMgr := &telephony.Manager{VM: vm, Ctx: mgr.Ctx, Obj: subObj}

		operName, err := subMgr.GetNetworkOperatorName()
		if err != nil {
			fmt.Fprintf(output, "  operator: (error)\n")
		} else {
			fmt.Fprintf(output, "  operator: %s\n", operName)
		}

		countryIso, err := subMgr.GetNetworkCountryIso0()
		if err != nil {
			fmt.Fprintf(output, "  country: (error)\n")
		} else {
			fmt.Fprintf(output, "  country: %s\n", countryIso)
		}

		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(subObj)
			return nil
		})
	}

	return nil
}
