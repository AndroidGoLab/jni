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
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/AndroidGoLab/jni/provider/documents"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// Exported constants for SAF intent actions.
	fmt.Fprintln(&output, "Intent action constants:")
	fmt.Fprintf(&output, "  ActionOpenDocument:     %s\n", documents.ActionOpenDocument)
	fmt.Fprintf(&output, "  ActionOpenDocumentTree: %s\n", documents.ActionOpenDocumentTree)
	fmt.Fprintf(&output, "  ActionCreateDocument:   %s\n", documents.ActionCreateDocument)

	// Exported constants for intent extras.
	fmt.Fprintln(&output, "Intent extra constants:")
	fmt.Fprintf(&output, "  ExtraMimeTypes:     %s\n", documents.ExtraMimeTypes)
	fmt.Fprintf(&output, "  ExtraAllowMultiple: %s\n", documents.ExtraAllowMultiple)
	fmt.Fprintf(&output, "  ExtraTitle:         %s\n", documents.ExtraTitle)

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
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
