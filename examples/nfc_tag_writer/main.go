//go:build android

// Command nfc_tag_writer demonstrates NdefMessage and NdefRecord creation.
// It builds an NDEF message with a URI record and shows the write-ready
// API surface.
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

	fmt.Fprintln(output, "=== NFC Tag Writer ===")

	// Check NFC availability.
	adapter := nfc.Adapter{VM: vm}
	adapterObj, err := adapter.GetDefaultAdapter(ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "GetDefaultAdapter: %v\n", err)
	} else if adapterObj == nil || adapterObj.Ref() == 0 {
		fmt.Fprintln(output, "NFC adapter not available.")
	} else {
		nfcAdapter := nfc.Adapter{VM: vm, Obj: adapterObj}
		enabled, err := nfcAdapter.IsEnabled()
		if err != nil {
			fmt.Fprintf(output, "IsEnabled: %v\n", err)
		} else {
			fmt.Fprintf(output, "NFC enabled: %v\n", enabled)
		}
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(adapterObj)
			return nil
		})
	}

	// Create a URI record using NdefRecord.createUri.
	fmt.Fprintln(output, "\n=== Creating NDEF Records ===")

	record := nfc.NdefRecord{VM: vm}
	uriRecordObj, err := record.CreateUri1_1("https://github.com/AndroidGoLab/jni")
	if err != nil {
		fmt.Fprintf(output, "CreateUri: %v\n", err)
		return nil
	}
	fmt.Fprintln(output, "URI record created successfully.")

	// Wrap the returned record to query its properties.
	uriRecord := nfc.NdefRecord{VM: vm, Obj: uriRecordObj}

	tnf, err := uriRecord.GetTnf()
	if err != nil {
		fmt.Fprintf(output, "GetTnf: %v\n", err)
	} else {
		fmt.Fprintf(output, "  TNF: %d (WELL_KNOWN=%d)\n", tnf, nfc.TnfWellKnown)
	}

	recordStr, err := uriRecord.ToString()
	if err != nil {
		fmt.Fprintf(output, "ToString: %v\n", err)
	} else {
		fmt.Fprintf(output, "  Record: %s\n", recordStr)
	}

	recordBytes, err := uriRecord.ToByteArray()
	if err != nil {
		fmt.Fprintf(output, "ToByteArray: %v\n", err)
	} else if recordBytes != nil {
		var byteLen int32
		vm.Do(func(env *jni.Env) error {
			byteLen = env.GetArrayLength((*jni.Array)(unsafe.Pointer(recordBytes)))
			env.DeleteGlobalRef(recordBytes)
			return nil
		})
		fmt.Fprintf(output, "  Byte size: %d bytes\n", byteLen)
	}

	// Create a text record.
	fmt.Fprintln(output, "\n=== Text Record ===")
	textRecordObj, err := record.CreateTextRecord("en", "Hello from Go JNI!")
	if err != nil {
		fmt.Fprintf(output, "CreateTextRecord: %v\n", err)
	} else {
		textRecord := nfc.NdefRecord{VM: vm, Obj: textRecordObj}

		textTnf, err := textRecord.GetTnf()
		if err == nil {
			fmt.Fprintf(output, "  TNF: %d\n", textTnf)
		}
		textStr, err := textRecord.ToString()
		if err == nil {
			fmt.Fprintf(output, "  Record: %s\n", textStr)
		}

		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(textRecordObj)
			return nil
		})
	}

	// Create an NdefMessage from the URI record.
	fmt.Fprintln(output, "\n=== NDEF Message ===")
	msg, err := nfc.NewNdefMessage(vm, uriRecord.Obj, nil)
	if err != nil {
		fmt.Fprintf(output, "NewNdefMessage: %v\n", err)
	} else {
		byteArrayLen, err := msg.GetByteArrayLength()
		if err != nil {
			fmt.Fprintf(output, "GetByteArrayLength: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Message size: %d bytes\n", byteArrayLen)
		}

		msgStr, err := msg.ToString()
		if err != nil {
			fmt.Fprintf(output, "ToString: %v\n", err)
		} else {
			fmt.Fprintf(output, "  Message: %s\n", msgStr)
		}

		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(msg.Obj)
			return nil
		})
	}

	// Clean up the URI record.
	vm.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(uriRecordObj)
		return nil
	})

	fmt.Fprintln(output, "\n=== Write-Ready Summary ===")
	fmt.Fprintln(output, "To write to a tag:")
	fmt.Fprintln(output, "  1. Enable reader mode with FlagReaderNfcA")
	fmt.Fprintln(output, "  2. Get Tag from onTagDiscovered callback")
	fmt.Fprintln(output, "  3. Connect via Ndef.get(tag)")
	fmt.Fprintln(output, "  4. Call ndef.writeNdefMessage(message)")

	fmt.Fprintln(output, "\nNFC tag writer example completed.")
	return nil
}
