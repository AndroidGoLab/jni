//go:build android

// Command permission demonstrates the Android permission string constants.
// It is built as a c-shared library and packaged into an APK.
//
// This example lists all exported permission constants from the
// permission package. These strings are used with the Android runtime
// permission system (e.g. ActivityCompat.requestPermissions).
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/AndroidGoLab/jni/content/permission"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// All permission constants match the android.Manifest.permission.* values.
	fmt.Fprintln(&output, "Android permission constants:")
	fmt.Fprintf(&output, "  AccessFineLocation       = %q\n", permission.AccessFineLocation)
	fmt.Fprintf(&output, "  AccessCoarseLocation     = %q\n", permission.AccessCoarseLocation)
	fmt.Fprintf(&output, "  Camera                   = %q\n", permission.Camera)
	fmt.Fprintf(&output, "  RecordAudio              = %q\n", permission.RecordAudio)
	fmt.Fprintf(&output, "  Bluetooth                = %q\n", permission.Bluetooth)
	fmt.Fprintf(&output, "  BluetoothConnect         = %q\n", permission.BluetoothConnect)
	fmt.Fprintf(&output, "  BluetoothScan            = %q\n", permission.BluetoothScan)
	fmt.Fprintf(&output, "  Internet                 = %q\n", permission.Internet)
	fmt.Fprintf(&output, "  ReadContacts             = %q\n", permission.ReadContacts)
	fmt.Fprintf(&output, "  WriteContacts            = %q\n", permission.WriteContacts)
	fmt.Fprintf(&output, "  ReadExternalStorage      = %q\n", permission.ReadExternalStorage)
	fmt.Fprintf(&output, "  WriteExternalStorage     = %q\n", permission.WriteExternalStorage)
	fmt.Fprintf(&output, "  ReadMediaImages          = %q\n", permission.ReadMediaImages)
	fmt.Fprintf(&output, "  ReadMediaVideo           = %q\n", permission.ReadMediaVideo)
	fmt.Fprintf(&output, "  ReadMediaAudio           = %q\n", permission.ReadMediaAudio)
	fmt.Fprintf(&output, "  PostNotifications        = %q\n", permission.PostNotifications)
	fmt.Fprintf(&output, "  AccessBackgroundLocation = %q\n", permission.AccessBackgroundLocation)
	fmt.Fprintf(&output, "  ReadPhoneState           = %q\n", permission.ReadPhoneState)
	fmt.Fprintf(&output, "  NearbyWifiDevices        = %q\n", permission.NearbyWifiDevices)

	// Example: build a list of permissions to request.
	required := []string{
		permission.AccessFineLocation,
		permission.Camera,
		permission.RecordAudio,
	}
	fmt.Fprintf(&output, "\nPermissions to request: %v\n", required)
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
