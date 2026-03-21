//go:build android

// Command usb demonstrates using the Android UsbManager system service,
// wrapped by the usb package. It is built as a c-shared library and
// packaged into an APK using the shared apk.mk infrastructure.
//
// The usb package wraps android.hardware.usb.UsbManager and related
// classes (UsbDeviceConnection, UsbDevice, UsbInterface, UsbEndpoint).
// It provides the Manager for enumerating devices and opening connections,
// the Connection for data transfer, and direction/transfer type constants.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/hardware/usb"
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

	mgr, err := usb.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("usb.NewManager: %w", err)
	}
	defer mgr.Close()

	// Manager provides unexported methods for USB device access:
	//   getDeviceList()                        -- returns connected USB devices.
	//   hasPermission(device)                  -- checks if app has permission.
	//   requestPermission(device, pendingIntent) -- requests USB access.
	//   openDevice(device)                     -- opens a device connection.

	fmt.Fprintln(&output, "UsbManager obtained successfully")

	// DeviceConnection wraps android.hardware.usb.UsbDeviceConnection,
	// obtained via the Manager's openDevice method.
	//
	// DeviceConnection exported methods:
	//   BulkTransfer4, BulkTransfer5_1, ClaimInterface,
	//   Close, ControlTransfer7, ControlTransfer8_1,
	//   GetFileDescriptor, GetRawDescriptors, ReleaseInterface,
	//   RequestWait, SetConfiguration, SetInterface

	// --- USB Intent Action Constants ---
	fmt.Fprintf(&output, "ActionUsbDeviceAttached:    %q\n", usb.ActionUsbDeviceAttached)
	fmt.Fprintf(&output, "ActionUsbDeviceDetached:    %q\n", usb.ActionUsbDeviceDetached)
	fmt.Fprintf(&output, "ActionUsbAccessoryAttached: %q\n", usb.ActionUsbAccessoryAttached)

	// --- Data Classes (all unexported) ---
	// usbDevice: Name, VendorID, ProductID, DeviceID, DeviceClass,
	//   DeviceSubclass, DeviceProtocol, ManufacturerName, ProductName,
	//   SerialNumber, interfaceCount.
	// usbInterface: ID, Class, Subclass, Protocol, endpointCount.
	// usbEndpoint: Address, Direction, Type, MaxPacket.

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
