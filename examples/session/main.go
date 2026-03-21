//go:build android

// Command session demonstrates using the Android MediaSession API
// to query active media sessions, wrapped by the session package.
// It is built as a c-shared library and packaged into an APK using
// the shared apk.mk infrastructure.
//
// The session package wraps android.media.session.MediaSessionManager.
// It provides a Manager obtained via NewManager, with methods for
// querying active sessions and registering listeners for session changes.
// The mediaController data class extracts PackageName and Tag from
// android.media.session.MediaController objects.
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
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/media/session"
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
	ctx, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := session.NewMediaSessionManager(ctx)
	if err != nil {
		return fmt.Errorf("session.NewMediaSessionManager: %w", err)
	}
	defer mgr.Close()

	// Manager provides unexported methods for media session control:
	//   getActiveSessionsRaw(notificationListener)
	//     -- returns a list of active MediaController JNI objects.
	//     Requires MEDIA_CONTENT_CONTROL permission or notification listener access.
	//   addOnActiveSessionsChangedListener(listener, notificationListener)
	//     -- registers a callback for session changes.
	//   removeOnActiveSessionsChangedListener(listener)
	//     -- unregisters a previously added callback.
	//
	// The mediaController data class (unexported) extracts:
	//   PackageName string -- the package owning the media session.
	//   Tag         string -- the tag identifying the session.
	//
	// The onActiveSessionsChangedListener callback (unexported) provides:
	//   OnChanged func(arg0 *jni.Object) -- invoked when active sessions change.

	fmt.Fprintln(output, "MediaSessionManager obtained successfully")
	fmt.Fprintln(output, "Available raw methods: getActiveSessionsRaw, addOnActiveSessionsChangedListener, removeOnActiveSessionsChangedListener")
	fmt.Fprintln(output, "Data class mediaController fields: PackageName, Tag")

	return nil
}

// getAppContext obtains an Android Context via ActivityThread.currentApplication().
func getAppContext(vm *jni.VM) (*app.Context, error) {
	var ctx app.Context
	ctx.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}

		atClass, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread: %w", err)
		}

		curAppMid, err := env.GetStaticMethodID(atClass, "currentApplication", "()Landroid/app/Application;")
		if err != nil {
			return fmt.Errorf("get currentApplication: %w", err)
		}
		appObj, err := env.CallStaticObjectMethod(atClass, curAppMid)
		if err != nil {
			return fmt.Errorf("call currentApplication: %w", err)
		}
		if appObj == nil || appObj.Ref() == 0 {
			return fmt.Errorf("currentApplication returned null")
		}

		ctx.Obj = env.NewGlobalRef(appObj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}
