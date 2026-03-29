//go:build android

// Command usb_accessory demonstrates the Android USB accessory API.
// It uses UsbManager to get the accessory list, inspects each accessory's
// metadata, and shows the permission/open workflow.
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
	"github.com/AndroidGoLab/jni/hardware/usb"
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

	mgr, err := usb.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("usb.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== USB Accessory ===")
	fmt.Fprintln(output, "")

	// Intent actions for USB accessories
	fmt.Fprintln(output, "USB Accessory Intent Actions:")
	fmt.Fprintf(output, "  Attached: %s\n", usb.ActionUsbAccessoryAttached)
	fmt.Fprintf(output, "  Detached: %s\n", usb.ActionUsbAccessoryDetached)
	fmt.Fprintln(output, "")

	// Get accessory list
	fmt.Fprintln(output, "Connected USB Accessories:")
	accList, err := mgr.GetAccessoryList()
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else if accList == nil || accList.Ref() == 0 {
		fmt.Fprintln(output, "  (none connected)")
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "  USB accessories are external hardware that communicates")
		fmt.Fprintln(output, "  with Android in accessory mode. The accessory (not the")
		fmt.Fprintln(output, "  Android device) acts as the USB host.")
	} else {
		vm.Do(func(env *jni.Env) error {
			arr := (*jni.ObjectArray)(unsafe.Pointer(accList))
			n := env.GetArrayLength((*jni.Array)(unsafe.Pointer(arr)))
			fmt.Fprintf(output, "  Count: %d\n", n)

			for i := int32(0); i < n; i++ {
				accObj, err := env.GetObjectArrayElement(arr, i)
				if err != nil || accObj == nil {
					continue
				}

				gRef := env.NewGlobalRef(accObj)
				acc := usb.Accessory{VM: vm, Obj: gRef}

				mfr, _ := acc.GetManufacturer()
				model, _ := acc.GetModel()
				desc, _ := acc.GetDescription()
				ver, _ := acc.GetVersion()
				uri, _ := acc.GetUri()
				serial, _ := acc.GetSerial()

				fmt.Fprintf(output, "\n  Accessory [%d]:\n", i)
				fmt.Fprintf(output, "    Manufacturer: %s\n", mfr)
				fmt.Fprintf(output, "    Model:        %s\n", model)
				fmt.Fprintf(output, "    Description:  %s\n", desc)
				fmt.Fprintf(output, "    Version:      %s\n", ver)
				fmt.Fprintf(output, "    URI:          %s\n", uri)
				fmt.Fprintf(output, "    Serial:       %s\n", serial)

				// Check permission
				hasPerm, err := mgr.HasPermission1_1((*jni.Object)(unsafe.Pointer(gRef)))
				if err != nil {
					fmt.Fprintf(output, "    Permission:   error: %v\n", err)
				} else {
					fmt.Fprintf(output, "    Permission:   %v\n", hasPerm)
				}

				env.DeleteGlobalRef(gRef)
			}
			return nil
		})
	}

	// Describe the accessory workflow
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "--- USB Accessory Workflow ---")
	fmt.Fprintln(output, "  1. Register BroadcastReceiver for ACTION_USB_ACCESSORY_ATTACHED")
	fmt.Fprintln(output, "  2. mgr.GetAccessoryList() to find connected accessories")
	fmt.Fprintln(output, "  3. mgr.HasPermission1_1(accessory) to check permission")
	fmt.Fprintln(output, "  4. mgr.RequestPermission2_1(accessory, pendingIntent) if needed")
	fmt.Fprintln(output, "  5. mgr.OpenAccessory(accessory) to get ParcelFileDescriptor")
	fmt.Fprintln(output, "  6. Read/write using the file descriptor streams")
	fmt.Fprintln(output, "  7. mgr.OpenAccessoryInputStream / OpenAccessoryOutputStream")

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "USB accessory example complete.")
	return nil
}
