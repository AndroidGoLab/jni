//go:build android

// Command notif_grouped posts multiple notifications in a group with
// a summary notification, demonstrating the bundled notification pattern.
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

	channelID   = "grouped_demo"
	channelName = "Grouped Demo"
	groupKey    = "demo_group"
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

type notifDef struct {
	id    int32
	title string
	text  string
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

	fmt.Fprintln(output, "=== Grouped Notifs ===")
	fmt.Fprintln(output)

	// Create channel.
	ch, err := notification.NewChannel(
		vm, channelID, channelName,
		int32(notification.ImportanceDefault),
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

	// Post individual child notifications.
	children := []notifDef{
		{10, "Message from Alice", "Hey, are you free?"},
		{11, "Message from Bob", "Meeting at 3pm"},
		{12, "Message from Carol", "Check this out!"},
	}

	for _, child := range children {
		b, err := notification.NewBuilder(vm, appCtxObj, channelID)
		if err != nil {
			return fmt.Errorf("new builder: %w", err)
		}
		if _, err := b.SetSmallIcon1_1(iconDialogInfo); err != nil {
			return fmt.Errorf("set icon: %w", err)
		}
		if _, err := b.SetContentTitle(child.title); err != nil {
			return fmt.Errorf("set title: %w", err)
		}
		if _, err := b.SetContentText(child.text); err != nil {
			return fmt.Errorf("set text: %w", err)
		}
		if _, err := b.SetGroup(groupKey); err != nil {
			return fmt.Errorf("set group: %w", err)
		}
		if _, err := b.SetAutoCancel(true); err != nil {
			return fmt.Errorf("set auto cancel: %w", err)
		}
		notif, err := b.Build()
		if err != nil {
			return fmt.Errorf("build: %w", err)
		}
		if err := mgr.Notify2(child.id, notif); err != nil {
			return fmt.Errorf("notify child: %w", err)
		}
		fmt.Fprintf(output, "posted: %s\n", child.title)
	}

	// Post the group summary notification.
	sb, err := notification.NewBuilder(vm, appCtxObj, channelID)
	if err != nil {
		return fmt.Errorf("new summary builder: %w", err)
	}
	if _, err := sb.SetSmallIcon1_1(iconDialogInfo); err != nil {
		return fmt.Errorf("set icon: %w", err)
	}
	if _, err := sb.SetContentTitle("3 new messages"); err != nil {
		return fmt.Errorf("set title: %w", err)
	}
	if _, err := sb.SetContentText("Alice, Bob, Carol"); err != nil {
		return fmt.Errorf("set text: %w", err)
	}
	if _, err := sb.SetGroup(groupKey); err != nil {
		return fmt.Errorf("set group: %w", err)
	}
	if _, err := sb.SetGroupSummary(true); err != nil {
		return fmt.Errorf("set group summary: %w", err)
	}
	if _, err := sb.SetAutoCancel(true); err != nil {
		return fmt.Errorf("set auto cancel: %w", err)
	}
	summary, err := sb.Build()
	if err != nil {
		return fmt.Errorf("build summary: %w", err)
	}
	if err := mgr.Notify2(0, summary); err != nil {
		return fmt.Errorf("notify summary: %w", err)
	}
	fmt.Fprintln(output, "summary posted!")
	fmt.Fprintf(output, "\ntotal: %d + summary\n", len(children))

	return nil
}
