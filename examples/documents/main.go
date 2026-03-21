//go:build android

// Command documents demonstrates the Android DocumentsContract API constants.
// It is built as a c-shared library and packaged into an APK.
//
// This package provides Go bindings for android.provider.DocumentsContract.
// All methods are on the unexported documentsContract type, but the package
// exports several intent action and extra constants used with the Storage
// Access Framework (SAF).
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
	"unsafe"
	"bytes"
	"fmt"

	"github.com/AndroidGoLab/jni/provider/documents"
	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/capi"
	"github.com/AndroidGoLab/jni/exampleui"
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
	// Exported constants from DocumentsContract.
	fmt.Fprintln(output, "DocumentsContract constants:")
	fmt.Fprintf(output, "  ActionDocumentSettings: %s\n", documents.ActionDocumentSettings)
	fmt.Fprintf(output, "  ExtraError:             %s\n", documents.ExtraError)
	fmt.Fprintf(output, "  ExtraInitialUri:        %s\n", documents.ExtraInitialUri)
	fmt.Fprintf(output, "  ProviderInterface:      %s\n", documents.ProviderInterface)

	// The following methods are all on the unexported documentsContract type
	// and are used by higher-level Go wrappers:
	//
	//   documentsContract.createDocumentRaw(resolver, parentUri *jni.Object, mimeType, displayName string) (*jni.Object, error)
	//     Creates a new document under the given parent URI.
	//
	//   documentsContract.renameDocumentRaw(resolver, uri *jni.Object, displayName string) (*jni.Object, error)
	//     Renames the document at the given URI.
	//
	//   documentsContract.copyDocumentRaw(resolver, srcUri, destParentUri *jni.Object) (*jni.Object, error)
	//     Copies a document to a destination parent URI.
	//
	//   documentsContract.moveDocumentRaw(resolver, srcUri, srcParentUri, destParentUri *jni.Object) (*jni.Object, error)
	//     Moves a document from one parent to another.
	//
	//   documentsContract.deleteDocumentRaw(resolver, uri *jni.Object) (bool, error)
	//     Deletes the document at the given URI.
	//
	//   documentsContract.isDocumentUriRaw(ctx, uri *jni.Object) bool
	//     Tests if the given URI represents a document.
	return nil
}
