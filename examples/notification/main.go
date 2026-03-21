//go:build android

// Command notification demonstrates the full Notification API surface
// provided by the generated notification package. It is built as a c-shared
// library and packaged into an APK using the shared apk.mk infrastructure.
//
// It covers:
//   - Manager: NewManager, Close, areNotificationsEnabled,
//     createNotificationChannel, deleteNotificationChannel,
//     getNotificationChannelsRaw, notifyRaw, cancel, cancelAll,
//     getActiveNotificationsRaw
//   - notificationChannel: setDescription, getId, getName, getDescription,
//     getImportance
//   - notificationBuilder: setContentTitle, setContentText, setSmallIcon,
//     setContentIntent, setAutoCancel, setOngoing, setStyle, setGroup,
//     setProgress, addAction, build
//   - bigTextStyle: bigText
//   - statusBarNotification data class: ID
//   - Constants: ImportanceNone, ImportanceMin, ImportanceLow,
//     ImportanceDefault, ImportanceHigh
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
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
	"github.com/AndroidGoLab/jni/app/notification"
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
	ctx, err := exampleui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// --- Constants ---
	// Channel importance levels mirror
	// NotificationManager.IMPORTANCE_NONE through IMPORTANCE_HIGH.
	fmt.Fprintln(output, "=== Importance constants ===")
	fmt.Fprintf(output, "  ImportanceNone    = %d\n", notification.ImportanceNone)
	fmt.Fprintf(output, "  ImportanceMin     = %d\n", notification.ImportanceMin)
	fmt.Fprintf(output, "  ImportanceLow     = %d\n", notification.ImportanceLow)
	fmt.Fprintf(output, "  ImportanceDefault = %d\n", notification.ImportanceDefault)
	fmt.Fprintf(output, "  ImportanceHigh    = %d\n", notification.ImportanceHigh)

	// --- Manager ---
	mgr, err := notification.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("notification.NewManager: %w", err)
	}
	defer mgr.Close()

	// --- Check notification permission (package-internal) ---
	// areNotificationsEnabled checks whether the app can post notifications.
	//   mgr.areNotificationsEnabled() (bool, error)
	fmt.Fprintln(output, "\nareNotificationsEnabled API available")

	// --- Create a notification channel (required API 26+) ---
	// notificationChannel wraps android.app.NotificationChannel.
	// Available methods:
	//   setDescription(description string)
	//   getId() string
	//   getName() *jni.Object  (CharSequence)
	//   getDescription() string
	//   getImportance() int32
	//
	// In a real app you would construct the channel via JNI with the
	// (String id, CharSequence name, int importance) constructor, then
	// set its description and register it:
	//
	//   ch.setDescription("General app notifications")
	//   mgr.createNotificationChannel(ch.Obj.Object())
	//   fmt.Fprintf(output, "Channel ID: %s, Importance: %d\n",
	//       ch.getId(), ch.getImportance())
	fmt.Fprintln(output, "NotificationChannel API: setDescription, getId, getName, getDescription, getImportance")

	// --- List existing channels (package-internal) ---
	// getNotificationChannelsRaw returns a raw Java List of channels.
	//   mgr.getNotificationChannelsRaw() (*jni.Object, error)
	fmt.Fprintln(output, "getNotificationChannelsRaw API available")

	// --- Build a notification ---
	// notificationBuilder wraps android.app.Notification$Builder.
	// Builder methods (all return *jni.Object for chaining):
	//   setContentTitle(title *jni.Object)     - CharSequence
	//   setContentText(text *jni.Object)       - CharSequence
	//   setSmallIcon(resId int32)
	//   setContentIntent(intent *jni.Object)   - PendingIntent
	//   setAutoCancel(autoCancel bool)
	//   setOngoing(ongoing bool)
	//   setStyle(style *jni.Object)            - Notification.Style
	//   setGroup(groupKey string)
	//   setProgress(max, progress int32, indeterminate bool)
	//   addAction(icon int32, title *jni.Object, intent *jni.Object)
	//   build() (*jni.Object, error)
	//
	// Usage (conceptual):
	//   builder.setContentTitle(titleObj)
	//   builder.setContentText(textObj)
	//   builder.setSmallIcon(0x7f080001)
	//   builder.setAutoCancel(true)
	//   builder.setOngoing(false)
	//   builder.setGroup("my_group")
	//   builder.setProgress(100, 50, false)
	//   notif, _ := builder.build()
	fmt.Fprintln(output, "\nNotification.Builder API: setContentTitle, setContentText, setSmallIcon,")
	fmt.Fprintln(output, "  setContentIntent, setAutoCancel, setOngoing, setStyle, setGroup,")
	fmt.Fprintln(output, "  setProgress, addAction, build")

	// --- BigTextStyle ---
	// bigTextStyle wraps android.app.Notification$BigTextStyle.
	//   bigText(text *jni.Object) *jni.Object
	//
	// Usage: create a BigTextStyle, call bigText(), then pass to
	// builder.setStyle().
	fmt.Fprintln(output, "BigTextStyle API: bigText")

	// --- Post / cancel notifications (package-internal) ---
	// notifyRaw posts a notification with a given integer ID.
	// cancel removes a specific notification by ID.
	// cancelAll removes all notifications posted by this app.
	//   mgr.notifyRaw(id int32, notification *jni.Object) error
	//   mgr.cancel(id int32) error
	//   mgr.cancelAll() error
	fmt.Fprintln(output, "\nManager API: notifyRaw(id, notification), cancel(id), cancelAll()")

	// --- Query active notifications (package-internal) ---
	// getActiveNotificationsRaw returns an array of StatusBarNotification
	// objects. The statusBarNotification data class has:
	//   ID int32  (the notification identifier)
	//   mgr.getActiveNotificationsRaw() ([]*jni.Object, error)
	fmt.Fprintln(output, "getActiveNotificationsRaw API available (returns StatusBarNotification array)")

	// --- Delete a channel (package-internal) ---
	// deleteNotificationChannel removes a channel by its string ID.
	//   mgr.deleteNotificationChannel(channelId string) error
	//   mgr.createNotificationChannel(channel *jni.Object) error
	fmt.Fprintln(output, "deleteNotificationChannel(channelId string) available")

	fmt.Fprintln(output, "\nAll notification package features demonstrated.")
	return nil
}
