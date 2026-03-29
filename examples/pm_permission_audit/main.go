//go:build android

// Command pm_permission_audit uses PackageManager to get info about our own
// package: version info and system feature checks, all via typed wrappers.
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

	fmt.Fprintln(output, "=== PM Permission Audit ===")

	// Get PackageManager.
	pmObj, err := ctx.GetPackageManager()
	if err != nil {
		return fmt.Errorf("GetPackageManager: %w", err)
	}
	if pmObj == nil || pmObj.Ref() == 0 {
		return fmt.Errorf("PackageManager is null")
	}

	mgr := pm.PackageManager{VM: vm, Obj: pmObj}

	// Get our package name.
	pkgName, err := ctx.GetPackageName()
	if err != nil {
		return fmt.Errorf("GetPackageName: %w", err)
	}
	fmt.Fprintf(output, "Package: %s\n\n", pkgName)

	// Get package info with GET_PERMISSIONS flag (0x1000 = 4096).
	const getPermissions = 4096
	infoObj, err := mgr.GetPackageInfo2_3(pkgName, getPermissions)
	if err != nil {
		fmt.Fprintf(output, "GetPackageInfo: %v\n", err)
		return nil
	}
	if infoObj == nil || infoObj.Ref() == 0 {
		fmt.Fprintln(output, "PackageInfo is null")
		return nil
	}

	info := pm.PackageInfo{VM: vm, Obj: infoObj}

	// Version info.
	versionCode, err := info.GetLongVersionCode()
	if err != nil {
		fmt.Fprintf(output, "VersionCode: ERR %v\n", err)
	} else {
		fmt.Fprintf(output, "VersionCode: %d\n", versionCode)
	}

	// Check system features.
	fmt.Fprintln(output, "\nKey System Features:")
	features := []struct {
		name  string
		value string
	}{
		{"Camera", pm.FeatureCamera},
		{"Bluetooth", pm.FeatureBluetooth},
		{"WiFi", pm.FeatureWifi},
		{"Telephony", pm.FeatureTelephony},
		{"NFC", pm.FeatureNfc},
	}

	for _, f := range features {
		has, err := mgr.HasSystemFeature1(f.value)
		if err != nil {
			fmt.Fprintf(output, "  %-12s ERR: %v\n", f.name, err)
		} else if has {
			fmt.Fprintf(output, "  %-12s YES\n", f.name)
		} else {
			fmt.Fprintf(output, "  %-12s no\n", f.name)
		}
	}

	fmt.Fprintln(output, "\nPM permission audit complete.")
	return nil
}
