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
#include <android/native_activity.h>
extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static void _setCallbacks(ANativeActivity* a) { a->callbacks->onResume = _onResume; a->callbacks->onNativeWindowCreated = _onWindowCreated; }
*/
import "C"
import (
	"unsafe"
	"bytes"
	"fmt"

	_ "github.com/AndroidGoLab/jni/content/preferences"
	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
)

func main() {}

func init() { exampleui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	exampleui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	exampleui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	exampleui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
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
	fmt.Fprintln(output, "SharedPreferences: typed getters + editor for batch writes")
	return nil
}
