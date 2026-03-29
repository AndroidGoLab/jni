//go:build android

// Command nsd_service_browser demonstrates NSD (Network Service Discovery).
// It obtains the NsdManager system service, exercises its typed methods,
// and displays all NSD constants including protocol types, failure codes,
// and state values.
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
	"github.com/AndroidGoLab/jni/net/nsd"
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

// newServiceInfo creates a new NsdServiceInfo via its no-arg constructor.
// The wrapper package does not export a constructor, so we use the init-provided
// class to construct one.
func newServiceInfo(vm *jni.VM) (*nsd.ServiceInfo, error) {
	var si nsd.ServiceInfo
	si.VM = vm
	err := vm.Do(func(env *jni.Env) error {
		if err := nsd.Init(env); err != nil {
			return err
		}
		cls, err := env.FindClass("android/net/nsd/NsdServiceInfo")
		if err != nil {
			return fmt.Errorf("FindClass NsdServiceInfo: %w", err)
		}
		mid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("GetMethodID <init>: %w", err)
		}
		obj, err := env.NewObject(cls, mid)
		if err != nil {
			return fmt.Errorf("NewObject: %w", err)
		}
		si.Obj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &si, nil
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== NSD Service Browser ===")

	// 1. Obtain the NsdManager system service.
	nsdMgr, err := nsd.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("nsd.NewManager: %w", err)
	}
	defer nsdMgr.Close()
	fmt.Fprintln(output, "NsdManager: obtained OK")

	// 2. NsdManager.ToString.
	mgrStr, err := nsdMgr.ToString()
	if err != nil {
		fmt.Fprintf(output, "  ToString: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  ToString: %s\n", mgrStr)
	}

	// 3. Show NSD constants.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "NSD constants:")
	fmt.Fprintf(output, "  ProtocolDnsSd: %d\n", nsd.ProtocolDnsSd)
	fmt.Fprintf(output, "  NsdStateEnabled: %d\n", nsd.NsdStateEnabled)
	fmt.Fprintf(output, "  NsdStateDisabled: %d\n", nsd.NsdStateDisabled)
	fmt.Fprintf(output, "  FailureAlreadyActive: %d\n", nsd.FailureAlreadyActive)
	fmt.Fprintf(output, "  FailureBadParameters: %d\n", nsd.FailureBadParameters)
	fmt.Fprintf(output, "  FailureInternalError: %d\n", nsd.FailureInternalError)
	fmt.Fprintf(output, "  FailureMaxLimit: %d\n", nsd.FailureMaxLimit)
	fmt.Fprintf(output, "  FailureOperationNotRunning: %d\n", nsd.FailureOperationNotRunning)

	// 4-19. Create an NsdServiceInfo and exercise all typed setters/getters.
	fmt.Fprintln(output)
	fmt.Fprintln(output, "NsdServiceInfo round-trip test:")

	si, err := newServiceInfo(vm)
	if err != nil {
		fmt.Fprintf(output, "  create: %v\n", err)
		fmt.Fprintln(output)
		fmt.Fprintln(output, "NSD service browser example complete.")
		return nil
	}
	defer func() {
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(si.Obj)
			return nil
		})
	}()

	// 4. SetServiceName.
	if err := si.SetServiceName("GoTestService"); err != nil {
		fmt.Fprintf(output, "  SetServiceName: %v\n", err)
	} else {
		fmt.Fprintln(output, "  SetServiceName('GoTestService'): OK")
	}

	// 5. SetServiceType.
	if err := si.SetServiceType("_http._tcp"); err != nil {
		fmt.Fprintf(output, "  SetServiceType: %v\n", err)
	} else {
		fmt.Fprintln(output, "  SetServiceType('_http._tcp'): OK")
	}

	// 6. SetPort.
	if err := si.SetPort(8080); err != nil {
		fmt.Fprintf(output, "  SetPort: %v\n", err)
	} else {
		fmt.Fprintln(output, "  SetPort(8080): OK")
	}

	// 7. SetAttribute.
	if err := si.SetAttribute("path", "/api/v1"); err != nil {
		fmt.Fprintf(output, "  SetAttribute(path): %v\n", err)
	} else {
		fmt.Fprintln(output, "  SetAttribute('path', '/api/v1'): OK")
	}

	// 8. SetAttribute (second).
	if err := si.SetAttribute("version", "2.0"); err != nil {
		fmt.Fprintf(output, "  SetAttribute(version): %v\n", err)
	} else {
		fmt.Fprintln(output, "  SetAttribute('version', '2.0'): OK")
	}

	// 9. GetServiceName.
	name, err := si.GetServiceName()
	if err != nil {
		fmt.Fprintf(output, "  GetServiceName: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  GetServiceName: %s\n", name)
	}

	// 10. GetServiceType.
	svcType, err := si.GetServiceType()
	if err != nil {
		fmt.Fprintf(output, "  GetServiceType: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  GetServiceType: %s\n", svcType)
	}

	// 11. GetPort.
	port, err := si.GetPort()
	if err != nil {
		fmt.Fprintf(output, "  GetPort: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  GetPort: %d\n", port)
	}

	// 12. GetHost.
	host, err := si.GetHost()
	if err != nil {
		fmt.Fprintf(output, "  GetHost: error: %v\n", err)
	} else if host == nil || host.Ref() == 0 {
		fmt.Fprintln(output, "  GetHost: (null) -- not set")
	} else {
		fmt.Fprintf(output, "  GetHost: obtained (ref=%d)\n", host.Ref())
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(host)
			return nil
		})
	}

	// 13. GetHostname (API 34+).
	hostname, err := si.GetHostname()
	if err != nil {
		fmt.Fprintf(output, "  GetHostname: %v\n", err)
	} else {
		fmt.Fprintf(output, "  GetHostname: %q\n", hostname)
	}

	// 14. GetHostAddresses.
	addrs, err := si.GetHostAddresses()
	if err != nil {
		fmt.Fprintf(output, "  GetHostAddresses: %v\n", err)
	} else if addrs == nil || addrs.Ref() == 0 {
		fmt.Fprintln(output, "  GetHostAddresses: (null)")
	} else {
		fmt.Fprintf(output, "  GetHostAddresses: list obtained\n")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(addrs)
			return nil
		})
	}

	// 15. GetNetwork.
	network, err := si.GetNetwork()
	if err != nil {
		fmt.Fprintf(output, "  GetNetwork: %v\n", err)
	} else if network == nil || network.Ref() == 0 {
		fmt.Fprintln(output, "  GetNetwork: (null) -- not set")
	} else {
		fmt.Fprintln(output, "  GetNetwork: obtained")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(network)
			return nil
		})
	}

	// 16. GetSubtypes (API 34+).
	subtypes, err := si.GetSubtypes()
	if err != nil {
		fmt.Fprintf(output, "  GetSubtypes: %v\n", err)
	} else if subtypes == nil || subtypes.Ref() == 0 {
		fmt.Fprintln(output, "  GetSubtypes: (null)")
	} else {
		fmt.Fprintln(output, "  GetSubtypes: set obtained")
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(subtypes)
			return nil
		})
	}

	// 17. DescribeContents.
	dc, err := si.DescribeContents()
	if err != nil {
		fmt.Fprintf(output, "  DescribeContents: %v\n", err)
	} else {
		fmt.Fprintf(output, "  DescribeContents: %d\n", dc)
	}

	// 18. RemoveAttribute.
	if err := si.RemoveAttribute("version"); err != nil {
		fmt.Fprintf(output, "  RemoveAttribute('version'): %v\n", err)
	} else {
		fmt.Fprintln(output, "  RemoveAttribute('version'): OK")
	}

	// 19. ToString.
	siStr, err := si.ToString()
	if err != nil {
		fmt.Fprintf(output, "  ToString: error: %v\n", err)
	} else {
		fmt.Fprintf(output, "  ToString: %s\n", siStr)
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "NSD service browser example complete.")
	return nil
}
