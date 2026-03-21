//go:build android

// Command usb demonstrates using the Android UsbManager system service,
// wrapped by the usb package. It enumerates connected USB devices and
// accessories using live JNI calls.
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
	"github.com/AndroidGoLab/jni/hardware/usb"
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

	mgr, err := usb.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("usb.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== USB Manager ===")
	fmt.Fprintln(output, "UsbManager obtained OK")

	// --- Intent Action Constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Intent actions:")
	fmt.Fprintf(output, "  Attached:  %s\n", usb.ActionUsbDeviceAttached)
	fmt.Fprintf(output, "  Detached:  %s\n", usb.ActionUsbDeviceDetached)
	fmt.Fprintf(output, "  Accessory: %s\n", usb.ActionUsbAccessoryAttached)

	// --- Enumerate USB devices via raw JNI ---
	// The generated bindings don't include getDeviceList, so we call it
	// directly through JNI. It returns HashMap<String, UsbDevice>.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "USB Devices:")

	var deviceCount int
	var deviceErr error
	vm.Do(func(env *jni.Env) error {
		mgrCls := env.GetObjectClass(mgr.Obj)

		// UsbManager.getDeviceList() -> HashMap<String, UsbDevice>
		getDeviceListMid, err := env.GetMethodID(mgrCls, "getDeviceList",
			"()Ljava/util/HashMap;")
		if err != nil {
			deviceErr = fmt.Errorf("getDeviceList method: %w", err)
			return nil
		}

		deviceMap, err := env.CallObjectMethod(mgr.Obj, getDeviceListMid)
		if err != nil {
			deviceErr = fmt.Errorf("getDeviceList call: %w", err)
			return nil
		}
		if deviceMap == nil || deviceMap.Ref() == 0 {
			fmt.Fprintln(output, "  (no device map)")
			return nil
		}

		// HashMap.size()
		mapCls := env.GetObjectClass(deviceMap)
		sizeMid, err := env.GetMethodID(mapCls, "size", "()I")
		if err != nil {
			deviceErr = fmt.Errorf("HashMap.size: %w", err)
			return nil
		}
		size, err := env.CallIntMethod(deviceMap, sizeMid)
		if err != nil {
			deviceErr = fmt.Errorf("size call: %w", err)
			return nil
		}
		deviceCount = int(size)
		fmt.Fprintf(output, "  Count: %d\n", deviceCount)

		if deviceCount == 0 {
			return nil
		}

		// HashMap.values() -> Collection
		valuesMid, err := env.GetMethodID(mapCls, "values",
			"()Ljava/util/Collection;")
		if err != nil {
			return nil
		}
		valuesObj, err := env.CallObjectMethod(deviceMap, valuesMid)
		if err != nil || valuesObj == nil {
			return nil
		}

		// Collection.toArray() -> Object[]
		colCls := env.GetObjectClass(valuesObj)
		toArrayMid, err := env.GetMethodID(colCls, "toArray",
			"()[Ljava/lang/Object;")
		if err != nil {
			return nil
		}
		arrayObj, err := env.CallObjectMethod(valuesObj, toArrayMid)
		if err != nil || arrayObj == nil {
			return nil
		}

		arr := (*jni.ObjectArray)(unsafe.Pointer(arrayObj))
		arrLen := env.GetArrayLength((*jni.Array)(unsafe.Pointer(arr)))

		for i := int32(0); i < arrLen; i++ {
			devObj, err := env.GetObjectArrayElement(arr, i)
			if err != nil || devObj == nil {
				continue
			}

			// Wrap in usb.Device to call exported methods
			gRef := env.NewGlobalRef(devObj)
			dev := usb.Device{VM: vm, Obj: gRef}

			name, _ := dev.GetDeviceName0()
			vid, _ := dev.GetVendorId()
			pid, _ := dev.GetProductId()
			cls, _ := dev.GetDeviceClass()
			mfr, _ := dev.GetManufacturerName()
			prod, _ := dev.GetProductName()
			ifCount, _ := dev.GetInterfaceCount()

			fmt.Fprintf(output, "  [%d] %s\n", i, name)
			fmt.Fprintf(output, "    VID:0x%04X PID:0x%04X\n", vid, pid)
			fmt.Fprintf(output, "    Class:%d Mfr:%q\n", cls, mfr)
			fmt.Fprintf(output, "    Product:%q\n", prod)
			fmt.Fprintf(output, "    Interfaces: %d\n", ifCount)

			// Show interface details
			for j := int32(0); j < ifCount; j++ {
				ifObj, err := dev.GetInterface(j)
				if err != nil || ifObj == nil {
					continue
				}
				iface := usb.Interface{VM: vm, Obj: env.NewGlobalRef(ifObj)}
				ifID, _ := iface.GetId()
				ifClass, _ := iface.GetInterfaceClass()
				ifSub, _ := iface.GetInterfaceSubclass()
				epCount, _ := iface.GetEndpointCount()
				fmt.Fprintf(output, "      IF%d: cls=%d sub=%d eps=%d\n",
					ifID, ifClass, ifSub, epCount)
			}
		}

		return nil
	})

	if deviceErr != nil {
		fmt.Fprintf(output, "  Error: %v\n", deviceErr)
	}

	// --- Enumerate USB accessories ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "USB Accessories:")
	accList, err := mgr.GetAccessoryList()
	if err != nil {
		fmt.Fprintf(output, "  Error: %v\n", err)
	} else if accList == nil || accList.Ref() == 0 {
		fmt.Fprintln(output, "  (none)")
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
				accCls := env.GetObjectClass(accObj)
				toStrMid, err := env.GetMethodID(accCls, "toString", "()Ljava/lang/String;")
				if err != nil {
					continue
				}
				strObj, err := env.CallObjectMethod(accObj, toStrMid)
				if err != nil {
					continue
				}
				s := env.GoString((*jni.String)(unsafe.Pointer(strObj)))
				fmt.Fprintf(output, "  [%d] %s\n", i, s)
			}
			return nil
		})
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "USB example complete.")
	return nil
}
