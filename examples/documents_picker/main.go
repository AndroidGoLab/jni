//go:build android

// Command documents_picker demonstrates the DocumentsContract API surface.
// It exercises static URI builder methods, document classification checks,
// and ID extraction -- all using typed wrappers without raw JNI.
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/provider/documents"
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

	fmt.Fprintln(output, "=== Documents Picker ===")
	fmt.Fprintln(output)

	// --- Document column constants ---
	fmt.Fprintln(output, "Document columns:")
	fmt.Fprintf(output, "  DisplayName: %s\n", documents.ColumnDisplayName)
	fmt.Fprintf(output, "  DocumentId:  %s\n", documents.ColumnDocumentId)
	fmt.Fprintf(output, "  MimeType:    %s\n", documents.ColumnMimeType)
	fmt.Fprintf(output, "  Size:        %s\n", documents.ColumnSize)
	fmt.Fprintf(output, "  Flags:       %s\n", documents.ColumnFlags)
	fmt.Fprintf(output, "  LastModified: %s\n", documents.ColumnLastModified)
	fmt.Fprintln(output)

	// --- Root columns ---
	fmt.Fprintln(output, "Root columns:")
	fmt.Fprintf(output, "  RootId:        %s\n", documents.ColumnRootId)
	fmt.Fprintf(output, "  Title:         %s\n", documents.ColumnTitle)
	fmt.Fprintf(output, "  Summary:       %s\n", documents.ColumnSummary)
	fmt.Fprintf(output, "  AvailableBytes: %s\n", documents.ColumnAvailableBytes)
	fmt.Fprintf(output, "  CapacityBytes:  %s\n", documents.ColumnCapacityBytes)
	fmt.Fprintf(output, "  MimeTypes:     %s\n", documents.ColumnMimeTypes)
	fmt.Fprintln(output)

	dc := documents.Contract{VM: vm}

	const (
		authority = "com.android.externalstorage.documents"
		docID     = "primary:Documents/test.txt"
		rootID    = "primary"
	)

	// --- Build URIs ---
	fmt.Fprintln(output, "URI Builders:")

	docUri, err := dc.BuildDocumentUri(authority, docID)
	if err != nil {
		fmt.Fprintf(output, "  BuildDocumentUri: ERR %v\n", err)
	} else if docUri != nil && docUri.Ref() != 0 {
		fmt.Fprintln(output, "  BuildDocumentUri: OK")
	}

	rootUri, err := dc.BuildRootUri(authority, rootID)
	if err != nil {
		fmt.Fprintf(output, "  BuildRootUri: ERR %v\n", err)
	} else if rootUri != nil && rootUri.Ref() != 0 {
		fmt.Fprintln(output, "  BuildRootUri: OK")
	}

	rootsUri, err := dc.BuildRootsUri(authority)
	if err != nil {
		fmt.Fprintf(output, "  BuildRootsUri: ERR %v\n", err)
	} else if rootsUri != nil && rootsUri.Ref() != 0 {
		fmt.Fprintln(output, "  BuildRootsUri: OK")
	}

	treeUri, err := dc.BuildTreeDocumentUri(authority, docID)
	if err != nil {
		fmt.Fprintf(output, "  BuildTreeDocUri: ERR %v\n", err)
	} else if treeUri != nil && treeUri.Ref() != 0 {
		fmt.Fprintln(output, "  BuildTreeDocUri: OK")
	}

	childUri, err := dc.BuildChildDocumentsUri(authority, docID)
	if err != nil {
		fmt.Fprintf(output, "  BuildChildDocsUri: ERR %v\n", err)
	} else if childUri != nil && childUri.Ref() != 0 {
		fmt.Fprintln(output, "  BuildChildDocsUri: OK")
	}

	recentUri, err := dc.BuildRecentDocumentsUri(authority, rootID)
	if err != nil {
		fmt.Fprintf(output, "  BuildRecentDocsUri: ERR %v\n", err)
	} else if recentUri != nil && recentUri.Ref() != 0 {
		fmt.Fprintln(output, "  BuildRecentDocsUri: OK")
	}

	searchUri, err := dc.BuildSearchDocumentsUri(authority, rootID, "*.pdf")
	if err != nil {
		fmt.Fprintf(output, "  BuildSearchDocsUri: ERR %v\n", err)
	} else if searchUri != nil && searchUri.Ref() != 0 {
		fmt.Fprintln(output, "  BuildSearchDocsUri: OK")
	}
	fmt.Fprintln(output)

	// --- URI classification ---
	fmt.Fprintln(output, "URI Classification:")

	if docUri != nil && docUri.Ref() != 0 {
		isDoc, err := dc.IsDocumentUri(ctx.Obj, docUri)
		if err != nil {
			fmt.Fprintf(output, "  IsDocumentUri(docUri): ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  IsDocumentUri(docUri): %v\n", isDoc)
		}
	}

	if treeUri != nil && treeUri.Ref() != 0 {
		isTree, err := dc.IsTreeUri(treeUri)
		if err != nil {
			fmt.Fprintf(output, "  IsTreeUri(treeUri): ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  IsTreeUri(treeUri): %v\n", isTree)
		}
	}

	if rootsUri != nil && rootsUri.Ref() != 0 {
		isRoots, err := dc.IsRootsUri(ctx.Obj, rootsUri)
		if err != nil {
			fmt.Fprintf(output, "  IsRootsUri(rootsUri): ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  IsRootsUri(rootsUri): %v\n", isRoots)
		}
	}
	fmt.Fprintln(output)

	// --- Extract IDs ---
	fmt.Fprintln(output, "ID Extraction:")

	if docUri != nil && docUri.Ref() != 0 {
		id, err := dc.GetDocumentId(docUri)
		if err != nil {
			fmt.Fprintf(output, "  GetDocumentId: ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  DocID: %s\n", id)
		}
	}

	if rootUri != nil && rootUri.Ref() != 0 {
		id, err := dc.GetRootId(rootUri)
		if err != nil {
			fmt.Fprintf(output, "  GetRootId: ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  RootID: %s\n", id)
		}
	}

	if treeUri != nil && treeUri.Ref() != 0 {
		id, err := dc.GetTreeDocumentId(treeUri)
		if err != nil {
			fmt.Fprintf(output, "  GetTreeDocId: ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  TreeDocID: %s\n", id)
		}
	}

	if searchUri != nil && searchUri.Ref() != 0 {
		query, err := dc.GetSearchDocumentsQuery(searchUri)
		if err != nil {
			fmt.Fprintf(output, "  GetSearchQuery: ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  SearchQuery: %s\n", query)
		}
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Documents picker example complete.")
	return nil
}
