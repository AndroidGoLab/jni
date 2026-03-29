//go:build android

// Command telephony_call_state reads the current call state
// (idle/ringing/offhook) from TelephonyManager.
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

func callStateName(s int32) string {
	switch int(s) {
	case telephony.CallStateIdle:
		return "IDLE"
	case telephony.CallStateRinging:
		return "RINGING"
	case telephony.CallStateOffhook:
		return "OFFHOOK"
	default:
		return fmt.Sprintf("(%d)", s)
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

	fmt.Fprintln(output, "=== Call State ===")
	fmt.Fprintln(output)

	callState, err := mgr.GetCallState()
	if err != nil {
		return fmt.Errorf("getCallState: %w", err)
	}
	fmt.Fprintf(output, "call state: %s\n", callStateName(callState))

	// Also try the subscription-aware variant.
	callStateSub, err := mgr.GetCallStateForSubscription()
	if err != nil {
		fmt.Fprintf(output, "sub call state: (unavail)\n")
	} else {
		fmt.Fprintf(output, "sub call state: %s\n", callStateName(callStateSub))
	}

	voiceCapable, err := mgr.IsVoiceCapable()
	if err != nil {
		fmt.Fprintf(output, "voice capable: (unavail)\n")
	} else {
		fmt.Fprintf(output, "voice capable: %v\n", voiceCapable)
	}

	smsCapable, err := mgr.IsSmsCapable()
	if err != nil {
		fmt.Fprintf(output, "SMS capable: (unavail)\n")
	} else {
		fmt.Fprintf(output, "SMS capable: %v\n", smsCapable)
	}

	concurrentVD, err := mgr.IsConcurrentVoiceAndDataSupported()
	if err != nil {
		fmt.Fprintf(output, "voice+data: (unavail)\n")
	} else {
		fmt.Fprintf(output, "voice+data: %v\n", concurrentVD)
	}

	return nil
}
