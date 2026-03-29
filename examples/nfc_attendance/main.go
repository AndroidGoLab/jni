//go:build android

// Command nfc_attendance builds an attendance tracking concept: checks NFC
// availability, prepares for tag scanning, and shows how tag data would be
// processed. Combines NFC with timestamp logging.
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

// attendanceEntry represents a simulated attendance record.
type attendanceEntry struct {
	TagID     string
	Name      string
	Timestamp time.Time
	Status    string
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== NFC Attendance System ===")

	// Check NFC hardware availability.
	adapter := nfc.Adapter{VM: vm}
	adapterObj, err := adapter.GetDefaultAdapter(ctx.Obj)

	nfcAvailable := false
	nfcEnabled := false

	if err != nil {
		fmt.Fprintf(output, "GetDefaultAdapter: %v\n", err)
	} else if adapterObj == nil || adapterObj.Ref() == 0 {
		fmt.Fprintln(output, "NFC hardware: not available")
	} else {
		nfcAvailable = true
		fmt.Fprintln(output, "NFC hardware: available")

		nfcAdapter := nfc.Adapter{VM: vm, Obj: adapterObj}
		var enabledVal bool
		enabledVal, err = nfcAdapter.IsEnabled()
		if err != nil {
			fmt.Fprintf(output, "IsEnabled: %v\n", err)
		} else {
			nfcEnabled = enabledVal
			fmt.Fprintf(output, "NFC enabled: %v\n", nfcEnabled)
		}

		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(adapterObj)
			return nil
		})
	}

	// Show readiness status.
	fmt.Fprintln(output, "\n=== System Readiness ===")
	fmt.Fprintf(output, "  NFC Hardware:  %s\n", boolStatus(nfcAvailable))
	fmt.Fprintf(output, "  NFC Enabled:   %s\n", boolStatus(nfcEnabled))
	if nfcAvailable && nfcEnabled {
		fmt.Fprintln(output, "  Status: READY for attendance scanning")
	} else if nfcAvailable {
		fmt.Fprintln(output, "  Status: Enable NFC in Settings")
	} else {
		fmt.Fprintln(output, "  Status: NFC hardware not found")
	}

	// Show discovery intents used for tag scanning.
	fmt.Fprintln(output, "\n=== Tag Discovery Configuration ===")
	fmt.Fprintf(output, "  NDEF discovered: %q\n", nfc.ActionNdefDiscovered)
	fmt.Fprintf(output, "  Tech discovered: %q\n", nfc.ActionTechDiscovered)
	fmt.Fprintf(output, "  Tag discovered:  %q\n", nfc.ActionTagDiscovered)

	// Reader mode flags for attendance.
	fmt.Fprintln(output, "\n=== Reader Mode Config ===")
	readerFlags := nfc.FlagReaderNfcA | nfc.FlagReaderNfcB
	fmt.Fprintf(output, "  FlagReaderNfcA: 0x%X\n", nfc.FlagReaderNfcA)
	fmt.Fprintf(output, "  FlagReaderNfcB: 0x%X\n", nfc.FlagReaderNfcB)
	fmt.Fprintf(output, "  Combined flags: 0x%X\n", readerFlags)

	// Simulate attendance records.
	fmt.Fprintln(output, "\n=== Simulated Attendance Log ===")
	now := time.Now()
	entries := []attendanceEntry{
		{"04:A2:B3:C4:D5:E6:F7", "Alice", now.Add(-2 * time.Hour), "CHECK-IN"},
		{"04:11:22:33:44:55:66", "Bob", now.Add(-90 * time.Minute), "CHECK-IN"},
		{"04:AA:BB:CC:DD:EE:FF", "Carol", now.Add(-1 * time.Hour), "CHECK-IN"},
		{"04:A2:B3:C4:D5:E6:F7", "Alice", now.Add(-30 * time.Minute), "CHECK-OUT"},
	}

	for _, e := range entries {
		fmt.Fprintf(output, "  [%s] %s %-9s tag=%s\n",
			e.Timestamp.Format("15:04:05"),
			e.Status,
			e.Name,
			e.TagID,
		)
	}

	fmt.Fprintf(output, "\n  Total entries: %d\n", len(entries))
	fmt.Fprintf(output, "  Unique tags:   %d\n", countUniqueTags(entries))

	fmt.Fprintln(output, "\n=== Attendance Workflow ===")
	fmt.Fprintln(output, "  1. enableReaderMode(activity, callback, flags)")
	fmt.Fprintln(output, "  2. onTagDiscovered(tag) receives Tag object")
	fmt.Fprintln(output, "  3. Read tag ID via tag.getId()")
	fmt.Fprintln(output, "  4. Log timestamp + tag ID as attendance record")
	fmt.Fprintln(output, "  5. Optionally read NDEF for employee name")

	fmt.Fprintln(output, "\nNFC attendance example completed.")
	return nil
}

func boolStatus(v bool) string {
	if v {
		return "OK"
	}
	return "NOT AVAILABLE"
}

func countUniqueTags(entries []attendanceEntry) int {
	seen := make(map[string]bool)
	for _, e := range entries {
		seen[e.TagID] = true
	}
	return len(seen)
}
