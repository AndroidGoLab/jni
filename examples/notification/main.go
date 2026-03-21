//go:build android

// Command notification posts a notification from Go via JNI,
// with no custom Java code.
//
// On Android 13+ the POST_NOTIFICATIONS permission must be granted.
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
	"github.com/AndroidGoLab/jni/app/notification"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/examples/common/ui"
)

const (
	// android.R.drawable.ic_dialog_info
	iconDialogInfo = 17301620

	notificationChannelID   = "demo"
	notificationChannelName = "Demo"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	println("ANativeActivity_onCreate called")
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),
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

	// Create a NotificationManager from the app context.
	mgr, err := notification.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("notification.NewManager: %w", err)
	}
	defer mgr.Close()
	fmt.Fprintln(output, "NotificationManager created")

	// Create and register a notification channel (required on API 26+).
	ch, err := notification.NewChannel(
		vm, notificationChannelID, notificationChannelName,
		int32(notification.ImportanceHigh),
	)
	if err != nil {
		return fmt.Errorf("new channel: %w", err)
	}
	if err := mgr.CreateNotificationChannel(ch.Obj); err != nil {
		return fmt.Errorf("create channel: %w", err)
	}
	fmt.Fprintln(output, "Channel created (importance=high)")

	// Get the application context for the builder.
	appCtxObj, err := ctx.GetApplicationContext()
	if err != nil {
		return fmt.Errorf("get app context: %w", err)
	}

	// Build the notification.
	b, err := notification.NewBuilder(vm, appCtxObj, notificationChannelID)
	if err != nil {
		return fmt.Errorf("new builder: %w", err)
	}
	if _, err := b.SetSmallIcon1_1(iconDialogInfo); err != nil {
		return fmt.Errorf("set icon: %w", err)
	}
	if _, err := b.SetContentTitle("Hello from Go!"); err != nil {
		return fmt.Errorf("set title: %w", err)
	}
	if _, err := b.SetContentText("Posted via JNI — no Java code"); err != nil {
		return fmt.Errorf("set text: %w", err)
	}
	notif, err := b.Build()
	if err != nil {
		return fmt.Errorf("build notification: %w", err)
	}
	fmt.Fprintln(output, "Notification built")

	// Post the notification.
	if err := mgr.Notify2(1, notif); err != nil {
		return fmt.Errorf("notify: %w", err)
	}
	fmt.Fprintln(output, "Notification posted!")

	return nil
}
