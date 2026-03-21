//go:build android

// Command settings demonstrates the Android Settings API
// provided by the settings package. It is built as a c-shared
// library and packaged into an APK.
//
// The settings package wraps android.provider.Settings, providing
// methods for reading system, secure, and global settings.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	_ "github.com/AndroidGoLab/jni/provider/settings"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// The settings package provides access to Android settings via
	// the Settings content provider. The settings types (system, secure,
	// global) provide methods for reading setting values.
	fmt.Fprintln(&output, "settings package available (system, secure, global)")
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
