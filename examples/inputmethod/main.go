//go:build android

// Command inputmethod demonstrates the InputMethodManager JNI bindings.
// It lists enabled and installed input methods and reports IME state.
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/view/inputmethod"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
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

	fmt.Fprintln(output, "=== Input Methods ===")

	// Filtered: GetEnabledInputMethodList returns generic type (List<InputMethodInfo>)
	// enabledList, err := mgr.GetEnabledInputMethodList()
	// if err != nil {
	// 	fmt.Fprintf(output, "GetEnabledList: %v\n", err)
	// } else {
	// 	printIMEList(vm, output, "Enabled", enabledList)
	// }

	// Filtered: GetInputMethodList returns generic type (List<InputMethodInfo>)
	// allList, err := mgr.GetInputMethodList()
	// if err != nil {
	// 	fmt.Fprintf(output, "GetAllList: %v\n", err)
	// } else {
	// 	printIMEList(vm, output, "Installed", allList)
	// }

	active, err := mgr.IsActive0()
	if err != nil {
		fmt.Fprintf(output, "IsActive: %v\n", err)
	} else {
		fmt.Fprintf(output, "IME active: %v\n", active)
	}

	accepting, err := mgr.IsAcceptingText()
	if err != nil {
		fmt.Fprintf(output, "IsAcceptingText: %v\n", err)
	} else {
		fmt.Fprintf(output, "Accepting text: %v\n", accepting)
	}

	fullscreen, err := mgr.IsFullscreenMode()
	if err != nil {
		fmt.Fprintf(output, "IsFullscreenMode: %v\n", err)
	} else {
		fmt.Fprintf(output, "Fullscreen: %v\n", fullscreen)
	}

	return nil
}

func printIMEList(
	vm *jni.VM,
	output *bytes.Buffer,
	label string,
	listObj *jni.Object,
) {
	if listObj == nil {
		fmt.Fprintf(output, "%s: (null)\n", label)
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
		fmt.Fprintf(output, "%s IMEs: %d\n", label, size)

		objCls, err := env.FindClass("java/lang/Object")
		if err != nil {
			return err
		}
		toStrMid, err := env.GetMethodID(objCls, "toString", "()Ljava/lang/String;")
		if err != nil {
			return err
		}

		for i := int32(0); i < size; i++ {
			elem, err := env.CallObjectMethod(listObj, getMid, jni.IntValue(i))
			if err != nil {
				fmt.Fprintf(output, "  [%d] err: %v\n", i, err)
				continue
			}
			strObj, err := env.CallObjectMethod(elem, toStrMid)
			if err != nil {
				fmt.Fprintf(output, "  [%d] toString: %v\n", i, err)
				continue
			}
			fmt.Fprintf(output, "  [%d] %s\n", i, env.GoString((*jni.String)(unsafe.Pointer(strObj))))
		}
		return nil
	})
}
