//go:build android

// Command session demonstrates using the Android MediaSession API
// to query active media sessions and the media key event handler.
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/media/session"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
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

	mgr, err := session.NewMediaSessionManager(ctx)
	if err != nil {
		return fmt.Errorf("NewMediaSessionManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== MediaSessionManager ===")

	// GetActiveSessions requires a NotificationListenerService
	// ComponentName (null means only sessions owned by this app).
	sessions, err := mgr.GetActiveSessions(nil)
	if err != nil {
		fmt.Fprintf(output, "GetActiveSessions: %v\n", err)
	} else {
		printMediaSessions(vm, output, sessions)
	}

	// Query media key event session package name (API 28+).
	keyPkg, err := mgr.GetMediaKeyEventSessionPackageName()
	if err != nil {
		fmt.Fprintf(output, "MediaKeyEventPkg: %v\n", err)
	} else {
		fmt.Fprintf(output, "Media key handler: %s\n", keyPkg)
	}

	// Query Session2 tokens (API 28+).
	tokens, err := mgr.GetSession2Tokens()
	if err != nil {
		fmt.Fprintf(output, "Session2Tokens: %v\n", err)
	} else {
		var tokenCount int32
		_ = vm.Do(func(env *jni.Env) error {
			if tokens == nil {
				return nil
			}
			listCls, err := env.FindClass("java/util/List")
			if err != nil {
				return err
			}
			sizeMid, err := env.GetMethodID(listCls, "size", "()I")
			if err != nil {
				return err
			}
			tokenCount, err = env.CallIntMethod(tokens, sizeMid)
			return err
		})
		fmt.Fprintf(output, "Session2 tokens: %d\n", tokenCount)
	}

	return nil
}

func printMediaSessions(
	vm *jni.VM,
	output *bytes.Buffer,
	listObj *jni.Object,
) {
	if listObj == nil {
		fmt.Fprintln(output, "Active sessions: (null)")
		return
	}

	_ = vm.Do(func(env *jni.Env) error {
		listCls, err := env.FindClass("java/util/List")
		if err != nil {
			return err
		}
		sizeMid, err := env.GetMethodID(listCls, "size", "()I")
		if err != nil {
			return err
		}
		getMid, err := env.GetMethodID(listCls, "get", "(I)Ljava/lang/Object;")
		if err != nil {
			return err
		}

		size, err := env.CallIntMethod(listObj, sizeMid)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "Active sessions: %d\n", size)

		for i := int32(0); i < size; i++ {
			elem, err := env.CallObjectMethod(listObj, getMid, jni.IntValue(i))
			if err != nil {
				fmt.Fprintf(output, "  [%d] error: %v\n", i, err)
				continue
			}

			// Wrap as MediaController to use typed methods.
			ctrl := session.MediaController{
				VM:  vm,
				Obj: env.NewGlobalRef(elem),
			}

			pkg, _ := ctrl.GetPackageName()
			tag, _ := ctrl.GetTag()
			fmt.Fprintf(output, "  [%d] pkg=%s tag=%s\n", i, pkg, tag)
		}
		return nil
	})
}
