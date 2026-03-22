//go:build android

// Command pm demonstrates using the PackageManager API.
// It queries system feature availability and lists installed packages
// using live JNI calls.
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
	"github.com/AndroidGoLab/jni/content/pm"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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

	fmt.Fprintln(output, "=== PackageManager ===")

	// --- Obtain PackageManager from Context ---
	pmObj, err := ctx.GetPackageManager()
	if err != nil {
		return fmt.Errorf("GetPackageManager: %w", err)
	}
	if pmObj == nil || pmObj.Ref() == 0 {
		return fmt.Errorf("PackageManager is null")
	}

	// Wrap the raw JNI object in the generated PackageManager type.
	// GetPackageManager() already returns a global reference.
	mgr := pm.PackageManager{
		VM:  vm,
		Obj: pmObj,
	}
	fmt.Fprintln(output, "PackageManager: obtained")

	// --- Check system features ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "System Features:")

	features := []struct {
		name  string
		value string
	}{
		{"Camera", pm.FeatureCamera},
		{"CameraFront", pm.FeatureCameraFront},
		{"Bluetooth", pm.FeatureBluetooth},
		{"BluetoothLE", pm.FeatureBluetoothLe},
		{"NFC", pm.FeatureNfc},
		{"GPS", pm.FeatureLocationGps},
		{"Telephony", pm.FeatureTelephony},
		{"WiFi", pm.FeatureWifi},
		{"Fingerprint", pm.FeatureFingerprint},
		{"USB Host", pm.FeatureUsbHost},
		{"Touchscreen", pm.FeatureTouchscreen},
		{"Microphone", pm.FeatureMicrophone},
		{"Sensor Accel", pm.FeatureSensorAccelerometer},
		{"Sensor Gyro", pm.FeatureSensorGyroscope},
	}

	for _, f := range features {
		has, err := mgr.HasSystemFeature1(f.value)
		if err != nil {
			fmt.Fprintf(output, "  %-14s ERR: %v\n", f.name, err)
		} else if has {
			fmt.Fprintf(output, "  %-14s YES\n", f.name)
		} else {
			fmt.Fprintf(output, "  %-14s no\n", f.name)
		}
	}

	// --- Get our own package info ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Own Package Info:")

	// Get our package name from the context
	var pkgName string
	vm.Do(func(env *jni.Env) error {
		ctxCls := env.GetObjectClass(ctx.Obj)
		getPkgMid, err := env.GetMethodID(ctxCls, "getPackageName", "()Ljava/lang/String;")
		if err != nil {
			return nil
		}
		pkgObj, err := env.CallObjectMethod(ctx.Obj, getPkgMid)
		if err != nil {
			return nil
		}
		pkgName = env.GoString((*jni.String)(unsafe.Pointer(pkgObj)))
		return nil
	})

	if pkgName != "" {
		fmt.Fprintf(output, "  Package: %s\n", pkgName)

		// GetPackageInfo2_3(packageName string, flags int32) -> *jni.Object
		infoObj, err := mgr.GetPackageInfo2_3(pkgName, 0)
		if err != nil {
			fmt.Fprintf(output, "  Error: %v\n", err)
		} else if infoObj != nil && infoObj.Ref() != 0 {
			// Wrap in PackageInfo to call exported methods
			info := pm.PackageInfo{VM: vm, Obj: infoObj}

			versionCode, err := info.GetLongVersionCode()
			if err != nil {
				fmt.Fprintf(output, "  VersionCode: ERR %v\n", err)
			} else {
				fmt.Fprintf(output, "  VersionCode: %d\n", versionCode)
			}

			// Get versionName via raw JNI (it's a field, not a method)
			vm.Do(func(env *jni.Env) error {
				infoCls := env.GetObjectClass(infoObj)
				vnFid, err := env.GetFieldID(infoCls, "versionName", "Ljava/lang/String;")
				if err != nil {
					return nil
				}
				vnObj := env.GetObjectField(infoObj, vnFid)
				if vnObj != nil && vnObj.Ref() != 0 {
					vn := env.GoString((*jni.String)(unsafe.Pointer(vnObj)))
					fmt.Fprintf(output, "  VersionName: %s\n", vn)
				}

				// Get packageName field
				pnFid, err := env.GetFieldID(infoCls, "packageName", "Ljava/lang/String;")
				if err != nil {
					return nil
				}
				pnObj := env.GetObjectField(infoObj, pnFid)
				if pnObj != nil && pnObj.Ref() != 0 {
					pn := env.GoString((*jni.String)(unsafe.Pointer(pnObj)))
					fmt.Fprintf(output, "  PkgName: %s\n", pn)
				}
				return nil
			})
		}
	}

	// Filtered: GetInstalledPackages1_1 returns generic type (List<PackageInfo>)
	// listObj, err := mgr.GetInstalledPackages1_1(0)
	// ...

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "PM example complete.")
	return nil
}
