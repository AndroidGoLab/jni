//go:build android

// Command media_ringtone queries the MediaStore for version information and
// uses the ContentResolver and Cursor typed wrappers to list audio entries
// from internal and external storage.
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
	"github.com/AndroidGoLab/jni/content/resolver"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	media "github.com/AndroidGoLab/jni/provider/media"
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

// queryAudioEntries queries audio entries from the given URI using typed
// ContentResolver and Cursor wrappers, printing up to maxEntries.
func queryAudioEntries(
	vm *jni.VM,
	cr *resolver.ContentResolver,
	uri *resolver.Uri,
	output *bytes.Buffer,
	maxEntries int,
) {
	// Use Query4 (Bundle-based, API 26+): query(Uri, projection, bundle, cancellationSignal).
	// Pass nil for projection (all columns), bundle, and cancellation signal.
	cursorObj, err := cr.Query4(uri.Obj, nil, nil, nil)
	if err != nil {
		fmt.Fprintf(output, "  (query failed: %v)\n", err)
		return
	}
	if cursorObj == nil || cursorObj.Ref() == 0 {
		fmt.Fprintln(output, "  (no results)")
		return
	}

	cursor := &resolver.Cursor{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(cursorObj))}

	count, err := cursor.GetCount()
	if err != nil {
		fmt.Fprintf(output, "  (getCount failed: %v)\n", err)
		_ = cursor.Close()
		vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(cursorObj); return nil })
		return
	}
	fmt.Fprintf(output, "  Total entries: %d\n", count)

	// Get column indices by name.
	idCol, _ := cursor.GetColumnIndex("_id")
	titleCol, _ := cursor.GetColumnIndex("title")
	mimeCol, _ := cursor.GetColumnIndex("mime_type")

	shown := 0
	for shown < maxEntries && shown < 3 {
		moved, err := cursor.MoveToNext()
		if err != nil || !moved {
			break
		}

		id, _ := cursor.GetString(idCol)
		if id == "" {
			id = "(null)"
		}
		title, _ := cursor.GetString(titleCol)
		if title == "" {
			title = "(null)"
		}
		mime, _ := cursor.GetString(mimeCol)
		if mime == "" {
			mime = "(null)"
		}

		fmt.Fprintf(output, "  [%s] %s (%s)\n", id, title, mime)
		shown++
	}

	if int(count) > shown {
		fmt.Fprintf(output, "  ... and %d more\n", int(count)-shown)
	}

	_ = cursor.Close()
	vm.Do(func(env *jni.Env) error { env.DeleteGlobalRef(cursorObj); return nil })
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== Ringtone Query Demo ===")
	ui.RenderOutput()

	// Get MediaStore version info.
	store, err := media.NewMediaStore(vm)
	if err != nil {
		return fmt.Errorf("NewMediaStore: %w", err)
	}
	if store == nil || store.Obj == nil || store.Obj.Ref() == 0 {
		return fmt.Errorf("NewMediaStore: returned null")
	}
	defer func() {
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(store.Obj)
			return nil
		})
	}()

	ver, err := store.GetVersion1(ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "MediaStore version: err=%v\n", err)
	} else {
		fmt.Fprintf(output, "MediaStore version: %s\n", ver)
	}
	ui.RenderOutput()

	// Get content resolver via typed Context method.
	resolverObj, err := ctx.GetContentResolver()
	if err != nil {
		return fmt.Errorf("getContentResolver: %w", err)
	}
	if resolverObj == nil || resolverObj.Ref() == 0 {
		return fmt.Errorf("content resolver is null")
	}
	cr := &resolver.ContentResolver{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(resolverObj))}
	fmt.Fprintln(output, "ContentResolver: obtained")
	ui.RenderOutput()

	// Use Uri.Parse (static method) to create URI objects.
	// We need a Uri instance to call Parse on (it's a static method on the wrapper).
	uriHelper := &resolver.Uri{VM: vm}

	// Query internal audio media.
	internalUriObj, err := uriHelper.Parse("content://media/internal/audio/media")
	if err != nil {
		fmt.Fprintf(output, "Parse internal URI: %v\n", err)
	} else if internalUriObj != nil && internalUriObj.Ref() != 0 {
		internalUri := &resolver.Uri{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(internalUriObj))}
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "Internal audio media:")
		queryAudioEntries(vm, cr, internalUri, output, 10)
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(internalUriObj)
			return nil
		})
		ui.RenderOutput()
	}

	// Query external audio media.
	externalUriObj, err := uriHelper.Parse("content://media/external/audio/media")
	if err != nil {
		fmt.Fprintf(output, "Parse external URI: %v\n", err)
	} else if externalUriObj != nil && externalUriObj.Ref() != 0 {
		externalUri := &resolver.Uri{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(externalUriObj))}
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "External audio media:")
		queryAudioEntries(vm, cr, externalUri, output, 10)
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(externalUriObj)
			return nil
		})
		ui.RenderOutput()
	}

	// Show MediaStore constants.
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "MediaStore constants:")
	fmt.Fprintf(output, "  Authority: %s\n", media.Authority)
	fmt.Fprintf(output, "  VOLUME_EXTERNAL: %s\n", media.VolumeExternal)
	fmt.Fprintf(output, "  VOLUME_INTERNAL: %s\n", media.VolumeInternal)
	ui.RenderOutput()

	// Clean up resolver ref.
	vm.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(resolverObj)
		return nil
	})

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Ringtone example complete.")
	return nil
}
