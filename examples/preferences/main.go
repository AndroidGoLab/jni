//go:build android

// Command preferences demonstrates using the SharedPreferences API.
// It is built as a c-shared library and packaged into an APK.
//
// The preferences package wraps android.content.SharedPreferences,
// obtained from Context.getSharedPreferences(). It provides typed
// getters (string, int, bool, float, long), key existence checks,
// and an editor for batch writes.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	_ "github.com/xaionaro-go/jni/content/preferences"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// The Preferences type wraps android.content.SharedPreferences.
	// It is obtained from Context.getSharedPreferences() and provides:
	//
	// Exported methods:
	//   prefs.Close()
	//   prefs.GetString(key, defaultVal string) (string, error)
	//   prefs.GetInt(key string, defaultVal int32) (int32, error)
	//   prefs.GetBool(key string, defaultVal bool) (bool, error)
	//   prefs.GetFloat(key string, defaultVal float32) (float32, error)
	//   prefs.GetLong(key string, defaultVal int64) (int64, error)
	//   prefs.Contains(key string) (bool, error)
	//
	// Unexported editor methods:
	//   editor.putString(key, value)
	//   editor.putInt(key, value)
	//   editor.putBoolean(key, value)
	//   editor.putFloat(key, value)
	//   editor.putLong(key, value)
	//   editor.remove(key)
	//   editor.clear()
	//   editor.apply()
	fmt.Fprintln(&output, "SharedPreferences: typed getters + editor for batch writes")
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
