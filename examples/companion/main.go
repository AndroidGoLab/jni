//go:build android

// Command companion demonstrates using the Android CompanionDeviceManager API.
// It checks availability and lists existing device associations.
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
	"strings"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/companion"
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
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== CompanionDeviceManager ===")

	mgr, err := companion.NewDeviceManager(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "service not available") {
			fmt.Fprintln(output, "Status: NOT AVAILABLE")
			fmt.Fprintln(output, "(CompanionDeviceManager requires API 26+)")
			return nil
		}
		return fmt.Errorf("companion.NewDeviceManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "Status: available")

	// Query existing associations (returns java.util.List).
	assocList, err := mgr.GetAssociations()
	if err != nil {
		fmt.Fprintf(output, "GetAssociations: %v\n", err)
	} else {
		var listSize int32
		_ = vm.Do(func(env *jni.Env) error {
			if assocList == nil {
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
			listSize, err = env.CallIntMethod(assocList, sizeMid)
			return err
		})
		fmt.Fprintf(output, "Associations: %d\n", listSize)

		// Print each association's toString().
		if listSize > 0 {
			_ = vm.Do(func(env *jni.Env) error {
				listCls, err := env.FindClass("java/util/List")
				if err != nil {
					return err
				}
				getMid, err := env.GetMethodID(listCls, "get", "(I)Ljava/lang/Object;")
				if err != nil {
					return err
				}
				objCls, err := env.FindClass("java/lang/Object")
				if err != nil {
					return err
				}
				toStrMid, err := env.GetMethodID(objCls, "toString", "()Ljava/lang/String;")
				if err != nil {
					return err
				}

				for i := int32(0); i < listSize; i++ {
					elem, err := env.CallObjectMethod(assocList, getMid, jni.IntValue(i))
					if err != nil {
						continue
					}
					strObj, err := env.CallObjectMethod(elem, toStrMid)
					if err != nil {
						continue
					}
					fmt.Fprintf(output, "  [%d] %s\n", i, env.GoString((*jni.String)(unsafe.Pointer(strObj))))
				}
				return nil
			})
		}
	}

	// Try getMyAssociations (API 33+).
	myAssoc, err := mgr.GetMyAssociations()
	if err != nil {
		fmt.Fprintf(output, "GetMyAssociations: %v\n", err)
	} else {
		var myCount int32
		_ = vm.Do(func(env *jni.Env) error {
			if myAssoc == nil {
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
			myCount, err = env.CallIntMethod(myAssoc, sizeMid)
			return err
		})
		fmt.Fprintf(output, "My associations: %d\n", myCount)
	}

	return nil
}
