//go:build android

// Command environment demonstrates the Android Environment API constants.
// It is built as a c-shared library and packaged into an APK.
//
// This package provides Go bindings for android.os.Environment. The class
// methods are exposed via an unexported environment type, but the package
// exports constants for standard directory names and external storage states.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/AndroidGoLab/jni/os/environment"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// Exported constants for external storage state values.
	// These correspond to Environment.MEDIA_* constants in Java.
	fmt.Fprintln(&output, "External storage state constants:")
	fmt.Fprintf(&output, "  MediaMounted:         %s\n", environment.MediaMounted)
	fmt.Fprintf(&output, "  MediaMountedReadOnly: %s\n", environment.MediaMountedReadOnly)
	fmt.Fprintf(&output, "  MediaRemoved:         %s\n", environment.MediaRemoved)
	fmt.Fprintf(&output, "  MediaUnmounted:       %s\n", environment.MediaUnmounted)

	// The environment class methods (getExternalStorageDirectory,
	// getExternalStoragePublicDirectory, getExternalStorageState, etc.)
	// are accessed via an unexported environment type.
	//
	// The javaFile type (also unexported) wraps java.io.File with:
	//   javaFile.getAbsolutePath() string
	//     Returns the absolute path of the file.
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
