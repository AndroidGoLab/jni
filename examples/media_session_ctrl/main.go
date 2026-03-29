//go:build android

// Command media_session_ctrl demonstrates the MediaSession API. It creates
// a session, sets it active/inactive, queries session state, and shows
// the session lifecycle. It also queries the MediaSessionManager for the
// current media key event handler.
package main

/*
#include <android/native_activity.h>
extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static void _setCallbacks(ANativeActivity* a) { a->callbacks->onResume = _onResume; a->callbacks->onNativeWindowCreated = _onWindowCreated; }
static uintptr_t _getVM(ANativeActivity* a) { return (uintptr_t)a->vm; }
static uintptr_t _getClazz(ANativeActivity* a) { return (uintptr_t)a->clazz; }
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/media/session"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(C._getVM(activity))),
		jni.ObjectFromUintptr(uintptr(C._getClazz(activity))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(C._getClazz(activity))),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== MediaSession Ctrl Demo ===")
	ui.RenderOutput()

	// --- Create a MediaSession ---
	// NewMediaSession(vm, context, tag)
	sess, err := session.NewMediaSession(vm, ctx.Obj, "JNI_Session_Demo")
	if err != nil {
		fmt.Fprintf(output, "NewMediaSession: %v\n", err)
		fmt.Fprintln(output, "\nSession ctrl example complete (session unavailable).")
		return nil
	}
	if sess == nil || sess.Obj == nil || sess.Obj.Ref() == 0 {
		fmt.Fprintln(output, "MediaSession: null")
		fmt.Fprintln(output, "\nSession ctrl example complete (session null).")
		return nil
	}
	defer func() {
		sess.Release()
		vm.Do(func(env *jni.Env) error {
			if sess.Obj != nil {
				env.DeleteGlobalRef(sess.Obj)
				sess.Obj = nil
			}
			return nil
		})
	}()
	fmt.Fprintln(output, "Session created: tag=JNI_Session_Demo")
	ui.RenderOutput()

	// Query initial active state.
	active, err := sess.IsActive()
	if err != nil {
		fmt.Fprintf(output, "IsActive: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "IsActive (initial): %v\n", active)
	}
	ui.RenderOutput()

	// Set session active.
	if err := sess.SetActive(true); err != nil {
		fmt.Fprintf(output, "SetActive(true): %v\n", err)
	} else {
		fmt.Fprintln(output, "SetActive(true): OK")
	}
	ui.RenderOutput()

	active, err = sess.IsActive()
	if err != nil {
		fmt.Fprintf(output, "IsActive: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "IsActive (after set): %v\n", active)
	}
	ui.RenderOutput()

	// Set media button handling flags.
	if err := sess.SetFlags(int32(session.FlagHandlesMediaButtons | session.FlagHandlesTransportControls)); err != nil {
		fmt.Fprintf(output, "SetFlags: %v\n", err)
	} else {
		fmt.Fprintln(output, "SetFlags: MEDIA_BUTTONS|TRANSPORT")
	}
	ui.RenderOutput()

	// Set queue title.
	if err := sess.SetQueueTitle("Demo Queue"); err != nil {
		fmt.Fprintf(output, "SetQueueTitle: %v\n", err)
	} else {
		fmt.Fprintln(output, "SetQueueTitle: Demo Queue")
	}
	ui.RenderOutput()

	// Get session token.
	token, err := sess.GetSessionToken()
	if err != nil {
		fmt.Fprintf(output, "GetSessionToken: %v\n", err)
	} else if token == nil || token.Ref() == 0 {
		fmt.Fprintln(output, "GetSessionToken: null")
	} else {
		fmt.Fprintln(output, "GetSessionToken: obtained OK")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(token)
			return nil
		})
	}
	ui.RenderOutput()

	// Get controller.
	ctrl, err := sess.GetController()
	if err != nil {
		fmt.Fprintf(output, "GetController: %v\n", err)
	} else if ctrl == nil || ctrl.Ref() == 0 {
		fmt.Fprintln(output, "GetController: null")
	} else {
		fmt.Fprintln(output, "GetController: obtained OK")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(ctrl)
			return nil
		})
	}
	ui.RenderOutput()

	// Set session inactive.
	if err := sess.SetActive(false); err != nil {
		fmt.Fprintf(output, "SetActive(false): %v\n", err)
	} else {
		fmt.Fprintln(output, "SetActive(false): OK")
	}
	ui.RenderOutput()

	active, err = sess.IsActive()
	if err != nil {
		fmt.Fprintf(output, "IsActive: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "IsActive (after deactivate): %v\n", active)
	}
	ui.RenderOutput()

	// --- Query MediaSessionManager ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "--- SessionManager ---")
	mgr, err := session.NewMediaSessionManager(ctx)
	if err != nil {
		fmt.Fprintf(output, "NewMediaSessionManager: %v\n", err)
	} else if mgr == nil || mgr.Obj == nil || mgr.Obj.Ref() == 0 {
		fmt.Fprintln(output, "MediaSessionManager: null")
	} else {
		defer mgr.Close()

		keyPkg, err := mgr.GetMediaKeyEventSessionPackageName()
		if err != nil {
			fmt.Fprintf(output, "MediaKeyEventPkg: %v\n", err)
		} else {
			fmt.Fprintf(output, "Media key handler: %s\n", keyPkg)
		}
	}
	ui.RenderOutput()

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Session ctrl example complete.")
	return nil
}
