//go:build android

// Command mediastore demonstrates the MediaStore JNI bindings.
// It is built as a c-shared library and packaged into an APK.
//
// This example prints all available MediaStore constants including
// intent actions, volume names, content URIs, and column names used
// for querying media files. It also describes the static methods
// available for creating media modification requests.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/AndroidGoLab/jni/provider/media"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// Intent action constants for capturing and picking media.
	fmt.Fprintln(&output, "Intent actions:")
	fmt.Fprintf(&output, "  ActionImageCapture = %q\n", media.ActionImageCapture)
	fmt.Fprintf(&output, "  ActionVideoCapture = %q\n", media.ActionVideoCapture)
	fmt.Fprintf(&output, "  ActionPickImages   = %q\n", media.ActionPickImages)
	fmt.Fprintf(&output, "  ExtraPickImagesMax = %q\n", media.ExtraPickImagesMax)

	// Volume constants identify storage volumes.
	fmt.Fprintln(&output, "Storage volumes:")
	fmt.Fprintf(&output, "  VolumeInternal        = %q\n", media.VolumeInternal)
	fmt.Fprintf(&output, "  VolumeExternal        = %q\n", media.VolumeExternal)
	fmt.Fprintf(&output, "  VolumeExternalPrimary = %q\n", media.VolumeExternalPrimary)

	// Content URI constants for different media collections.
	fmt.Fprintln(&output, "Content URIs:")
	fmt.Fprintf(&output, "  ImagesURI    = %q\n", media.ImagesURI)
	fmt.Fprintf(&output, "  VideoURI     = %q\n", media.VideoURI)
	fmt.Fprintf(&output, "  AudioURI     = %q\n", media.AudioURI)
	fmt.Fprintf(&output, "  DownloadsURI = %q\n", media.DownloadsURI)
	fmt.Fprintf(&output, "  FilesURI     = %q\n", media.FilesURI)

	// Column name constants used in ContentResolver queries.
	fmt.Fprintln(&output, "Column names:")
	fmt.Fprintf(&output, "  ColumnID           = %q\n", media.ColumnID)
	fmt.Fprintf(&output, "  ColumnDisplayName  = %q\n", media.ColumnDisplayName)
	fmt.Fprintf(&output, "  ColumnMimeType     = %q\n", media.ColumnMimeType)
	fmt.Fprintf(&output, "  ColumnSize         = %q\n", media.ColumnSize)
	fmt.Fprintf(&output, "  ColumnDateAdded    = %q\n", media.ColumnDateAdded)
	fmt.Fprintf(&output, "  ColumnDateModified = %q\n", media.ColumnDateModified)
	fmt.Fprintf(&output, "  ColumnRelativePath = %q\n", media.ColumnRelativePath)
	fmt.Fprintf(&output, "  ColumnDuration     = %q\n", media.ColumnDuration)
	fmt.Fprintf(&output, "  ColumnWidth        = %q\n", media.ColumnWidth)
	fmt.Fprintf(&output, "  ColumnHeight       = %q\n", media.ColumnHeight)

	// The mediaStore wrapper provides static methods:
	//   getExternalVolumeNamesRaw(ctx)              - get external volume names
	//   createWriteRequestRaw(resolver, uris)       - create write permission request
	//   createTrashRequestRaw(resolver, uris, trash)- create trash/untrash request
	//   createDeleteRequestRaw(resolver, uris)      - create delete permission request
	//   createFavoriteRequestRaw(resolver, uris, fav)- create favorite toggle request
	fmt.Fprintln(&output, "mediastore bindings available for media operations")
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
