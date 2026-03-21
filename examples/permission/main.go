//go:build android

// Command permission demonstrates the Android permission package.
// It is built as a c-shared library and packaged into an APK.
//
// The permission package provides Init for initializing the JNI
// bindings related to Android runtime permissions.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	_ "github.com/AndroidGoLab/jni/content/permission"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// The content/permission package provides Init for JNI initialization.
	// Permission checking and requesting is done via the app.Context and
	// the Android runtime permission system.
	fmt.Fprintln(&output, "permission package available (Init for JNI setup)")
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
