//go:build android

// Command wifi_p2p demonstrates using the Android Wi-Fi P2P (Wi-Fi Direct)
// API, wrapped by the wifi_p2p package. It is built as a c-shared library
// and packaged into an APK using the shared apk.mk infrastructure.
//
// The wifi_p2p package wraps android.net.wifi.p2p.WifiP2pManager and
// provides data classes for P2P devices and groups, status constants,
// and callback types for asynchronous P2P operations. It requires
// ACCESS_FINE_LOCATION and NEARBY_WIFI_DEVICES permissions.
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
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/net/wifi/p2p"
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

	mgr, err := p2p.NewWifiP2pManager(ctx)
	if err != nil {
		return fmt.Errorf("p2p.NewWifiP2pManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "WifiP2pManager obtained successfully")

	// Manager provides unexported methods for P2P operations:
	//   initializeRaw(ctx, looper, listener)         -- initializes a P2P channel.
	//   discoverPeersRaw(channel, listener)           -- starts peer discovery.
	//   stopPeerDiscoveryRaw(channel, listener)       -- stops peer discovery.
	//   connectRaw(channel, config, listener)         -- connects to a peer.
	//   cancelConnectRaw(channel, listener)           -- cancels a pending connection.
	//   createGroupRaw(channel, listener)             -- creates a P2P group.
	//   removeGroupRaw(channel, listener)             -- removes the current group.
	//   requestConnectionInfoRaw(channel, listener)   -- requests connection info.
	//   requestPeersRaw(channel, listener)            -- requests the peer list.

	// --- P2P Constants ---
	fmt.Fprintf(output, "GroupOwnerBandAuto: %d\n", p2p.GroupOwnerBandAuto)
	fmt.Fprintf(output, "GroupOwnerBand2ghz: %d\n", p2p.GroupOwnerBand2ghz)
	fmt.Fprintf(output, "GroupOwnerBand5ghz: %d\n", p2p.GroupOwnerBand5ghz)

	// --- Callback Types (all unexported) ---
	// actionListener: OnSuccess, OnFailure(reason int32).
	// peerListListener: OnPeersAvailable(peerList *jni.Object).
	// connectionInfoListener: OnConnectionInfoAvailable(info *jni.Object).
	// channelListener: OnChannelDisconnected.
	//
	// p2pConfig (unexported): wraps WifiP2pConfig, created via Newp2pConfig.

	return nil
}
