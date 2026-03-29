//go:build android

// Command usb_enumerator demonstrates the UsbManager system service.
// It shows USB class constants and endpoint type constants using
// the typed wrapper. Since the UsbManager.getDeviceList() method
// returns a HashMap which is not directly wrapped, we show the
// accessory list and USB constants available via typed wrappers.
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

	fmt.Fprintln(output, "=== USB Device Enumerator ===")
	fmt.Fprintln(output, "UsbManager: obtained")

	// --- USB device class constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "USB device class constants:")
	fmt.Fprintf(output, "  PER_INTERFACE     = %d\n", usb.UsbClassPerInterface)
	fmt.Fprintf(output, "  AUDIO             = %d\n", usb.UsbClassAudio)
	fmt.Fprintf(output, "  COMM              = %d\n", usb.UsbClassComm)
	fmt.Fprintf(output, "  HID               = %d\n", usb.UsbClassHid)
	fmt.Fprintf(output, "  PHYSICAL          = %d\n", usb.UsbClassPhysica)
	fmt.Fprintf(output, "  STILL_IMAGE       = %d\n", usb.UsbClassStillImage)
	fmt.Fprintf(output, "  PRINTER           = %d\n", usb.UsbClassPrinter)
	fmt.Fprintf(output, "  MASS_STORAGE      = %d\n", usb.UsbClassMassStorage)
	fmt.Fprintf(output, "  HUB               = %d\n", usb.UsbClassHub)
	fmt.Fprintf(output, "  CDC_DATA          = %d\n", usb.UsbClassCdcData)
	fmt.Fprintf(output, "  CSCID             = %d\n", usb.UsbClassCscid)
	fmt.Fprintf(output, "  CONTENT_SEC       = %d\n", usb.UsbClassContentSec)
	fmt.Fprintf(output, "  VIDEO             = %d\n", usb.UsbClassVideo)
	fmt.Fprintf(output, "  WIRELESS_CTRL     = %d\n", usb.UsbClassWirelessController)
	fmt.Fprintf(output, "  MISC              = %d\n", usb.UsbClassMisc)
	fmt.Fprintf(output, "  APP_SPEC          = %d\n", usb.UsbClassAppSpec)
	fmt.Fprintf(output, "  VENDOR_SPEC       = %d\n", usb.UsbClassVendorSpec)

	// --- USB endpoint transfer type constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Endpoint transfer types:")
	fmt.Fprintf(output, "  CONTROL     = %d\n", usb.UsbEndpointXferControl)
	fmt.Fprintf(output, "  ISOCHRONOUS = %d\n", usb.UsbEndpointXferIsoc)
	fmt.Fprintf(output, "  BULK        = %d\n", usb.UsbEndpointXferBulk)
	fmt.Fprintf(output, "  INTERRUPT   = %d\n", usb.UsbEndpointXferInt)

	// --- USB direction constants ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Direction constants:")
	fmt.Fprintf(output, "  USB_DIR_IN  = %d\n", usb.UsbDirIn)
	fmt.Fprintf(output, "  USB_DIR_OUT = %d\n", usb.UsbDirOut)

	// --- Check accessory list ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "=== USB Accessories ===")
	accList, err := mgr.GetAccessoryList()
	if err != nil {
		fmt.Fprintf(output, "GetAccessoryList: %v\n", err)
	} else if accList == nil || accList.Ref() == 0 {
		fmt.Fprintln(output, "No USB accessories connected")
	} else {
		fmt.Fprintln(output, "USB accessory list obtained")
	}

	fmt.Fprintln(output, "\nUSB device enumeration complete.")
	return nil
}
