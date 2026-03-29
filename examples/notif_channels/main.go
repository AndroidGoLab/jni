//go:build android

// Command notif_channels creates multiple notification channels with
// different importance levels (high, default, low, min) and lists them.
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

type channelDef struct {
	id         string
	name       string
	importance int32
	label      string
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

	fmt.Fprintln(output, "=== Notif Channels ===")
	fmt.Fprintln(output)

	channels := []channelDef{
		{"ch_high", "High Priority", int32(notification.ImportanceHigh), "HIGH"},
		{"ch_default", "Default Priority", int32(notification.ImportanceDefault), "DEFAULT"},
		{"ch_low", "Low Priority", int32(notification.ImportanceLow), "LOW"},
		{"ch_min", "Minimal Priority", int32(notification.ImportanceMin), "MIN"},
	}

	for _, def := range channels {
		ch, err := notification.NewChannel(vm, def.id, def.name, def.importance)
		if err != nil {
			return fmt.Errorf("new channel %s: %w", def.id, err)
		}
		if err := mgr.CreateNotificationChannel(ch.Obj); err != nil {
			return fmt.Errorf("create channel %s: %w", def.id, err)
		}
		fmt.Fprintf(output, "created: %s (%s)\n", def.id, def.label)
	}

	fmt.Fprintln(output)

	// Verify channels exist by querying each one back.
	for _, def := range channels {
		chObj, err := mgr.GetNotificationChannel1(def.id)
		if err != nil {
			fmt.Fprintf(output, "query %s: error\n", def.id)
			continue
		}
		if chObj == nil || chObj.Ref() == 0 {
			fmt.Fprintf(output, "query %s: null\n", def.id)
		} else {
			fmt.Fprintf(output, "query %s: OK\n", def.id)
			vm.Do(func(env *jni.Env) error {
				env.DeleteGlobalRef(chObj)
				return nil
			})
		}
	}

	fmt.Fprintln(output)
	fmt.Fprintf(output, "total: %d channels\n", len(channels))

	return nil
}
