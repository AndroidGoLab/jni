//go:build android

// Command pm_package_info uses PackageManager to query info about
// installed packages. Lists a few installed apps with version code,
// and checks system features via typed wrappers.
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

	fmt.Fprintln(output, "=== Package Info ===")

	pmObj, err := ctx.GetPackageManager()
	if err != nil {
		return fmt.Errorf("GetPackageManager: %w", err)
	}
	if pmObj == nil || pmObj.Ref() == 0 {
		return fmt.Errorf("PackageManager is null")
	}

	mgr := pm.PackageManager{VM: vm, Obj: pmObj}
	fmt.Fprintln(output, "PackageManager: obtained")

	// Query well-known package names.
	packages := []string{
		"com.android.settings",
		"com.android.phone",
		"com.android.systemui",
		"com.android.vending",
		"com.google.android.gms",
	}

	// Also add our own package.
	ourPkg, err := ctx.GetPackageName()
	if err == nil && ourPkg != "" {
		packages = append([]string{ourPkg}, packages...)
	}

	for _, pkg := range packages {
		infoObj, err := mgr.GetPackageInfo2_3(pkg, 0)
		if err != nil {
			fmt.Fprintf(output, "\n%s: not found\n", pkg)
			continue
		}
		if infoObj == nil || infoObj.Ref() == 0 {
			fmt.Fprintf(output, "\n%s: null info\n", pkg)
			continue
		}

		info := pm.PackageInfo{VM: vm, Obj: infoObj}

		versionCode, _ := info.GetLongVersionCode()

		fmt.Fprintf(output, "\n%s\n", pkg)
		fmt.Fprintf(output, "  VersionCode: %d\n", versionCode)
	}

	// Check system features.
	fmt.Fprintln(output, "\nSystem Features:")
	features := []struct {
		name  string
		value string
	}{
		{"Camera", pm.FeatureCamera},
		{"Bluetooth", pm.FeatureBluetooth},
		{"WiFi", pm.FeatureWifi},
		{"Telephony", pm.FeatureTelephony},
		{"Fingerprint", pm.FeatureFingerprint},
	}
	for _, f := range features {
		has, err := mgr.HasSystemFeature1(f.value)
		if err != nil {
			fmt.Fprintf(output, "  %-12s ERR\n", f.name)
		} else if has {
			fmt.Fprintf(output, "  %-12s YES\n", f.name)
		} else {
			fmt.Fprintf(output, "  %-12s no\n", f.name)
		}
	}

	fmt.Fprintln(output, "\nPackage info complete.")
	return nil
}
