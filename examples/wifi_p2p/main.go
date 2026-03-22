//go:build android

// Command wifi_p2p demonstrates using the Android Wi-Fi P2P (Wi-Fi Direct)
// API. It obtains the WifiP2pManager, initializes a channel, and queries
// device capability flags.
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
	"github.com/AndroidGoLab/jni/net/wifi/p2p"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(activity.vm)),
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
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

	mgr, err := p2p.NewWifiP2pManager(ctx)
	if err != nil {
		return fmt.Errorf("WifiP2pManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "WifiP2pManager OK")

	// Print P2P state constants for reference.
	fmt.Fprintf(output, "P2P enabled const:  %d\n", p2p.WifiP2pStateEnabled)
	fmt.Fprintf(output, "P2P disabled const: %d\n", p2p.WifiP2pStateDisabled)

	// Initialize a P2P channel.
	// initialize(context, looper, channelListener)
	// We pass the app context and nil for looper (uses main looper) and listener.
	channel, err := mgr.Initialize(ctx.Obj, nil, nil)
	if err != nil {
		fmt.Fprintf(output, "Initialize: %v\n", err)
	} else if channel == nil || channel.Ref() == 0 {
		fmt.Fprintln(output, "Initialize returned nil channel")
	} else {
		fmt.Fprintln(output, "P2P channel initialized")

		// Request P2P state via the channel.
		// requestP2pState(channel, listener) -- requires a listener callback.
		// Since we cannot easily create a Java callback from Go without a proxy,
		// we just confirm the channel is valid by printing its toString.
		vm.Do(func(env *jni.Env) error {
			cls := env.GetObjectClass(channel)
			mid, mErr := env.GetMethodID(cls, "toString", "()Ljava/lang/String;")
			if mErr != nil {
				return nil
			}
			strObj, cErr := env.CallObjectMethod(channel, mid)
			if cErr != nil {
				return nil
			}
			s := env.GoString((*jni.String)(unsafe.Pointer(strObj)))
			fmt.Fprintf(output, "Channel: %s\n", s)
			return nil
		})

		// Clean up channel.
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(channel)
			return nil
		})
	}

	// Query capability flags on the manager.
	printBool := func(name string, fn func() (bool, error)) {
		val, err := fn()
		if err != nil {
			fmt.Fprintf(output, "%s: %v\n", name, err)
		} else {
			fmt.Fprintf(output, "%s: %v\n", name, val)
		}
	}

	printBool("ChannelConstrainedDiscovery", mgr.IsChannelConstrainedDiscoverySupported)
	printBool("GroupClientRemoval", mgr.IsGroupClientRemovalSupported)
	printBool("GroupOwnerIPv6LinkLocal", mgr.IsGroupOwnerIPv6LinkLocalAddressProvided)
	printBool("SetVendorElements", mgr.IsSetVendorElementsSupported)

	maxVE, err := mgr.GetP2pMaxAllowedVendorElementsLengthBytes()
	if err != nil {
		fmt.Fprintf(output, "MaxVendorElementsLen: %v\n", err)
	} else {
		fmt.Fprintf(output, "MaxVendorElementsLen: %d\n", maxVE)
	}

	fmt.Fprintln(output, "\nBand constants:")
	fmt.Fprintf(output, "  Auto: %d\n", p2p.GroupOwnerBandAuto)
	fmt.Fprintf(output, "  2GHz: %d\n", p2p.GroupOwnerBand2ghz)
	fmt.Fprintf(output, "  5GHz: %d\n", p2p.GroupOwnerBand5ghz)
	fmt.Fprintf(output, "  6GHz: %d\n", p2p.GroupOwnerBand6ghz)

	return nil
}
