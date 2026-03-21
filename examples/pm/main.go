//go:build android

// Command pm demonstrates using the PackageManager API.
// It is built as a c-shared library and packaged into an APK.
//
// This example shows all system feature constants and the exported
// PackageInfo data class. The Manager type provides methods for
// feature detection, package queries, and activity resolution. In a
// complete implementation, the Manager is obtained via
// NewManager(ctx) (not yet generated).
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/AndroidGoLab/jni/content/pm"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// --- Constants ---
	// Feature constants for HasSystemFeature checks.
	fmt.Fprintln(&output, "System feature constants:")
	fmt.Fprintf(&output, "  FeatureCamera      = %q\n", pm.FeatureCamera)
	fmt.Fprintf(&output, "  FeatureCameraFront = %q\n", pm.FeatureCameraFront)
	fmt.Fprintf(&output, "  FeatureBluetooth   = %q\n", pm.FeatureBluetooth)
	fmt.Fprintf(&output, "  FeatureBluetoothLe = %q\n", pm.FeatureBluetoothLe)
	fmt.Fprintf(&output, "  FeatureNfc         = %q\n", pm.FeatureNfc)
	fmt.Fprintf(&output, "  FeatureLocationGps = %q\n", pm.FeatureLocationGps)
	fmt.Fprintf(&output, "  FeatureTelephony   = %q\n", pm.FeatureTelephony)
	fmt.Fprintf(&output, "  FeatureWifi        = %q\n", pm.FeatureWifi)
	fmt.Fprintf(&output, "  FeatureFingerprint = %q\n", pm.FeatureFingerprint)
	fmt.Fprintf(&output, "  FeatureUSBHost     = %q\n", pm.FeatureUsbHost)

	// --- PackageInfo type ---
	// PackageInfo wraps android.content.pm.PackageInfo with VM and Obj
	// fields for JNI access. Its methods (DescribeContents, etc.) are
	// called through JNI.
	var info pm.PackageInfo
	_ = info

	// --- Manager methods ---
	// The Manager wraps android.content.pm.PackageManager and provides:
	//
	//   mgr.HasSystemFeature(feature string) (bool, error) [exported]
	//     Check whether the device has a hardware feature.
	//     Use with feature constants like pm.FeatureCamera.
	//
	//   mgr.getPackageInfoRaw(pkgName string, flags int32) [unexported]
	//     Query info about a specific installed package.
	//     The raw JNI object is converted to PackageInfo via
	//     extractPackageInfo.
	//
	//   mgr.getInstalledPackagesRaw(flags int32) [unexported]
	//     List all installed packages.
	//
	//   mgr.resolveActivityRaw(intent, flags) [unexported]
	//     Check if an Intent can be handled before starting it.

	// Example: list features to check at startup.
	featuresToCheck := []string{
		pm.FeatureCamera,
		pm.FeatureBluetooth,
		pm.FeatureNfc,
		pm.FeatureLocationGps,
		pm.FeatureWifi,
		pm.FeatureFingerprint,
		pm.FeatureUsbHost,
	}
	fmt.Fprintf(&output, "\nFeatures to check at startup: %d\n", len(featuresToCheck))
	for _, f := range featuresToCheck {
		fmt.Fprintf(&output, "  %s\n", f)
	}
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
