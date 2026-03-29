//go:build android

// Command telephony_data_monitor checks the mobile data connection
// state and data activity direction via TelephonyManager.
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
	"github.com/AndroidGoLab/jni/telephony"
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

func dataStateName(s int32) string {
	switch int(s) {
	case telephony.DataDisconnected:
		return "DISCONNECTED"
	case telephony.DataConnecting:
		return "CONNECTING"
	case telephony.DataConnected:
		return "CONNECTED"
	case telephony.DataSuspended:
		return "SUSPENDED"
	case telephony.DataDisconnecting:
		return "DISCONNECTING"
	case telephony.DataHandoverInProgress:
		return "HANDOVER"
	case telephony.DataUnknown:
		return "UNKNOWN"
	default:
		return fmt.Sprintf("(%d)", s)
	}
}

func dataActivityName(a int32) string {
	switch int(a) {
	case telephony.DataActivityNone:
		return "NONE"
	case telephony.DataActivityIn:
		return "IN"
	case telephony.DataActivityOut:
		return "OUT"
	case telephony.DataActivityInout:
		return "INOUT"
	case telephony.DataActivityDormant:
		return "DORMANT"
	default:
		return fmt.Sprintf("(%d)", a)
	}
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := telephony.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("telephony.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Data Monitor ===")
	fmt.Fprintln(output)

	dataState, err := mgr.GetDataState()
	if err != nil {
		return fmt.Errorf("getDataState: %w", err)
	}
	fmt.Fprintf(output, "data state: %s\n", dataStateName(dataState))

	dataActivity, err := mgr.GetDataActivity()
	if err != nil {
		return fmt.Errorf("getDataActivity: %w", err)
	}
	fmt.Fprintf(output, "data activity: %s\n", dataActivityName(dataActivity))

	dataEnabled, err := mgr.IsDataEnabled()
	if err != nil {
		fmt.Fprintf(output, "data enabled: (unavail)\n")
	} else {
		fmt.Fprintf(output, "data enabled: %v\n", dataEnabled)
	}

	dataCapable, err := mgr.IsDataCapable()
	if err != nil {
		fmt.Fprintf(output, "data capable: (unavail)\n")
	} else {
		fmt.Fprintf(output, "data capable: %v\n", dataCapable)
	}

	dataConnAllowed, err := mgr.IsDataConnectionAllowed()
	if err != nil {
		fmt.Fprintf(output, "data conn allowed: (unavail)\n")
	} else {
		fmt.Fprintf(output, "data conn allowed: %v\n", dataConnAllowed)
	}

	return nil
}
