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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/app"
	"github.com/xaionaro-go/jni/net/wifi/p2p"
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

	mgr, err := p2p.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("p2p.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(&output, "WifiP2pManager obtained successfully")

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

	// --- Device Status Constants ---
	fmt.Fprintf(&output, "StatusConnected:   %d\n", p2p.StatusConnected)
	fmt.Fprintf(&output, "StatusInvited:     %d\n", p2p.StatusInvited)
	fmt.Fprintf(&output, "StatusFailed:      %d\n", p2p.StatusFailed)
	fmt.Fprintf(&output, "StatusAvailable:   %d\n", p2p.StatusAvailable)
	fmt.Fprintf(&output, "StatusUnavailable: %d\n", p2p.StatusUnavailable)

	// --- Device Data Class ---
	// Device holds data from android.net.wifi.p2p.WifiP2pDevice:
	var dev p2p.Device
	fmt.Fprintf(&output, "Device.Name:    %q\n", dev.Name)
	fmt.Fprintf(&output, "Device.Address: %q\n", dev.Address)
	fmt.Fprintf(&output, "Device.Status:  %d\n", dev.Status)

	// --- Group Data Class ---
	// Group holds data from android.net.wifi.p2p.WifiP2pGroup:
	var grp p2p.Group
	fmt.Fprintf(&output, "Group.NetworkName: %q\n", grp.NetworkName)
	fmt.Fprintf(&output, "Group.Passphrase:  %q\n", grp.Passphrase)
	fmt.Fprintf(&output, "Group.IsOwner:     %v\n", grp.IsOwner)

	// --- Callback Types (all unexported) ---
	// actionListener: OnSuccess, OnFailure(reason int32).
	// peerListListener: OnPeersAvailable(peerList *jni.Object).
	// connectionInfoListener: OnConnectionInfoAvailable(info *jni.Object).
	// channelListener: OnChannelDisconnected.
	//
	// p2pConfig (unexported): wraps WifiP2pConfig, created via Newp2pConfig.

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
