//go:build android

// Command app_framework demonstrates the core Android application framework
// types: Context, Activity, Intent, and PendingIntent.
//
// The app package wraps android.content.Context, android.app.Activity,
// android.content.Intent, and android.app.PendingIntent.
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
	"github.com/AndroidGoLab/jni/app"
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
	// --- Context ---
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// Context provides system service lookup.
	svc, err := ctx.GetSystemService("alarm")
	if err != nil {
		fmt.Fprintf(output, "  GetSystemService(alarm): %v\n", err)
	} else {
		fmt.Fprintf(output, "alarm service: %v\n", svc)
	}

	// Content resolver and package manager.
	resolver, err := ctx.GetContentResolver()
	if err != nil {
		fmt.Fprintf(output, "  ContentResolver: %v\n", err)
	} else {
		fmt.Fprintf(output, "content resolver: %v\n", resolver != nil)
	}

	pkgMgr, err := ctx.GetPackageManager()
	if err != nil {
		fmt.Fprintf(output, "  PackageManager: %v\n", err)
	} else {
		fmt.Fprintf(output, "package manager: %v\n", pkgMgr != nil)
	}

	// --- Intent ---
	intent, err := app.NewIntent(vm)
	if err != nil {
		return fmt.Errorf("app.NewIntent: %w", err)
	}

	// Intent methods: SetAction, GetAction, SetFlags, AddFlags, etc.
	if _, err := intent.SetAction(app.ActionView); err != nil {
		return fmt.Errorf("set action: %w", err)
	}
	action, err := intent.GetAction()
	if err != nil {
		fmt.Fprintf(output, "  GetAction: %v\n", err)
	} else {
		fmt.Fprintf(output, "intent action: %s\n", action)
	}

	intent.SetFlags(int32(app.FlagActivityNewTask))
	intent.AddFlags(int32(app.FlagActivityClearTop))

	// Intent extras.
	if err := intent.PutStringExtra("key", "value"); err != nil {
		fmt.Fprintf(output, "  PutStringExtra: %v\n", err)
	}
	if err := intent.PutIntExtra("count", 42); err != nil {
		fmt.Fprintf(output, "  PutIntExtra: %v\n", err)
	}
	if err := intent.PutBoolExtra("enabled", true); err != nil {
		fmt.Fprintf(output, "  PutBoolExtra: %v\n", err)
	}

	extra, err := intent.GetStringExtra("key")
	if err != nil {
		fmt.Fprintf(output, "  GetStringExtra: %v\n", err)
	} else {
		fmt.Fprintf(output, "string extra: %s\n", extra)
	}

	intExtra, err := intent.GetIntExtra("count", 0)
	if err != nil {
		fmt.Fprintf(output, "  GetIntExtra: %v\n", err)
	} else {
		fmt.Fprintf(output, "int extra: %d\n", intExtra)
	}

	// Intent action constants.
	fmt.Fprintf(output, "intent actions: VIEW=%q, SEND=%q\n",
		app.ActionView, app.ActionSend)

	// Intent flag constants.
	fmt.Fprintf(output, "intent flags: NEW_TASK=0x%x, CLEAR_TOP=0x%x, GRANT_READ_URI=0x%x\n",
		app.FlagActivityNewTask, app.FlagActivityClearTop, app.FlagGrantReadUriPermission)

	return nil
}
