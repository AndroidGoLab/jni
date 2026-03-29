//go:build android

// Command inputmethod_keyboard demonstrates the InputMethodManager API.
// It obtains the IME manager, reports keyboard state (active, accepting
// text, fullscreen), toggles the soft keyboard, and lists enabled input
// methods using the typed InputMethodInfo wrapper.
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
	"github.com/AndroidGoLab/jni/view/inputmethod"
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

	mgr, err := inputmethod.NewInputMethodManager(ctx)
	if err != nil {
		return fmt.Errorf("NewInputMethodManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Input Method Keyboard ===")

	// --- IME state ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "IME state:")

	active, err := mgr.IsActive0()
	if err != nil {
		fmt.Fprintf(output, "  isActive: %v\n", err)
	} else {
		fmt.Fprintf(output, "  active: %v\n", active)
	}

	accepting, err := mgr.IsAcceptingText()
	if err != nil {
		fmt.Fprintf(output, "  isAcceptingText: %v\n", err)
	} else {
		fmt.Fprintf(output, "  accepting text: %v\n", accepting)
	}

	fullscreen, err := mgr.IsFullscreenMode()
	if err != nil {
		fmt.Fprintf(output, "  isFullscreenMode: %v\n", err)
	} else {
		fmt.Fprintf(output, "  fullscreen: %v\n", fullscreen)
	}

	// --- Show/hide constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Show/hide constants:")
	fmt.Fprintf(output, "  SHOW_IMPLICIT = %d\n", inputmethod.ShowImplicit)
	fmt.Fprintf(output, "  SHOW_FORCED   = %d\n", inputmethod.ShowForced)
	fmt.Fprintf(output, "  HIDE_IMPLICIT = %d\n", inputmethod.HideImplicitOnly)
	fmt.Fprintf(output, "  HIDE_NOT_ALWAYS = %d\n", inputmethod.HideNotAlways)

	// --- Result constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Result constants:")
	fmt.Fprintf(output, "  RESULT_SHOWN    = %d\n", inputmethod.ResultShown)
	fmt.Fprintf(output, "  RESULT_HIDDEN   = %d\n", inputmethod.ResultHidden)
	fmt.Fprintf(output, "  RESULT_UNCHANGED_SHOWN  = %d\n", inputmethod.ResultUnchangedShown)
	fmt.Fprintf(output, "  RESULT_UNCHANGED_HIDDEN = %d\n", inputmethod.ResultUnchangedHidden)

	// --- Toggle soft keyboard ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Toggle Keyboard ===")
	if err := mgr.ToggleSoftInput(int32(inputmethod.ShowImplicit), 0); err != nil {
		fmt.Fprintf(output, "toggleSoftInput: %v\n", err)
	} else {
		fmt.Fprintln(output, "toggleSoftInput called")
	}

	// Check state again after toggle.
	active2, err := mgr.IsActive0()
	if err != nil {
		fmt.Fprintf(output, "isActive after toggle: %v\n", err)
	} else {
		fmt.Fprintf(output, "active after toggle: %v\n", active2)
	}

	// --- Current input method info ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Current Input Method ===")
	curIME, err := mgr.GetCurrentInputMethodInfo()
	if err != nil {
		fmt.Fprintf(output, "GetCurrentInputMethodInfo: %v\n", err)
	} else if curIME == nil || curIME.Ref() == 0 {
		fmt.Fprintln(output, "current IME: (null)")
	} else {
		imi := inputmethod.InputMethodInfo{VM: vm, Obj: curIME}
		id, err := imi.GetId()
		if err != nil {
			fmt.Fprintf(output, "current IME id: %v\n", err)
		} else {
			fmt.Fprintf(output, "current IME id: %s\n", id)
		}
		pkgName, err := imi.GetPackageName()
		if err != nil {
			fmt.Fprintf(output, "current IME package: %v\n", err)
		} else {
			fmt.Fprintf(output, "current IME package: %s\n", pkgName)
		}
		svcName, err := imi.GetServiceName()
		if err != nil {
			fmt.Fprintf(output, "current IME service: %v\n", err)
		} else {
			fmt.Fprintf(output, "current IME service: %s\n", svcName)
		}
		subtypeCount, err := imi.GetSubtypeCount()
		if err != nil {
			fmt.Fprintf(output, "subtypeCount: %v\n", err)
		} else {
			fmt.Fprintf(output, "subtypeCount: %d\n", subtypeCount)
		}
	}

	// --- Enabled input methods list ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== Enabled Input Methods ===")
	listObj, err := mgr.GetEnabledInputMethodList()
	if err != nil {
		fmt.Fprintf(output, "getEnabledInputMethodList: %v\n", err)
	} else if listObj == nil || listObj.Ref() == 0 {
		fmt.Fprintln(output, "(null)")
	} else {
		fmt.Fprintln(output, "enabled input method list: obtained")
	}

	fmt.Fprintln(output, "\ninputmethod_keyboard complete")
	return nil
}
