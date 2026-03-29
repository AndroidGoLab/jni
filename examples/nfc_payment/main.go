//go:build android

// Command nfc_payment demonstrates Host Card Emulation (HCE) concepts:
// checks the NFC adapter, shows the card emulation API surface, and lists
// supported tech types and HCE constants.
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
	"github.com/AndroidGoLab/jni/nfc"
	"github.com/AndroidGoLab/jni/nfc/cardemulation"
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

	fmt.Fprintln(output, "=== NFC Payment (HCE) ===")

	// Check NFC adapter.
	adapter := nfc.Adapter{VM: vm}
	adapterObj, err := adapter.GetDefaultAdapter(ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "GetDefaultAdapter: %v\n", err)
		printHCEConstants(output)
		return nil
	}
	if adapterObj == nil || adapterObj.Ref() == 0 {
		fmt.Fprintln(output, "NFC adapter not available (no NFC hardware).")
		printHCEConstants(output)
		return nil
	}

	nfcAdapter := nfc.Adapter{VM: vm, Obj: adapterObj}
	defer func() {
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(adapterObj)
			return nil
		})
	}()

	fmt.Fprintln(output, "NFC adapter: available")

	enabled, err := nfcAdapter.IsEnabled()
	if err != nil {
		fmt.Fprintf(output, "IsEnabled: %v\n", err)
	} else {
		fmt.Fprintf(output, "NFC enabled: %v\n", enabled)
	}

	// Check Secure NFC (important for payments).
	secureSupported, err := nfcAdapter.IsSecureNfcSupported()
	if err != nil {
		fmt.Fprintf(output, "IsSecureNfcSupported: %v\n", err)
	} else {
		fmt.Fprintf(output, "Secure NFC supported: %v\n", secureSupported)
		if secureSupported {
			secureEnabled, err := nfcAdapter.IsSecureNfcEnabled()
			if err != nil {
				fmt.Fprintf(output, "IsSecureNfcEnabled: %v\n", err)
			} else {
				fmt.Fprintf(output, "Secure NFC enabled: %v\n", secureEnabled)
			}
		}
	}

	// Check Observe Mode (used in payment flows).
	observeSupported, err := nfcAdapter.IsObserveModeSupported()
	if err != nil {
		fmt.Fprintf(output, "IsObserveModeSupported: %v\n", err)
	} else {
		fmt.Fprintf(output, "Observe mode supported: %v\n", observeSupported)
	}

	printHCEConstants(output)

	fmt.Fprintln(output, "\n=== HCE Payment Workflow ===")
	fmt.Fprintln(output, "  1. Register HostApduService in manifest")
	fmt.Fprintln(output, "  2. Define AID group for payment category")
	fmt.Fprintln(output, "  3. Process SELECT APDU in processCommandApdu")
	fmt.Fprintln(output, "  4. Exchange APDU commands with terminal")
	fmt.Fprintln(output, "  5. Handle deactivation in onDeactivated")

	fmt.Fprintln(output, "\nNFC payment example completed.")
	return nil
}

func printHCEConstants(output *bytes.Buffer) {
	fmt.Fprintln(output, "\n=== Card Emulation Categories ===")
	fmt.Fprintf(output, "  CATEGORY_PAYMENT: %q\n", cardemulation.CategoryPayment)
	fmt.Fprintf(output, "  CATEGORY_OTHER:   %q\n", cardemulation.CategoryOther)

	fmt.Fprintln(output, "\n=== Selection Modes ===")
	fmt.Fprintf(output, "  PREFER_DEFAULT:  %d\n", cardemulation.SelectionModePreferDefault)
	fmt.Fprintf(output, "  ALWAYS_ASK:      %d\n", cardemulation.SelectionModeAlwaysAsk)
	fmt.Fprintf(output, "  ASK_IF_CONFLICT: %d\n", cardemulation.SelectionModeAskIfConflict)

	fmt.Fprintln(output, "\n=== Deactivation Reasons ===")
	fmt.Fprintf(output, "  LINK_LOSS:  %d\n", cardemulation.DeactivationLinkLoss)
	fmt.Fprintf(output, "  DESELECTED: %d\n", cardemulation.DeactivationDeselected)

	fmt.Fprintln(output, "\n=== Routing Constants ===")
	fmt.Fprintf(output, "  ROUTE_DEFAULT: %d\n", cardemulation.ProtocolAndTechnologyRouteDefault)
	fmt.Fprintf(output, "  ROUTE_ESE:     %d\n", cardemulation.ProtocolAndTechnologyRouteEse)
	fmt.Fprintf(output, "  ROUTE_UICC:    %d\n", cardemulation.ProtocolAndTechnologyRouteUicc)
	fmt.Fprintf(output, "  ROUTE_DH:      %d\n", cardemulation.ProtocolAndTechnologyRouteDh)

	fmt.Fprintln(output, "\n=== Polling Loop Types ===")
	fmt.Fprintf(output, "  TYPE_A:       %d\n", cardemulation.PollingLoopTypeA)
	fmt.Fprintf(output, "  TYPE_B:       %d\n", cardemulation.PollingLoopTypeB)
	fmt.Fprintf(output, "  TYPE_F:       %d\n", cardemulation.PollingLoopTypeF)

	fmt.Fprintln(output, "\n=== Service Configuration ===")
	fmt.Fprintf(output, "  SERVICE_INTERFACE: %q\n", cardemulation.ServiceInterface)
	fmt.Fprintf(output, "  SERVICE_META_DATA: %q\n", cardemulation.ServiceMetaData)
	fmt.Fprintf(output, "  ACTION_CHANGE_DEFAULT: %q\n", cardemulation.ActionChangeDefault)

	fmt.Fprintln(output, "\n=== NFC Adapter States ===")
	fmt.Fprintf(output, "  Off:        %d\n", nfc.StateOff)
	fmt.Fprintf(output, "  TurningOn:  %d\n", nfc.StateTurningOn)
	fmt.Fprintf(output, "  On:         %d\n", nfc.StateOn)
	fmt.Fprintf(output, "  TurningOff: %d\n", nfc.StateTurningOff)
}
