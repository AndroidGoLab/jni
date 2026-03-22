//go:build android

// Command nfc demonstrates using the NFC adapter API. It obtains the
// default NfcAdapter, checks whether NFC is enabled, and queries
// various adapter capabilities.
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
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/nfc"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
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

	// Obtain the default NFC adapter via the static method
	// NfcAdapter.getDefaultAdapter(Context).
	adapter := nfc.Adapter{VM: vm}
	adapterObj, err := adapter.GetDefaultAdapter(ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "GetDefaultAdapter: %v\n", err)
		fmt.Fprintln(output, "NFC may not be available")
		printConstants(output)
		return nil
	}
	if adapterObj == nil || adapterObj.Ref() == 0 {
		fmt.Fprintln(output, "NFC adapter not available")
		fmt.Fprintln(output, "(device has no NFC hardware)")
		printConstants(output)
		return nil
	}

	// Wrap the returned object in an Adapter struct.
	nfcAdapter := nfc.Adapter{VM: vm, Obj: adapterObj}
	fmt.Fprintln(output, "NFC adapter obtained")

	// Check if NFC is enabled.
	enabled, err := nfcAdapter.IsEnabled()
	if err != nil {
		fmt.Fprintf(output, "IsEnabled: %v\n", err)
	} else {
		fmt.Fprintf(output, "NFC enabled: %v\n", enabled)
	}

	// Check Secure NFC support and state.
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

	// Check Observe Mode support (API 35+).
	observeSupported, err := nfcAdapter.IsObserveModeSupported()
	if err != nil {
		fmt.Fprintf(output, "ObserveMode: %v\n", err)
	} else {
		fmt.Fprintf(output, "Observe mode supported: %v\n", observeSupported)
		if observeSupported {
			observeEnabled, err := nfcAdapter.IsObserveModeEnabled()
			if err != nil {
				fmt.Fprintf(output, "IsObserveModeEnabled: %v\n", err)
			} else {
				fmt.Fprintf(output, "Observe mode enabled: %v\n", observeEnabled)
			}
		}
	}

	// Check Reader Option support.
	readerSupported, err := nfcAdapter.IsReaderOptionSupported()
	if err != nil {
		fmt.Fprintf(output, "ReaderOption: %v\n", err)
	} else {
		fmt.Fprintf(output, "Reader option supported: %v\n", readerSupported)
	}

	// Clean up the adapter global reference.
	vm.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(adapterObj)
		return nil
	})

	fmt.Fprintln(output, "")
	printConstants(output)

	return nil
}

func printConstants(output *bytes.Buffer) {
	fmt.Fprintln(output, "Reader mode flags:")
	fmt.Fprintf(output, "  NfcA:       0x%X\n", nfc.FlagReaderNfcA)
	fmt.Fprintf(output, "  NfcB:       0x%X\n", nfc.FlagReaderNfcB)
	fmt.Fprintf(output, "  NfcF:       0x%X\n", nfc.FlagReaderNfcF)
	fmt.Fprintf(output, "  NfcV:       0x%X\n", nfc.FlagReaderNfcV)
	fmt.Fprintf(output, "  Barcode:    0x%X\n", nfc.FlagReaderNfcBarcode)

	fmt.Fprintln(output, "\nAdapter states:")
	fmt.Fprintf(output, "  Off:        %d\n", nfc.StateOff)
	fmt.Fprintf(output, "  TurningOn:  %d\n", nfc.StateTurningOn)
	fmt.Fprintf(output, "  On:         %d\n", nfc.StateOn)
	fmt.Fprintf(output, "  TurningOff: %d\n", nfc.StateTurningOff)
}
