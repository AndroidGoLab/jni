//go:build android

// Command wifi_direct_transfer demonstrates WiFi P2P setup: creates a P2P
// manager, initializes a channel, and queries P2P support and capability
// information.
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
	"github.com/AndroidGoLab/jni/net/wifi"
	"github.com/AndroidGoLab/jni/net/wifi/p2p"
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

	fmt.Fprintln(output, "=== WiFi Direct (P2P) ===")

	// Check if P2P is supported via the WifiManager.
	wifiMgr, err := wifi.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("wifi.NewManager: %w", err)
	}
	defer wifiMgr.Close()

	p2pSupported, err := wifiMgr.IsP2pSupported()
	if err != nil {
		fmt.Fprintf(output, "IsP2pSupported error: %v\n", err)
	} else {
		fmt.Fprintf(output, "P2P supported: %v\n", p2pSupported)
	}

	// Create the WifiP2pManager system service.
	p2pMgr, err := p2p.NewWifiP2pManager(ctx)
	if err != nil {
		return fmt.Errorf("p2p.NewWifiP2pManager: %w", err)
	}
	defer p2pMgr.Close()
	fmt.Fprintln(output, "WifiP2pManager obtained")

	// Query P2P capability flags.
	chanDiscovery, err := p2pMgr.IsChannelConstrainedDiscoverySupported()
	if err != nil {
		fmt.Fprintf(output, "IsChannelConstrainedDiscoverySupported: %v\n", err)
	} else {
		fmt.Fprintf(output, "Channel-constrained discovery: %v\n", chanDiscovery)
	}

	groupRemoval, err := p2pMgr.IsGroupClientRemovalSupported()
	if err != nil {
		fmt.Fprintf(output, "IsGroupClientRemovalSupported: %v\n", err)
	} else {
		fmt.Fprintf(output, "Group client removal: %v\n", groupRemoval)
	}

	goIPv6, err := p2pMgr.IsGroupOwnerIPv6LinkLocalAddressProvided()
	if err != nil {
		fmt.Fprintf(output, "IsGroupOwnerIPv6LinkLocalAddressProvided: %v\n", err)
	} else {
		fmt.Fprintf(output, "GO IPv6 link-local: %v\n", goIPv6)
	}

	vendorElem, err := p2pMgr.IsSetVendorElementsSupported()
	if err != nil {
		fmt.Fprintf(output, "IsSetVendorElementsSupported: %v\n", err)
	} else {
		fmt.Fprintf(output, "Vendor elements: %v\n", vendorElem)
	}

	wifiDirectR2, err := p2pMgr.IsWiFiDirectR2Supported()
	if err != nil {
		fmt.Fprintf(output, "IsWiFiDirectR2Supported: %v\n", err)
	} else {
		fmt.Fprintf(output, "WiFi Direct R2: %v\n", wifiDirectR2)
	}

	maxVendorLen, err := p2pMgr.GetP2pMaxAllowedVendorElementsLengthBytes()
	if err != nil {
		fmt.Fprintf(output, "GetP2pMaxAllowedVendorElementsLengthBytes: %v\n", err)
	} else {
		fmt.Fprintf(output, "Max vendor elements length: %d bytes\n", maxVendorLen)
	}

	// Initialize a P2P channel using the activity context.
	// The Initialize method requires (Context, Looper, ChannelListener).
	// Passing nil for Looper uses the main looper, and nil for the listener.
	chanObj, err := p2pMgr.Initialize(ctx.Obj, nil, nil)
	if err != nil {
		fmt.Fprintf(output, "Initialize error: %v\n", err)
	} else if chanObj != nil && chanObj.Ref() != 0 {
		fmt.Fprintln(output, "P2P Channel initialized successfully")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(chanObj)
			return nil
		})
	} else {
		fmt.Fprintln(output, "P2P Channel: nil (may need Looper)")
	}

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "WiFi Direct example complete.")
	return nil
}
