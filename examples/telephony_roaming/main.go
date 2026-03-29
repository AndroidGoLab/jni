//go:build android

// Command telephony_roaming checks whether the device is currently
// roaming and displays network country ISO and operator info.
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

	fmt.Fprintln(output, "=== Roaming Status ===")
	fmt.Fprintln(output)

	roaming, err := mgr.IsNetworkRoaming()
	if err != nil {
		return fmt.Errorf("isNetworkRoaming: %w", err)
	}
	fmt.Fprintf(output, "roaming: %v\n", roaming)

	countryIso, err := mgr.GetNetworkCountryIso0()
	if err != nil {
		return fmt.Errorf("getNetworkCountryIso: %w", err)
	}
	fmt.Fprintf(output, "network country: %s\n", countryIso)

	operatorName, err := mgr.GetNetworkOperatorName()
	if err != nil {
		return fmt.Errorf("getNetworkOperatorName: %w", err)
	}
	fmt.Fprintf(output, "operator name: %s\n", operatorName)

	operator, err := mgr.GetNetworkOperator()
	if err != nil {
		return fmt.Errorf("getNetworkOperator: %w", err)
	}
	fmt.Fprintf(output, "operator code: %s\n", operator)

	simCountryIso, err := mgr.GetSimCountryIso()
	if err != nil {
		return fmt.Errorf("getSimCountryIso: %w", err)
	}
	fmt.Fprintf(output, "SIM country: %s\n", simCountryIso)

	dataRoaming, err := mgr.IsDataRoamingEnabled()
	if err != nil {
		fmt.Fprintf(output, "data roaming: (unavail)\n")
	} else {
		fmt.Fprintf(output, "data roaming: %v\n", dataRoaming)
	}

	return nil
}
