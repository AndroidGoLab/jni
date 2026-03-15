//go:build android

// Command build_info demonstrates reading Android device build
// information such as manufacturer, model, and SDK version.
//
// The build package wraps android.os.Build and android.os.Build.VERSION,
// providing GetBuildInfo() and GetVersionInfo() functions that read
// static fields into Go structs.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/os/build"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	vm := jni.VMFromPtr(unsafe.Pointer(cvm))
	if err := run(vm); err != nil {
		fmt.Fprintf(&output, "ERROR: %v\n", err)
	}
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}

func run(vm *jni.VM) error {
	// GetBuildInfo reads static fields from android.os.Build:
	// Manufacturer, Model, Brand, Device, Hardware, Product, Fingerprint, ID.
	info, err := build.GetBuildInfo(vm)
	if err != nil {
		return fmt.Errorf("build.GetBuildInfo: %w", err)
	}
	fmt.Fprintf(&output, "build info: %+v\n", info)

	// GetVersionInfo reads android.os.Build.VERSION fields:
	// SDKInt (API level as int32) and Release (version string like "14").
	ver, err := build.GetVersionInfo(vm)
	if err != nil {
		return fmt.Errorf("build.GetVersionInfo: %w", err)
	}
	fmt.Fprintf(&output, "version info: %+v\n", ver)

	return nil
}
