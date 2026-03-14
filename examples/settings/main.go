//go:build android

// Command settings demonstrates the Android Settings constants
// provided by the settings package. It is built as a c-shared
// library and packaged into an APK.
//
// The settings package exposes string constants that correspond to
// Android's Settings.System, Settings.Secure, and Settings.Global
// content provider tables, as well as commonly used setting keys.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/xaionaro-go/jni/provider/settings"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// The settings package defines constants for Android settings tables.
	// These correspond to the content provider table names used when
	// querying android.provider.Settings via a ContentResolver.

	// Settings tables.
	fmt.Fprintf(&output, "Settings.System table:  %q\n", settings.System)
	fmt.Fprintf(&output, "Settings.Secure table:  %q\n", settings.Secure)
	fmt.Fprintf(&output, "Settings.Global table:  %q\n", settings.Global)

	// Commonly used setting keys.
	fmt.Fprintf(&output, "Screen brightness key:  %q\n", settings.ScreenBrightness)
	fmt.Fprintf(&output, "Android ID key:         %q\n", settings.AndroidID)
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
