//go:build android

// Command print demonstrates using the PrintManager API. It is built as a
// c-shared library and packaged into an APK using the shared apk.mk
// infrastructure.
//
// This example obtains the PrintManager system service and shows all
// print job state constants. The Manager provides methods for querying
// print jobs and starting new ones. The printJob and printJobInfo data
// classes (unexported) hold job metadata extracted from JNI objects.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/print"
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

	// --- Constants ---
	// JobState is a typed constant representing print job states.
	fmt.Fprintln(&output, "Print job state constants (type JobState):")
	fmt.Fprintf(&output, "  JobCreated   = %d\n", print.StateCreated)
	fmt.Fprintf(&output, "  JobQueued    = %d\n", print.StateQueued)
	fmt.Fprintf(&output, "  JobStarted   = %d\n", print.StateStarted)
	fmt.Fprintf(&output, "  JobBlocked   = %d\n", print.StateBlocked)
	fmt.Fprintf(&output, "  JobCompleted = %d\n", print.StateCompleted)
	fmt.Fprintf(&output, "  JobFailed    = %d\n", print.StateFailed)
	fmt.Fprintf(&output, "  JobCanceled  = %d\n", print.StateCanceled)

	// --- NewManager ---
	mgr, err := print.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("print.NewManager: %v", err)
	}
	_ = mgr

	// --- Manager methods (unexported) ---
	//
	//   mgr.getPrintJobsRaw() (*jni.Object, error)
	//     Query all print jobs visible to this application.
	//
	//   mgr.printRaw(jobName string, adapter, attributes *jni.Object)
	//     Start a new print job with a PrintDocumentAdapter.

	// --- Data classes (unexported) ---
	//
	// printJobInfo{
	//   Label string
	//   State JobState
	// }
	// Extracted from android.print.PrintJobInfo via extractprintJobInfo.
	//
	// printJob{
	//   Info printJobInfo
	// }
	// Extracted from android.print.PrintJob via extractprintJob.

	// --- JobState typed constant ---
	// JobState values can be used in switch statements.
	state := print.StateCompleted
	switch state {
	case print.StateCreated:
		fmt.Fprintln(&output, "job: created")
	case print.StateQueued:
		fmt.Fprintln(&output, "job: queued")
	case print.StateStarted:
		fmt.Fprintln(&output, "job: started")
	case print.StateBlocked:
		fmt.Fprintln(&output, "job: blocked")
	case print.StateCompleted:
		fmt.Fprintln(&output, "job: completed")
	case print.StateFailed:
		fmt.Fprintln(&output, "job: failed")
	case print.StateCanceled:
		fmt.Fprintln(&output, "job: canceled")
	}
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
