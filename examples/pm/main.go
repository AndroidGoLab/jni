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

	"github.com/xaionaro-go/jni/content/pm"
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
	fmt.Fprintf(&output, "  FeatureBluetoothLE = %q\n", pm.FeatureBluetoothLE)
	fmt.Fprintf(&output, "  FeatureNFC         = %q\n", pm.FeatureNFC)
	fmt.Fprintf(&output, "  FeatureGPS         = %q\n", pm.FeatureGPS)
	fmt.Fprintf(&output, "  FeatureTelephony   = %q\n", pm.FeatureTelephony)
	fmt.Fprintf(&output, "  FeatureWifi        = %q\n", pm.FeatureWifi)
	fmt.Fprintf(&output, "  FeatureFingerprint = %q\n", pm.FeatureFingerprint)
	fmt.Fprintf(&output, "  FeatureUSBHost     = %q\n", pm.FeatureUSBHost)

	// --- PackageInfo data class ---
	// PackageInfo holds data extracted from android.content.pm.PackageInfo.
	info := pm.PackageInfo{
		PackageName:  "com.example.myapp",
		VersionName:  "2.1.0",
		VersionCode:  42,
		FirstInstall: 1700000000000, // Unix millis
		LastUpdate:   1700100000000,
	}
	fmt.Fprintf(&output, "\nPackageInfo:\n")
	fmt.Fprintf(&output, "  PackageName  = %s\n", info.PackageName)
	fmt.Fprintf(&output, "  VersionName  = %s\n", info.VersionName)
	fmt.Fprintf(&output, "  VersionCode  = %d\n", info.VersionCode)
	fmt.Fprintf(&output, "  FirstInstall = %d ms\n", info.FirstInstall)
	fmt.Fprintf(&output, "  LastUpdate   = %d ms\n", info.LastUpdate)

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
		pm.FeatureNFC,
		pm.FeatureGPS,
		pm.FeatureWifi,
		pm.FeatureFingerprint,
		pm.FeatureUSBHost,
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
