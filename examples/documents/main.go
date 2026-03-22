//go:build android

// Command documents demonstrates the Android DocumentsContract API.
// It exercises the static URI builder methods and tests URI classification,
// using live JNI calls with no custom Java code.
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
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/provider/documents"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromUintptr(uintptr(activity.vm)),
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromUintptr(uintptr(activity.clazz)),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
}

// uriToString calls Uri.toString() via JNI and returns the Go string.
func uriToString(vm *jni.VM, uriObj *jni.Object) string {
	var result string
	vm.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(uriObj)
		mid, err := env.GetMethodID(cls, "toString", "()Ljava/lang/String;")
		if err != nil {
			return nil
		}
		strObj, err := env.CallObjectMethod(uriObj, mid)
		if err != nil {
			return nil
		}
		result = env.GoString((*jni.String)(unsafe.Pointer(strObj)))
		return nil
	})
	return result
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	fmt.Fprintln(output, "=== DocumentsContract ===")

	// --- Constants ---
	fmt.Fprintln(output, "Constants:")
	fmt.Fprintf(output, "  ActionDocSettings: %s\n", documents.ActionDocumentSettings)
	fmt.Fprintf(output, "  ExtraError:        %s\n", documents.ExtraError)
	fmt.Fprintf(output, "  ExtraInitialUri:   %s\n", documents.ExtraInitialUri)
	fmt.Fprintf(output, "  ProviderInterface: %s\n", documents.ProviderInterface)
	fmt.Fprintf(output, "  MetadataExif:      %s\n", documents.MetadataExif)
	fmt.Fprintf(output, "  MetadataTreeCount: %s\n", documents.MetadataTreeCount)
	fmt.Fprintf(output, "  MetadataTreeSize:  %s\n", documents.MetadataTreeSize)

	// The Contract type wraps static methods. Create a zero-value
	// receiver with the VM set for JNI access.
	dc := documents.Contract{VM: vm}

	// --- Build URIs ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "URI Builders:")

	const (
		testAuthority = "com.example.docs"
		testDocID     = "doc:123"
		testRootID    = "root:primary"
	)

	// BuildDocumentUri(authority, docId) -> Uri
	docUri, err := dc.BuildDocumentUri(testAuthority, testDocID)
	if err != nil {
		fmt.Fprintf(output, "  BuildDocumentUri: ERR %v\n", err)
	} else if docUri != nil && docUri.Ref() != 0 {
		s := uriToString(vm, docUri)
		fmt.Fprintf(output, "  DocUri: %s\n", s)
	}

	// BuildRootUri(authority, rootId) -> Uri
	rootUri, err := dc.BuildRootUri(testAuthority, testRootID)
	if err != nil {
		fmt.Fprintf(output, "  BuildRootUri: ERR %v\n", err)
	} else if rootUri != nil && rootUri.Ref() != 0 {
		s := uriToString(vm, rootUri)
		fmt.Fprintf(output, "  RootUri: %s\n", s)
	}

	// BuildRootsUri(authority) -> Uri
	rootsUri, err := dc.BuildRootsUri(testAuthority)
	if err != nil {
		fmt.Fprintf(output, "  BuildRootsUri: ERR %v\n", err)
	} else if rootsUri != nil && rootsUri.Ref() != 0 {
		s := uriToString(vm, rootsUri)
		fmt.Fprintf(output, "  RootsUri: %s\n", s)
	}

	// BuildTreeDocumentUri(authority, docId) -> Uri
	treeUri, err := dc.BuildTreeDocumentUri(testAuthority, testDocID)
	if err != nil {
		fmt.Fprintf(output, "  BuildTreeDocUri: ERR %v\n", err)
	} else if treeUri != nil && treeUri.Ref() != 0 {
		s := uriToString(vm, treeUri)
		fmt.Fprintf(output, "  TreeDocUri: %s\n", s)
	}

	// BuildChildDocumentsUri(authority, parentDocId) -> Uri
	childUri, err := dc.BuildChildDocumentsUri(testAuthority, testDocID)
	if err != nil {
		fmt.Fprintf(output, "  BuildChildDocsUri: ERR %v\n", err)
	} else if childUri != nil && childUri.Ref() != 0 {
		s := uriToString(vm, childUri)
		fmt.Fprintf(output, "  ChildDocsUri: %s\n", s)
	}

	// BuildRecentDocumentsUri(authority, rootId) -> Uri
	recentUri, err := dc.BuildRecentDocumentsUri(testAuthority, testRootID)
	if err != nil {
		fmt.Fprintf(output, "  BuildRecentDocsUri: ERR %v\n", err)
	} else if recentUri != nil && recentUri.Ref() != 0 {
		s := uriToString(vm, recentUri)
		fmt.Fprintf(output, "  RecentDocsUri: %s\n", s)
	}

	// BuildSearchDocumentsUri(authority, rootId, query) -> Uri
	searchUri, err := dc.BuildSearchDocumentsUri(testAuthority, testRootID, "test")
	if err != nil {
		fmt.Fprintf(output, "  BuildSearchDocsUri: ERR %v\n", err)
	} else if searchUri != nil && searchUri.Ref() != 0 {
		s := uriToString(vm, searchUri)
		fmt.Fprintf(output, "  SearchDocsUri: %s\n", s)
	}

	// --- Test URI classification ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "URI Classification:")

	// Test if docUri is a document URI
	if docUri != nil && docUri.Ref() != 0 {
		isDoc, err := dc.IsDocumentUri(ctx.Obj, docUri)
		if err != nil {
			fmt.Fprintf(output, "  IsDocumentUri(docUri): ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  IsDocumentUri(docUri): %v\n", isDoc)
		}
	}

	// Test if treeUri is a tree URI
	if treeUri != nil && treeUri.Ref() != 0 {
		isTree, err := dc.IsTreeUri(treeUri)
		if err != nil {
			fmt.Fprintf(output, "  IsTreeUri(treeUri): ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  IsTreeUri(treeUri): %v\n", isTree)
		}
	}

	// Test if rootUri is a root URI
	if rootUri != nil && rootUri.Ref() != 0 {
		isRoot, err := dc.IsRootUri(ctx.Obj, rootUri)
		if err != nil {
			fmt.Fprintf(output, "  IsRootUri(rootUri): ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  IsRootUri(rootUri): %v\n", isRoot)
		}
	}

	// Test if rootsUri is a roots URI
	if rootsUri != nil && rootsUri.Ref() != 0 {
		isRoots, err := dc.IsRootsUri(ctx.Obj, rootsUri)
		if err != nil {
			fmt.Fprintf(output, "  IsRootsUri(rootsUri): ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  IsRootsUri(rootsUri): %v\n", isRoots)
		}
	}

	// --- Extract IDs from URIs ---
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "URI ID Extraction:")

	if docUri != nil && docUri.Ref() != 0 {
		docID, err := dc.GetDocumentId(docUri)
		if err != nil {
			fmt.Fprintf(output, "  GetDocumentId: ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  DocID: %s\n", docID)
		}
	}

	if rootUri != nil && rootUri.Ref() != 0 {
		rID, err := dc.GetRootId(rootUri)
		if err != nil {
			fmt.Fprintf(output, "  GetRootId: ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  RootID: %s\n", rID)
		}
	}

	if treeUri != nil && treeUri.Ref() != 0 {
		treeDocID, err := dc.GetTreeDocumentId(treeUri)
		if err != nil {
			fmt.Fprintf(output, "  GetTreeDocId: ERR %v\n", err)
		} else {
			fmt.Fprintf(output, "  TreeDocID: %s\n", treeDocID)
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

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Documents example complete.")
	return nil
}
