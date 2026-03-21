//go:build android

// Command job demonstrates the JobScheduler JNI bindings. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the JobScheduler system service and demonstrates
// the exported Cancel and CancelAll methods along with all available
// constants. The jobInfoBuilder type (created via NewjobInfoBuilder)
// provides package-internal methods for configuring job constraints.
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
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/app/job"
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
	ctx, err := getAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	sched, err := job.NewScheduler(ctx)
	if err != nil {
		return fmt.Errorf("job.NewScheduler: %v", err)
	}
	defer sched.Close()

	// Print all result code constants.
	fmt.Fprintln(output, "Result codes:")
	fmt.Fprintf(output, "  ResultSuccess = %d\n", job.ResultSuccess)
	fmt.Fprintf(output, "  ResultFailure = %d\n", job.ResultFailure)

	// Print all network type constants.
	fmt.Fprintln(output, "Network types:")
	fmt.Fprintf(output, "  NetworkTypeNone       = %d\n", job.NetworkTypeNone)
	fmt.Fprintf(output, "  NetworkTypeAny        = %d\n", job.NetworkTypeAny)
	fmt.Fprintf(output, "  NetworkTypeUnmetered  = %d\n", job.NetworkTypeUnmetered)
	fmt.Fprintf(output, "  NetworkTypeNotRoaming = %d\n", job.NetworkTypeNotRoaming)
	fmt.Fprintf(output, "  NetworkTypeCellular   = %d\n", job.NetworkTypeCellular)

	// Print backoff policy constants.
	fmt.Fprintln(output, "Backoff policies:")
	fmt.Fprintf(output, "  BackoffPolicyLinear      = %d\n", job.BackoffPolicyLinear)
	fmt.Fprintf(output, "  BackoffPolicyExponential = %d\n", job.BackoffPolicyExponential)

	// CancelInAllNamespaces cancels all jobs in all namespaces.
	if err := sched.CancelInAllNamespaces(); err != nil {
		fmt.Fprintf(output, "CancelInAllNamespaces: %v (expected on older APIs)\n", err)
	} else {
		fmt.Fprintln(output, "cancelled all jobs in all namespaces")
	}

	fmt.Fprintln(output, "JobScheduler example complete")
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
