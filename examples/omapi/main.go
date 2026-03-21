//go:build android

// Command omapi demonstrates using the Open Mobile API (OMAPI).
//
// This example creates an SEService, checks connectivity, and shows
// how the SE reader/session/channel hierarchy works. The
// onConnectedListener callback pattern is also described. Most
// lower-level methods are unexported (designed to be called from within
// the omapi package or via higher-level wrappers); exported methods
// are called directly.
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
	"github.com/AndroidGoLab/jni/se/omapi"
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
	// --- NewService ---
	svc, err := omapi.NewService(vm)
	if err != nil {
		return fmt.Errorf("omapi.NewService: %w", err)
	}
	defer svc.Close()

	// --- IsConnected ---
	connected, _ := svc.IsConnected()
	fmt.Fprintf(output, "SE service connected: %v\n", connected)

	// --- Shutdown ---
	defer svc.Shutdown()

	// --- onConnectedListener callback (unexported) ---
	// Registered via registeronConnectedListener to be notified when
	// the SE service becomes connected:
	//
	//   onConnectedListener{
	//     OnConnected func()
	//   }
	//   proxy, cleanup, err := registeronConnectedListener(env, cb)

	// --- Reader (exported methods) ---
	// Readers are obtained from svc.getReadersRaw() (unexported).
	// Each Reader exposes:
	//
	//   reader.GetName() string                          [exported]
	//   reader.IsSecureElementPresent() (bool, error)    [exported]
	//   reader.openSessionRaw() (*jni.Object, error)     [unexported]

	// --- Session (exported methods) ---
	// Sessions are opened from a Reader:
	//
	//   session.Close()                                  [exported]
	//   session.GetATR() *jni.Object                     [exported]
	//   session.openBasicChannelRaw(aid) (*jni.Object)   [unexported]
	//   session.openLogicalChannelRaw(aid) (*jni.Object) [unexported]
	//   session.closeRaw()                               [unexported]

	// --- Channel (exported methods) ---
	// Channels are opened from a Session:
	//
	//   channel.Close()                                  [exported]
	//   channel.Transmit(cmd) (*jni.Object, error)       [exported]
	//   channel.SelectNext() (bool, error)               [exported]
	//   channel.GetSelectResponse() *jni.Object          [exported]
	//   channel.IsBasicChannel() bool                    [exported]
	//   channel.closeRaw()                               [unexported]

	fmt.Fprintln(output, "OMAPI example complete.")
	return nil
}
