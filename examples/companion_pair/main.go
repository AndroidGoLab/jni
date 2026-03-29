//go:build android

// Command companion_pair uses CompanionDeviceManager to demonstrate
// the device association API. It calls GetAssociations, GetMyAssociations,
// BuildAssociationCancellationIntent, ToString, and exercises device
// profile constants.
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

	// 1-3. Check hardware prerequisites using PackageManager.
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

	// 4. Obtain CompanionDeviceManager.
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

	fmt.Fprintln(output, "\nCompanionDeviceManager: obtained OK")

	// 5. ToString.
	str, err := cdm.ToString()
	if err != nil {
		fmt.Fprintf(output, "  ToString: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  ToString: %s\n", str)
	}

	// 6. GetAssociations (returns List<String> on older APIs, List<AssociationInfo> on newer).
	fmt.Fprintln(output)
	fmt.Fprintln(output, "GetAssociations:")
	assocObj, err := cdm.GetAssociations()
	if err != nil {
		fmt.Fprintf(output, "  error: %v\n", err)
	} else if assocObj == nil || assocObj.Ref() == 0 {
		fmt.Fprintln(output, "  (null)")
	} else {
		fmt.Fprintf(output, "  list obtained (ref=%d)\n", assocObj.Ref())
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(assocObj)
			return nil
		})
	}

	// 7. GetMyAssociations (API 33+).
	fmt.Fprintln(output)
	fmt.Fprintln(output, "GetMyAssociations:")
	myAssocObj, err := cdm.GetMyAssociations()
	if err != nil {
		fmt.Fprintf(output, "  error: %v\n", err)
	} else if myAssocObj == nil || myAssocObj.Ref() == 0 {
		fmt.Fprintln(output, "  (null)")
	} else {
		fmt.Fprintf(output, "  list obtained (ref=%d)\n", myAssocObj.Ref())
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(myAssocObj)
			return nil
		})
	}

	// 8. BuildAssociationCancellationIntent (API 33+).
	fmt.Fprintln(output)
	fmt.Fprintln(output, "BuildAssociationCancellationIntent:")
	cancelIntent, err := cdm.BuildAssociationCancellationIntent()
	if err != nil {
		fmt.Fprintf(output, "  error: %v\n", err)
	} else if cancelIntent == nil || cancelIntent.Ref() == 0 {
		fmt.Fprintln(output, "  (null)")
	} else {
		fmt.Fprintf(output, "  intent obtained (ref=%d)\n", cancelIntent.Ref())
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(cancelIntent)
			return nil
		})
	}

	// 9. HasNotificationAccess (pass nil as ComponentName).
	fmt.Fprintln(output)
	fmt.Fprintln(output, "HasNotificationAccess (nil component):")
	hasNotif, err := cdm.HasNotificationAccess(nil)
	if err != nil {
		fmt.Fprintf(output, "  error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  result: %v\n", hasNotif)
	}

	// 10. DetachSystemDataTransport with a dummy association ID (API 35+).
	fmt.Fprintln(output)
	fmt.Fprintln(output, "DetachSystemDataTransport(0):")
	err = cdm.DetachSystemDataTransport(0)
	if err != nil {
		fmt.Fprintf(output, "  error: %v\n", err)
	} else {
		fmt.Fprintln(output, "  OK")
	}

	// 11. Device profile constants.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Device profiles:")
	fmt.Fprintf(output, "  WATCH                  = %s\n", companion.DeviceProfileWatch)
	fmt.Fprintf(output, "  COMPUTER               = %s\n", companion.DeviceProfileComputer)
	fmt.Fprintf(output, "  APP_STREAMING          = %s\n", companion.DeviceProfileAppStreaming)
	fmt.Fprintf(output, "  AUTOMOTIVE_PROJECTION  = %s\n", companion.DeviceProfileAutomotiveProjection)
	fmt.Fprintf(output, "  GLASSES                = %s\n", companion.DeviceProfileGlasses)
	fmt.Fprintf(output, "  NEARBY_DEVICE_STREAMING= %s\n", companion.DeviceProfileNearbyDeviceStreaming)

	// 12. Event constants.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Presence event constants:")
	fmt.Fprintf(output, "  BLE_APPEARED           = %d\n", companion.EventBleAppeared)
	fmt.Fprintf(output, "  BLE_DISAPPEARED        = %d\n", companion.EventBleDisappeared)
	fmt.Fprintf(output, "  BT_CONNECTED           = %d\n", companion.EventBtConnected)
	fmt.Fprintf(output, "  BT_DISCONNECTED        = %d\n", companion.EventBtDisconnected)

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Companion pair example complete.")
	return nil
}
