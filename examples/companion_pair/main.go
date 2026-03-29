//go:build android

// Command companion_pair uses CompanionDeviceManager to demonstrate
// the device association API surface. Checks availability and shows
// pairing workflow concepts.
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
	"strings"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/companion"
	"github.com/AndroidGoLab/jni/content/pm"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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

	fmt.Fprintln(output, "=== Companion Pair ===")

	// Check if Bluetooth and BLE are available (prerequisites).
	pmObj, err := ctx.GetPackageManager()
	if err == nil && pmObj != nil && pmObj.Ref() != 0 {
		mgr := pm.PackageManager{VM: vm, Obj: pmObj}

		hasBT, _ := mgr.HasSystemFeature1(pm.FeatureBluetooth)
		hasBLE, _ := mgr.HasSystemFeature1(pm.FeatureBluetoothLe)
		hasWifi, _ := mgr.HasSystemFeature1(pm.FeatureWifi)

		fmt.Fprintln(output, "Prerequisites:")
		fmt.Fprintf(output, "  Bluetooth: %v\n", hasBT)
		fmt.Fprintf(output, "  BLE: %v\n", hasBLE)
		fmt.Fprintf(output, "  WiFi: %v\n", hasWifi)
	}

	// Try to obtain CompanionDeviceManager.
	cdm, err := companion.NewDeviceManager(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "service not available") {
			fmt.Fprintln(output, "\nCompanionDeviceManager: NOT AVAILABLE")
			fmt.Fprintln(output, "(Requires API 26+)")
			return nil
		}
		return fmt.Errorf("companion.NewDeviceManager: %w", err)
	}
	defer cdm.Close()

	fmt.Fprintln(output, "\nCompanionDeviceManager: available")

	// Pairing workflow overview.
	fmt.Fprintln(output, "\nPairing Workflow:")
	fmt.Fprintln(output, "  1. Build AssociationRequest")
	fmt.Fprintln(output, "     (filter by BT, BLE, or WiFi)")
	fmt.Fprintln(output, "  2. Call CDM.associate(request, callback)")
	fmt.Fprintln(output, "  3. System shows device picker")
	fmt.Fprintln(output, "  4. Callback receives IntentSender")
	fmt.Fprintln(output, "  5. Launch IntentSender for result")
	fmt.Fprintln(output, "  6. Get device info from result Intent")

	// Note about filtered generic methods.
	fmt.Fprintln(output, "\nNote: GetAssociations and")
	fmt.Fprintln(output, "GetMyAssociations return generic")
	fmt.Fprintln(output, "types (List<>) and are filtered.")

	fmt.Fprintln(output, "\nCompanion pair complete.")
	return nil
}
