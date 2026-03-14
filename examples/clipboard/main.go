//go:build android

// Command clipboard demonstrates using the Android ClipboardManager API.
// It is built as a c-shared library and packaged into an APK using the
// shared apk.mk infrastructure.
//
// This example obtains the ClipboardManager system service and shows how the
// exported NewManager constructor and Close cleanup method are used.
// The clipboard package also provides unexported methods for clipboard
// operations (setPrimaryClip, getPrimaryClip, hasPrimaryClip, addListener,
// removeListener) and a clipChangedListener callback type with OnChanged,
// as well as unexported clipData and clipItem helper types.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/app"
	"github.com/xaionaro-go/jni/content/clipboard"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	vm := jni.VMFromPtr(unsafe.Pointer(cvm))
	if err := run(vm); err != nil {
		fmt.Fprintf(&output, "ERROR: %v\n", err)
	}
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}

func run(vm *jni.VM) error {
	ctx, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// NewManager obtains the ClipboardManager system service.
	mgr, err := clipboard.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("clipboard.NewManager: %v", err)
	}
	// Close releases the JNI global reference; always defer it.
	defer mgr.Close()

	fmt.Fprintln(&output, "ClipboardManager obtained successfully")

	// The following methods are unexported (package-internal) and are
	// used by higher-level Go wrappers that build on this package:
	//
	//   mgr.setPrimaryClip(clip *jni.Object) error
	//     Sets the current primary clip on the clipboard.
	//
	//   mgr.getPrimaryClip() (*jni.Object, error)
	//     Returns the current primary clip on the clipboard.
	//
	//   mgr.hasPrimaryClip() (bool, error)
	//     Returns true if there is currently a primary clip on the clipboard.
	//
	//   mgr.addListener(listener *jni.Object) error
	//     Registers a listener to be called when the primary clip changes.
	//
	//   mgr.removeListener(listener *jni.Object) error
	//     Removes a previously registered clip-changed listener.
	//
	// The clipChangedListener struct provides a Go-friendly callback interface:
	//
	//   clipChangedListener{
	//       OnChanged: func() { ... },
	//   }
	//
	// It is registered via the unexported registerclipChangedListener function,
	// which creates a Java proxy implementing
	// ClipboardManager.OnPrimaryClipChangedListener.
	//
	// Additional unexported helper types:
	//
	//   clipData wraps android.content.ClipData with methods:
	//     clipData.getItemAt(index int32) *jni.Object
	//
	//   clipItem wraps android.content.ClipData.Item with methods:
	//     clipItem.getText() *jni.Object

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
