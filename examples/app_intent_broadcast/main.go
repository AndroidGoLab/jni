//go:build android

// Command app_intent_broadcast creates an Intent with action and extras,
// sends it as a broadcast, and shows Intent building and broadcasting
// via Context.
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
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/examples/common/ui"
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

	fmt.Fprintln(output, "=== Intent Broadcast ===")

	// Create an Intent.
	intent, err := app.NewIntent(vm, nil, nil)
	if err != nil {
		return fmt.Errorf("app.NewIntent: %w", err)
	}

	// Build a custom broadcast action.
	customAction := "center.dx.jni.examples.ACTION_TEST_BROADCAST"
	if _, err := intent.SetAction(customAction); err != nil {
		return fmt.Errorf("SetAction: %w", err)
	}

	action, err := intent.GetAction()
	if err != nil {
		fmt.Fprintf(output, "GetAction: %v\n", err)
	} else {
		fmt.Fprintf(output, "Action: %s\n", action)
	}

	// Add extras.
	if err := intent.PutStringExtra("message", "Hello from Go!"); err != nil {
		fmt.Fprintf(output, "PutStringExtra: %v\n", err)
	} else {
		fmt.Fprintln(output, "Added string extra: message=Hello from Go!")
	}

	if err := intent.PutIntExtra("timestamp", 1234567890); err != nil {
		fmt.Fprintf(output, "PutIntExtra: %v\n", err)
	} else {
		fmt.Fprintln(output, "Added int extra: timestamp=1234567890")
	}

	if err := intent.PutBoolExtra("from_go", true); err != nil {
		fmt.Fprintf(output, "PutBoolExtra: %v\n", err)
	} else {
		fmt.Fprintln(output, "Added bool extra: from_go=true")
	}

	// Verify extras can be read back.
	msgExtra, err := intent.GetStringExtra("message")
	if err != nil {
		fmt.Fprintf(output, "GetStringExtra: %v\n", err)
	} else {
		fmt.Fprintf(output, "Read back string: %s\n", msgExtra)
	}

	tsExtra, err := intent.GetIntExtra("timestamp", 0)
	if err != nil {
		fmt.Fprintf(output, "GetIntExtra: %v\n", err)
	} else {
		fmt.Fprintf(output, "Read back int: %d\n", tsExtra)
	}

	// Send the broadcast.
	fmt.Fprintln(output, "\nSending broadcast...")
	err = ctx.SendBroadcast1(intent.Obj)
	if err != nil {
		fmt.Fprintf(output, "SendBroadcast: %v\n", err)
	} else {
		fmt.Fprintln(output, "Broadcast sent OK")
	}

	// Show Intent action constants.
	fmt.Fprintln(output, "\nIntent Action Constants:")
	fmt.Fprintf(output, "  ACTION_VIEW: %s\n", app.ActionView)
	fmt.Fprintf(output, "  ACTION_SEND: %s\n", app.ActionSend)
	fmt.Fprintf(output, "  ACTION_MAIN: %s\n", app.ActionMain)

	// Show Intent flag constants.
	fmt.Fprintln(output, "\nIntent Flag Constants:")
	fmt.Fprintf(output, "  FLAG_ACTIVITY_NEW_TASK:   0x%x\n", app.FlagActivityNewTask)
	fmt.Fprintf(output, "  FLAG_ACTIVITY_CLEAR_TOP:  0x%x\n", app.FlagActivityClearTop)

	fmt.Fprintln(output, "\nIntent broadcast complete.")
	return nil
}
