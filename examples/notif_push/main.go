//go:build android

// Command notif_push creates a notification manager, channel, builds
// and posts a push notification — all from Go via JNI.
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
	"github.com/AndroidGoLab/jni/app/notification"
	"github.com/AndroidGoLab/jni/examples/common/ui"
)

const (
	// android.R.drawable.ic_dialog_info
	iconDialogInfo = 17301620

	channelID   = "push_demo"
	channelName = "Push Demo"
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
	fmt.Fprintln(output, "=== Push Notification ===")
	fmt.Fprintln(output)

	enabled, err := mgr.AreNotificationsEnabled()
	if err != nil {
		fmt.Fprintf(output, "notifications enabled: (unavail)\n")
	} else {
		fmt.Fprintf(output, "notifications enabled: %v\n", enabled)
	}

	// Create notification channel (required on API 26+).
	ch, err := notification.NewChannel(
		vm, channelID, channelName,
		int32(notification.ImportanceHigh),
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

	// Build the notification.
	b, err := notification.NewBuilder(vm, appCtxObj, channelID)
	if err != nil {
		return fmt.Errorf("new builder: %w", err)
	}
	if _, err := b.SetSmallIcon1_1(iconDialogInfo); err != nil {
		return fmt.Errorf("set icon: %w", err)
	}
	if _, err := b.SetContentTitle("Push from Go"); err != nil {
		return fmt.Errorf("set title: %w", err)
	}
	if _, err := b.SetContentText("Delivered via JNI, no Java"); err != nil {
		return fmt.Errorf("set text: %w", err)
	}
	if _, err := b.SetAutoCancel(true); err != nil {
		return fmt.Errorf("set auto cancel: %w", err)
	}
	notif, err := b.Build()
	if err != nil {
		return fmt.Errorf("build notification: %w", err)
	}
	fmt.Fprintln(output, "notification built")

	if err := mgr.Notify2(1, notif); err != nil {
		return fmt.Errorf("notify: %w", err)
	}
	fmt.Fprintln(output, "notification posted!")

	return nil
}
