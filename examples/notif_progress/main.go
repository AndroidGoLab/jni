//go:build android

// Command notif_progress posts a notification with a progress bar,
// updates it from 0 to 100%, then marks it complete.
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
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app/notification"
	"github.com/AndroidGoLab/jni/examples/common/ui"
)

const (
	// android.R.drawable.ic_dialog_info
	iconDialogInfo = 17301620

	channelID   = "progress_demo"
	channelName = "Progress Demo"
	notifID     = 42
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

	mgr, err := notification.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("notification.NewManager: %w", err)
	}
	defer mgr.Close()

	fmt.Fprintln(output, "=== Progress Notif ===")
	fmt.Fprintln(output)

	// Create channel.
	ch, err := notification.NewChannel(
		vm, channelID, channelName,
		int32(notification.ImportanceLow),
	)
	if err != nil {
		return fmt.Errorf("new channel: %w", err)
	}
	if err := mgr.CreateNotificationChannel(ch.Obj); err != nil {
		return fmt.Errorf("create channel: %w", err)
	}
	fmt.Fprintln(output, "channel created")

	appCtxObj, err := ctx.GetApplicationContext()
	if err != nil {
		return fmt.Errorf("get app context: %w", err)
	}

	// Post progress updates from 0% to 100%.
	for pct := int32(0); pct <= 100; pct += 10 {
		b, err := notification.NewBuilder(vm, appCtxObj, channelID)
		if err != nil {
			return fmt.Errorf("new builder: %w", err)
		}
		if _, err := b.SetSmallIcon1_1(iconDialogInfo); err != nil {
			return fmt.Errorf("set icon: %w", err)
		}
		if _, err := b.SetContentTitle("Downloading..."); err != nil {
			return fmt.Errorf("set title: %w", err)
		}
		if _, err := b.SetContentText(fmt.Sprintf("%d%% complete", pct)); err != nil {
			return fmt.Errorf("set text: %w", err)
		}
		if _, err := b.SetProgress(100, pct, false); err != nil {
			return fmt.Errorf("set progress: %w", err)
		}
		if _, err := b.SetOngoing(true); err != nil {
			return fmt.Errorf("set ongoing: %w", err)
		}
		if _, err := b.SetOnlyAlertOnce(true); err != nil {
			return fmt.Errorf("set alert once: %w", err)
		}
		notif, err := b.Build()
		if err != nil {
			return fmt.Errorf("build: %w", err)
		}
		if err := mgr.Notify2(notifID, notif); err != nil {
			return fmt.Errorf("notify: %w", err)
		}
		fmt.Fprintf(output, "progress: %d%%\n", pct)

		if pct < 100 {
			time.Sleep(300 * time.Millisecond)
		}
	}

	// Final notification: complete.
	b, err := notification.NewBuilder(vm, appCtxObj, channelID)
	if err != nil {
		return fmt.Errorf("new builder: %w", err)
	}
	if _, err := b.SetSmallIcon1_1(iconDialogInfo); err != nil {
		return fmt.Errorf("set icon: %w", err)
	}
	if _, err := b.SetContentTitle("Download complete"); err != nil {
		return fmt.Errorf("set title: %w", err)
	}
	if _, err := b.SetContentText("File downloaded successfully"); err != nil {
		return fmt.Errorf("set text: %w", err)
	}
	if _, err := b.SetProgress(0, 0, false); err != nil {
		return fmt.Errorf("set progress: %w", err)
	}
	if _, err := b.SetOngoing(false); err != nil {
		return fmt.Errorf("set ongoing: %w", err)
	}
	if _, err := b.SetAutoCancel(true); err != nil {
		return fmt.Errorf("set auto cancel: %w", err)
	}
	notif, err := b.Build()
	if err != nil {
		return fmt.Errorf("build final: %w", err)
	}
	if err := mgr.Notify2(notifID, notif); err != nil {
		return fmt.Errorf("notify final: %w", err)
	}
	fmt.Fprintln(output, "download complete!")

	return nil
}
