//go:build android

// Command nfc_inventory demonstrates NFC inventory tracking concepts: checks
// the adapter, builds NDEF records for inventory items (URI + text records),
// and demonstrates a batch tag processing workflow.
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
	"time"
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

// inventoryItem represents an item to be written to an NFC tag.
type inventoryItem struct {
	SKU      string
	Name     string
	Location string
	URI      string
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== NFC Inventory Tracker ===")

	// Check NFC adapter.
	adapter := nfc.Adapter{VM: vm}
	adapterObj, err := adapter.GetDefaultAdapter(ctx.Obj)
	nfcReady := false
	if err != nil {
		fmt.Fprintf(output, "GetDefaultAdapter: %v\n", err)
	} else if adapterObj == nil || adapterObj.Ref() == 0 {
		fmt.Fprintln(output, "NFC hardware: not available")
	} else {
		nfcAdapter := nfc.Adapter{VM: vm, Obj: adapterObj}
		enabled, err := nfcAdapter.IsEnabled()
		if err != nil {
			fmt.Fprintf(output, "IsEnabled: %v\n", err)
		} else {
			nfcReady = enabled
			fmt.Fprintf(output, "NFC adapter: available, enabled=%v\n", enabled)
		}

		// Check reader option for batch scanning.
		readerSupported, err := nfcAdapter.IsReaderOptionSupported()
		if err != nil {
			fmt.Fprintf(output, "IsReaderOptionSupported: %v\n", err)
		} else {
			fmt.Fprintf(output, "Reader option supported: %v\n", readerSupported)
		}

		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(adapterObj)
			return nil
		})
	}

	if nfcReady {
		fmt.Fprintln(output, "Status: READY for inventory scanning")
	} else {
		fmt.Fprintln(output, "Status: NFC not ready (continuing with demo)")
	}

	// Define inventory items.
	items := []inventoryItem{
		{"SKU-001", "Laptop Dell XPS 15", "Shelf A1", "https://inventory.example.com/items/001"},
		{"SKU-002", "Monitor LG 27\"", "Shelf B3", "https://inventory.example.com/items/002"},
		{"SKU-003", "Keyboard Mechanical", "Shelf C2", "https://inventory.example.com/items/003"},
	}

	// Build NDEF records for each inventory item.
	fmt.Fprintln(output, "\n=== Building NDEF Records ===")
	record := nfc.NdefRecord{VM: vm}

	for i, item := range items {
		fmt.Fprintf(output, "\nItem %d: %s (%s)\n", i+1, item.Name, item.SKU)

		// Create a URI record for the item.
		uriObj, err := record.CreateUri1_1(item.URI)
		if err != nil {
			fmt.Fprintf(output, "  CreateUri: %v\n", err)
			continue
		}

		uriRec := nfc.NdefRecord{VM: vm, Obj: uriObj}
		tnf, err := uriRec.GetTnf()
		if err == nil {
			fmt.Fprintf(output, "  URI record TNF: %d\n", tnf)
		}

		// Create a text record with the item description.
		desc := fmt.Sprintf("%s | %s | %s", item.SKU, item.Name, item.Location)
		textObj, err := record.CreateTextRecord("en", desc)
		if err != nil {
			fmt.Fprintf(output, "  CreateTextRecord: %v\n", err)
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(uriObj)
				return nil
			})
			continue
		}

		textRec := nfc.NdefRecord{VM: vm, Obj: textObj}
		textTnf, err := textRec.GetTnf()
		if err == nil {
			fmt.Fprintf(output, "  Text record TNF: %d\n", textTnf)
		}

		// Build an NDEF message from the two records.
		msg, err := nfc.NewNdefMessage(vm, uriRec.Obj, textRec.Obj)
		if err != nil {
			fmt.Fprintf(output, "  NewNdefMessage: %v\n", err)
		} else {
			byteLen, err := msg.GetByteArrayLength()
			if err == nil {
				fmt.Fprintf(output, "  Message size: %d bytes\n", byteLen)
			}
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(msg.Obj)
				return nil
			})
		}

		// Clean up records.
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(uriObj)
			env.DeleteGlobalRef(textObj)
			return nil
		})
	}

	// Simulated batch scan results.
	fmt.Fprintln(output, "\n=== Simulated Batch Scan ===")
	now := time.Now()
	type scanResult struct {
		TagID   string
		SKU     string
		ScannedAt time.Time
		Found   bool
	}
	scans := []scanResult{
		{"04:A1:B2:C3:D4:E5:F6", "SKU-001", now.Add(-5 * time.Minute), true},
		{"04:11:22:33:44:55:66", "SKU-002", now.Add(-3 * time.Minute), true},
		{"04:AA:BB:CC:DD:EE:FF", "SKU-003", now.Add(-1 * time.Minute), true},
	}

	for _, s := range scans {
		status := "FOUND"
		if !s.Found {
			status = "MISSING"
		}
		fmt.Fprintf(output, "  [%s] %s %s tag=%s\n",
			s.ScannedAt.Format("15:04:05"), status, s.SKU, s.TagID)
	}

	fmt.Fprintf(output, "\n  Scanned: %d/%d items\n", len(scans), len(items))

	fmt.Fprintln(output, "\n=== Batch Processing Workflow ===")
	fmt.Fprintln(output, "  1. Load expected inventory from database")
	fmt.Fprintln(output, "  2. enableReaderMode with NfcA + NfcB flags")
	fmt.Fprintln(output, "  3. For each tag: read NDEF, extract SKU")
	fmt.Fprintln(output, "  4. Match against expected inventory")
	fmt.Fprintln(output, "  5. Report found/missing items")

	fmt.Fprintln(output, "\nNFC inventory example completed.")
	return nil
}
