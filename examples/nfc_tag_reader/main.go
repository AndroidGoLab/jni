//go:build android

// Command nfc_tag_reader gets the NFC adapter, checks if NFC is enabled,
// demonstrates NDEF message parsing constants and tag technology types.
// Since we cannot trigger a tag scan from code, this shows API readiness.
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
	"github.com/AndroidGoLab/jni/nfc/tech"
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

	fmt.Fprintln(output, "=== NFC Tag Reader ===")

	// Obtain the default NFC adapter.
	adapter := nfc.Adapter{VM: vm}
	adapterObj, err := adapter.GetDefaultAdapter(ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "GetDefaultAdapter: %v\n", err)
		fmt.Fprintln(output, "NFC may not be available on this device.")
		printTagReaderInfo(output)
		return nil
	}
	if adapterObj == nil || adapterObj.Ref() == 0 {
		fmt.Fprintln(output, "NFC adapter not available (no NFC hardware).")
		printTagReaderInfo(output)
		return nil
	}
	defer func() {
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(adapterObj)
			return nil
		})
	}()

	nfcAdapter := nfc.Adapter{VM: vm, Obj: adapterObj}
	fmt.Fprintln(output, "NFC adapter obtained successfully.")

	// Check NFC enabled state.
	enabled, err := nfcAdapter.IsEnabled()
	if err != nil {
		fmt.Fprintf(output, "IsEnabled: %v\n", err)
	} else {
		fmt.Fprintf(output, "NFC enabled: %v\n", enabled)
		if !enabled {
			fmt.Fprintln(output, "Enable NFC in Settings to scan tags.")
		}
	}

	// Check secure NFC support.
	secureSupported, err := nfcAdapter.IsSecureNfcSupported()
	if err != nil {
		fmt.Fprintf(output, "IsSecureNfcSupported: %v\n", err)
	} else {
		fmt.Fprintf(output, "Secure NFC supported: %v\n", secureSupported)
	}

	printTagReaderInfo(output)

	fmt.Fprintln(output, "\nNFC tag reader example completed.")
	return nil
}

func printTagReaderInfo(output *bytes.Buffer) {
	fmt.Fprintln(output, "\n=== NDEF Record Types ===")
	fmt.Fprintf(output, "  TNF_WELL_KNOWN:   %d\n", nfc.TnfWellKnown)
	fmt.Fprintf(output, "  TNF_MIME_MEDIA:   %d\n", nfc.TnfMimeMedia)
	fmt.Fprintf(output, "  TNF_ABSOLUTE_URI: %d\n", nfc.TnfAbsoluteUri)
	fmt.Fprintf(output, "  TNF_EXTERNAL:     %d\n", nfc.TnfExternalType)
	fmt.Fprintf(output, "  TNF_EMPTY:        %d\n", nfc.TnfEmpty)
	fmt.Fprintf(output, "  TNF_UNKNOWN:      %d\n", nfc.TnfUnknown)
	fmt.Fprintf(output, "  TNF_UNCHANGED:    %d\n", nfc.TnfUnchanged)

	fmt.Fprintln(output, "\n=== Well-Known RTD Constants ===")
	fmt.Fprintf(output, "  RTD_TEXT:          %q\n", nfc.RtdText)
	fmt.Fprintf(output, "  RTD_URI:           %q\n", nfc.RtdUri)
	fmt.Fprintf(output, "  RTD_SMART_POSTER:  %q\n", nfc.RtdSmartPoster)

	fmt.Fprintln(output, "\n=== Tag Technology Types ===")
	fmt.Fprintf(output, "  NFC Forum Type 1: %d\n", nfc.NfcForumType1)
	fmt.Fprintf(output, "  NFC Forum Type 2: %d\n", nfc.NfcForumType2)
	fmt.Fprintf(output, "  NFC Forum Type 3: %d\n", nfc.NfcForumType3)
	fmt.Fprintf(output, "  NFC Forum Type 4: %d\n", nfc.NfcForumType4)
	fmt.Fprintf(output, "  MIFARE Classic:   %d\n", nfc.MifareClassic)

	fmt.Fprintln(output, "\n=== Reader Mode Flags ===")
	fmt.Fprintf(output, "  NfcA:    0x%X\n", nfc.FlagReaderNfcA)
	fmt.Fprintf(output, "  NfcB:    0x%X\n", nfc.FlagReaderNfcB)
	fmt.Fprintf(output, "  NfcF:    0x%X\n", nfc.FlagReaderNfcF)
	fmt.Fprintf(output, "  NfcV:    0x%X\n", nfc.FlagReaderNfcV)
	fmt.Fprintf(output, "  Barcode: 0x%X\n", nfc.FlagReaderNfcBarcode)

	fmt.Fprintln(output, "\n=== MIFARE Constants ===")
	fmt.Fprintf(output, "  Size 1K:   %d\n", tech.Size1k)
	fmt.Fprintf(output, "  Size 4K:   %d\n", tech.Size4k)
	fmt.Fprintf(output, "  Block:     %d bytes\n", tech.BlockSize)

	fmt.Fprintln(output, "\n=== Tag Discovery Intents ===")
	fmt.Fprintf(output, "  NDEF_DISCOVERED: %q\n", nfc.ActionNdefDiscovered)
	fmt.Fprintf(output, "  TECH_DISCOVERED: %q\n", nfc.ActionTechDiscovered)
	fmt.Fprintf(output, "  TAG_DISCOVERED:  %q\n", nfc.ActionTagDiscovered)
}
