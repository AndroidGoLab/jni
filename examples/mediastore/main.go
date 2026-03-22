//go:build android

// Command mediastore demonstrates the MediaStore JNI bindings by
// calling real API methods to query media store version, generation,
// volume names, scanner URI, and more.
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
	"github.com/AndroidGoLab/jni/content/resolver"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/provider/media"
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

func run(vm *jni.VM, output *bytes.Buffer) error {
	fmt.Fprintln(output, "=== MediaStore ===")

	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	ms := media.MediaStore{VM: vm}

	// --- Constants ---
	fmt.Fprintln(output, "Constants:")
	fmt.Fprintf(output, "  Authority: %s\n", media.Authority)
	fmt.Fprintf(output, "  VolumeInternal: %s\n", media.VolumeInternal)
	fmt.Fprintf(output, "  VolumeExternal: %s\n", media.VolumeExternal)
	fmt.Fprintf(output, "  VolumeExternalPrimary: %s\n", media.VolumeExternalPrimary)

	// --- GetVersion ---
	fmt.Fprintln(output)
	version, err := ms.GetVersion1(ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "GetVersion: %v\n", err)
	} else {
		fmt.Fprintf(output, "Version: %s\n", version)
	}

	// GetVersion with volume name.
	versionExt, err := ms.GetVersion2_1(ctx.Obj, media.VolumeExternalPrimary)
	if err != nil {
		fmt.Fprintf(output, "GetVersion(external_primary): %v\n", err)
	} else {
		fmt.Fprintf(output, "Version(external_primary): %s\n", versionExt)
	}

	// --- GetGeneration ---
	gen, err := ms.GetGeneration(ctx.Obj, media.VolumeExternalPrimary)
	if err != nil {
		fmt.Fprintf(output, "GetGeneration: %v\n", err)
	} else {
		fmt.Fprintf(output, "Generation(external_primary): %d\n", gen)
	}

	genInt, err := ms.GetGeneration(ctx.Obj, media.VolumeInternal)
	if err != nil {
		fmt.Fprintf(output, "GetGeneration(internal): %v\n", err)
	} else {
		fmt.Fprintf(output, "Generation(internal): %d\n", genInt)
	}

	// --- GetMediaScannerUri ---
	fmt.Fprintln(output)
	scannerUriObj, err := ms.GetMediaScannerUri()
	if err != nil {
		fmt.Fprintf(output, "GetMediaScannerUri: %v\n", err)
	} else if scannerUriObj != nil {
		scannerUri := resolver.Uri{VM: vm, Obj: scannerUriObj}
		scannerStr, err := scannerUri.ToString()
		if err != nil {
			fmt.Fprintf(output, "ScannerUri.ToString: %v\n", err)
		} else {
			fmt.Fprintf(output, "MediaScannerUri: %s\n", scannerStr)
		}
	}

	// --- GetPickImagesMaxLimit ---
	maxLimit, err := ms.GetPickImagesMaxLimit()
	if err != nil {
		fmt.Fprintf(output, "GetPickImagesMaxLimit: %v\n", err)
	} else {
		fmt.Fprintf(output, "PickImagesMaxLimit: %d\n", maxLimit)
	}

	// --- CanManageMedia ---
	canManage, err := ms.CanManageMedia(ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "CanManageMedia: %v\n", err)
	} else {
		fmt.Fprintf(output, "CanManageMedia: %v\n", canManage)
	}

	// --- GetExternalVolumeNames ---
	fmt.Fprintln(output)
	volNamesObj, err := ms.GetExternalVolumeNames(ctx.Obj)
	if err != nil {
		fmt.Fprintf(output, "GetExternalVolumeNames: %v\n", err)
	} else if volNamesObj != nil {
		// The result is a java.util.Set<String>. Convert to array and print.
		fmt.Fprintln(output, "External volumes:")
		vm.Do(func(env *jni.Env) error {
			// Call Set.toArray() to get Object[].
			setCls, err := env.FindClass("java/util/Set")
			if err != nil {
				fmt.Fprintf(output, "  find Set class: %v\n", err)
				return nil
			}
			toArrayMid, err := env.GetMethodID(setCls, "toArray", "()[Ljava/lang/Object;")
			if err != nil {
				fmt.Fprintf(output, "  get toArray: %v\n", err)
				return nil
			}
			arrObj, err := env.CallObjectMethod(volNamesObj, toArrayMid)
			if err != nil {
				fmt.Fprintf(output, "  toArray: %v\n", err)
				return nil
			}
			if arrObj == nil {
				fmt.Fprintln(output, "  (empty)")
				return nil
			}
			arrLen := env.GetArrayLength((*jni.Array)(unsafe.Pointer(arrObj)))
			for i := int32(0); i < arrLen; i++ {
				elem, err := env.GetObjectArrayElement((*jni.ObjectArray)(unsafe.Pointer(arrObj)), i)
				if err != nil || elem == nil {
					continue
				}
				name := env.GoString((*jni.String)(unsafe.Pointer(elem)))
				fmt.Fprintf(output, "  [%d] %s\n", i, name)
			}
			return nil
		})
	}

	// --- Intent action constants ---
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Intent actions:")
	fmt.Fprintf(output, "  ImageCapture: %s\n", media.ActionImageCapture)
	fmt.Fprintf(output, "  VideoCapture: %s\n", media.ActionVideoCapture)
	fmt.Fprintf(output, "  PickImages: %s\n", media.ActionPickImages)

	// --- Query settings content URI via ContentResolver ---
	fmt.Fprintln(output)
	resolverObj, err := ctx.GetContentResolver()
	if err != nil {
		fmt.Fprintf(output, "GetContentResolver: %v\n", err)
	} else {
		uriHelper := resolver.Uri{VM: vm}
		// Query the MediaStore Images external content URI.
		imagesURI, err := uriHelper.Parse("content://media/external/images/media")
		if err != nil {
			fmt.Fprintf(output, "Uri.Parse: %v\n", err)
		} else {
			cr := resolver.ContentResolver{VM: vm, Obj: resolverObj}
			cursorObj, err := cr.Query4(imagesURI, nil, nil, nil)
			if err != nil {
				fmt.Fprintf(output, "Query images: %v\n", err)
			} else if cursorObj == nil {
				fmt.Fprintln(output, "Images query: null cursor")
			} else {
				cursor := resolver.Cursor{VM: vm, Obj: cursorObj}
				defer func() { _ = cursor.Close() }()
				count, err := cursor.GetCount()
				if err != nil {
					fmt.Fprintf(output, "count: %v\n", err)
				} else {
					fmt.Fprintf(output, "Images: %d rows\n", count)
				}
				colCount, err := cursor.GetColumnCount()
				if err == nil {
					fmt.Fprintf(output, "Images columns: %d\n", colCount)
					for c := int32(0); c < colCount && c < 10; c++ {
						name, err := cursor.GetColumnName(c)
						if err != nil {
							continue
						}
						fmt.Fprintf(output, "  col[%d]: %s\n", c, name)
					}
				}
			}
		}
	}

	return nil
}
