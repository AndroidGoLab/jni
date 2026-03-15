//go:build android

// Command app_framework demonstrates the core Android application framework
// types: Context, Activity, Intent, Bundle, and PendingIntent.
//
// The app package wraps android.content.Context, android.app.Activity,
// android.content.Intent, android.app.PendingIntent, and android.os.Bundle.
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
	// --- Context ---
	// Context wraps android.content.Context and is the central type
	// through which system services and app resources are accessed.
	ctx, err := app.NewContext(vm)
	if err != nil {
		return fmt.Errorf("app.NewContext: %w", err)
	}
	defer ctx.Close()

	pkgName, err := ctx.PackageName()
	if err != nil {
		fmt.Fprintf(&output, "  PackageName: %v\n", err)
	} else {
		fmt.Fprintf(&output, "package: %s\n", pkgName)
	}

	appCtx, err := ctx.ApplicationContext()
	if err != nil {
		fmt.Fprintf(&output, "  ApplicationContext: %v\n", err)
	} else {
		fmt.Fprintf(&output, "application context: %v\n", appCtx)
	}

	// System service lookup (used internally by alarm, audiomanager, etc.).
	svc, err := ctx.GetSystemService("alarm")
	if err != nil {
		fmt.Fprintf(&output, "  GetSystemService: %v\n", err)
	} else {
		fmt.Fprintf(&output, "alarm service: %v\n", svc)
	}

	// Directory paths.
	cacheDir, err := ctx.CacheDir()
	if err != nil {
		fmt.Fprintf(&output, "  CacheDir: %v\n", err)
	} else {
		fmt.Fprintf(&output, "cache dir: %v\n", cacheDir)
	}

	filesDir, err := ctx.FilesDir()
	if err != nil {
		fmt.Fprintf(&output, "  FilesDir: %v\n", err)
	} else {
		fmt.Fprintf(&output, "files dir: %v\n", filesDir)
	}

	externalFilesDir, err := ctx.ExternalFilesDir("")
	if err != nil {
		fmt.Fprintf(&output, "  ExternalFilesDir: %v\n", err)
	} else {
		fmt.Fprintf(&output, "external files dir: %v\n", externalFilesDir)
	}

	externalCacheDir, err := ctx.ExternalCacheDir()
	if err != nil {
		fmt.Fprintf(&output, "  ExternalCacheDir: %v\n", err)
	} else {
		fmt.Fprintf(&output, "external cache dir: %v\n", externalCacheDir)
	}

	dataDir, err := ctx.DataDir()
	if err != nil {
		fmt.Fprintf(&output, "  DataDir: %v\n", err)
	} else {
		fmt.Fprintf(&output, "data dir: %v\n", dataDir)
	}

	// Content resolver and package manager.
	resolver := ctx.ContentResolver()
	fmt.Fprintf(&output, "content resolver: %v\n", resolver)

	pkgMgr, err := ctx.PackageManager()
	if err != nil {
		fmt.Fprintf(&output, "  PackageManager: %v\n", err)
	} else {
		fmt.Fprintf(&output, "package manager: %v\n", pkgMgr)
	}

	// Broadcast and service operations.
	if err := ctx.SendBroadcast(nil); err != nil {
		fmt.Fprintf(&output, "  SendBroadcast: %v\n", err)
	}
	if _, err := ctx.StartService(nil); err != nil {
		fmt.Fprintf(&output, "  StartService: %v\n", err)
	}

	// Receiver registration and unregistration.
	if _, err := ctx.RegisterReceiverRaw(nil, nil); err != nil {
		fmt.Fprintf(&output, "  RegisterReceiverRaw: %v\n", err)
	}
	if err := ctx.UnregisterReceiver(nil); err != nil {
		fmt.Fprintf(&output, "  UnregisterReceiver: %v\n", err)
	}

	// --- Intent ---
	// Intent wraps android.content.Intent for launching activities,
	// sending broadcasts, and starting services.
	intent, err := app.NewIntent(vm)
	if err != nil {
		return fmt.Errorf("app.NewIntent: %w", err)
	}

	intent.SetAction(app.ActionView)
	intent.SetFlags(app.FlagNewTask)
	intent.AddFlags(app.FlagClearTop)
	intent.SetPackage("com.example.app")
	intent.SetType("text/plain")
	intent.AddCategory("android.intent.category.DEFAULT")
	intent.PutStringExtra("key", "value")
	intent.PutIntExtra("count", 42)
	intent.PutBoolExtra("enabled", true)
	intent.PutLongExtra("timestamp", 1709942400000)

	action := intent.GetAction()
	fmt.Fprintf(&output, "intent action: %s\n", action)
	data := intent.GetData()
	fmt.Fprintf(&output, "intent data: %v\n", data)
	extra := intent.GetStringExtra("key")
	fmt.Fprintf(&output, "string extra: %s\n", extra)
	intExtra := intent.GetIntExtra("count", 0)
	fmt.Fprintf(&output, "int extra: %d\n", intExtra)
	boolExtra := intent.GetBoolExtra("enabled", false)
	fmt.Fprintf(&output, "bool extra: %v\n", boolExtra)

	// Start an activity with the intent.
	if err := ctx.StartActivity(intent.Obj); err != nil {
		fmt.Fprintf(&output, "  StartActivity: %v\n", err)
	}

	// --- Activity ---
	// Activity wraps android.app.Activity for UI lifecycle management.
	activity, err := app.NewActivity(vm)
	if err != nil {
		fmt.Fprintf(&output, "  NewActivity: %v\n", err)
	} else {
		actIntent, err := activity.GetIntent()
		if err != nil {
			fmt.Fprintf(&output, "  GetIntent: %v\n", err)
		} else {
			fmt.Fprintf(&output, "activity intent: %v\n", actIntent)
		}

		window, err := activity.GetWindow()
		if err != nil {
			fmt.Fprintf(&output, "  GetWindow: %v\n", err)
		} else {
			fmt.Fprintf(&output, "window: %v\n", window)
		}

		activity.SetResult(0, nil)

		if err := activity.RunOnUiThread(nil); err != nil {
			fmt.Fprintf(&output, "  RunOnUiThread: %v\n", err)
		}

		if err := activity.StartActivityForResult(nil, 1); err != nil {
			fmt.Fprintf(&output, "  StartActivityForResult: %v\n", err)
		}

		if err := activity.Finish(); err != nil {
			fmt.Fprintf(&output, "  Finish: %v\n", err)
		}
	}

	// --- PendingIntent ---
	// PendingIntent wraps android.app.PendingIntent for deferred intent delivery.
	var pi app.PendingIntent
	pi.VM = vm

	if _, err := pi.NewPendingActivity(nil, 0, nil, app.PendingFlagImmutable); err != nil {
		fmt.Fprintf(&output, "  NewPendingActivity: %v\n", err)
	}

	if _, err := pi.NewPendingBroadcast(nil, 0, nil, app.PendingFlagMutable); err != nil {
		fmt.Fprintf(&output, "  NewPendingBroadcast: %v\n", err)
	}

	if _, err := pi.NewPendingService(nil, 0, nil, app.PendingFlagUpdateCurrent); err != nil {
		fmt.Fprintf(&output, "  NewPendingService: %v\n", err)
	}

	// --- Bundle ---
	// Bundle is a data class wrapping android.os.Bundle.
	var _ app.Bundle

	// Intent action constants.
	fmt.Fprintf(&output, "intent actions: VIEW=%q, SEND=%q, DIAL=%q, PICK=%q\n",
		app.ActionView, app.ActionSend, app.ActionDial, app.ActionPick)

	// Intent flag constants.
	fmt.Fprintf(&output, "intent flags: NEW_TASK=0x%x, CLEAR_TOP=0x%x, GRANT_READ_URI=0x%x\n",
		app.FlagNewTask, app.FlagClearTop, app.FlagGrantReadURI)

	// PendingIntent flag constants.
	fmt.Fprintf(&output, "pending intent flags: IMMUTABLE=0x%x, MUTABLE=0x%x, UPDATE_CURRENT=0x%x\n",
		app.PendingFlagImmutable, app.PendingFlagMutable, app.PendingFlagUpdateCurrent)

	return nil
}
