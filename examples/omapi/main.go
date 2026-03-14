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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/se/omapi"
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
	// --- NewService ---
	svc, err := omapi.NewService(vm)
	if err != nil {
		return fmt.Errorf("omapi.NewService: %w", err)
	}
	defer svc.Close()

	// --- IsConnected ---
	connected := svc.IsConnected()
	fmt.Fprintf(&output, "SE service connected: %v\n", connected)

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

	fmt.Fprintln(&output, "OMAPI example complete.")
	return nil
}
