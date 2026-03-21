//go:build android

// Command companion demonstrates using the Android CompanionDeviceManager API.
// It is built as a c-shared library and packaged into an APK using the
// shared apk.mk infrastructure.
//
// This example obtains the CompanionDeviceManager system service using the
// exported NewManager constructor. The companion package provides unexported
// methods for device association (associateRaw, disassociateByIdRaw,
// getAssociationsRaw), an unexported associationRequestBuilder type, and a
// companionCallback type with OnDeviceFound and OnFailure callbacks.
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
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/companion"
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

	// NewManager obtains the CompanionDeviceManager system service.
	mgr, err := companion.NewDeviceManager(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "service not available") {
			fmt.Fprintln(output, "CompanionDeviceManager not available on this device")
			fmt.Fprintln(output, "")
			fmt.Fprintln(output, "Package companion provides the following API surface:")
			fmt.Fprintln(output, "  Manager type (wraps android.companion.CompanionDeviceManager)")
			fmt.Fprintln(output, "    - associateRaw(request, callback, handler) error")
			fmt.Fprintln(output, "    - disassociateByIdRaw(associationId int32) error")
			fmt.Fprintln(output, "    - getAssociationsRaw() (*jni.Object, error)")
			fmt.Fprintln(output, "  associationRequestBuilder type (wraps AssociationRequest.Builder)")
			fmt.Fprintln(output, "    - setSingleDevice(bool) *jni.Object")
			fmt.Fprintln(output, "    - addDeviceFilter(filter) *jni.Object")
			fmt.Fprintln(output, "    - build() *jni.Object")
			fmt.Fprintln(output, "  companionCallback (OnDeviceFound, OnFailure)")
			return nil
		}
		return fmt.Errorf("companion.NewDeviceManager: %v", err)
	}

	fmt.Fprintln(output, "CompanionDeviceManager obtained successfully")
	_ = mgr

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
