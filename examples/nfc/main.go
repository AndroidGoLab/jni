//go:build android

// Command nfc demonstrates using the NFC reader API.
// It is built as a c-shared library and packaged into an APK.
//
// This example shows the NFC adapter, NDEF tag, and ISO-DEP tag
// types along with all NFC reader mode flags and TNF type constants.
// Most methods in this package are unexported and intended to be
// called from within the nfc package itself or via higher-level
// wrappers; this example shows the exported types, constants, and
// the overall usage pattern.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/AndroidGoLab/jni/nfc"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// --- NFC reader mode flags ---
	// These constants control which tag technologies are polled
	// when enabling reader mode on the NfcAdapter.
	fmt.Fprintln(&output, "NFC reader mode flags:")
	fmt.Fprintf(&output, "  FlagReaderNfcA             = 0x%X\n", nfc.FlagReaderNfcA)
	fmt.Fprintf(&output, "  FlagReaderNfcB             = 0x%X\n", nfc.FlagReaderNfcB)
	fmt.Fprintf(&output, "  FlagReaderNfcF             = 0x%X\n", nfc.FlagReaderNfcF)
	fmt.Fprintf(&output, "  FlagReaderNfcV             = 0x%X\n", nfc.FlagReaderNfcV)
	fmt.Fprintf(&output, "  FlagReaderNfcBarcode       = 0x%X\n", nfc.FlagReaderNfcBarcode)
	fmt.Fprintf(&output, "  FlagReaderNoPlatformSounds = 0x%X\n", nfc.FlagReaderNoPlatformSounds)
	fmt.Fprintf(&output, "  FlagReaderSkipNdefCheck    = 0x%X\n", nfc.FlagReaderSkipNdefCheck)

	// Combine flags for multi-technology polling.
	flags := nfc.FlagReaderNfcA | nfc.FlagReaderNfcB | nfc.FlagReaderNoPlatformSounds
	fmt.Fprintf(&output, "\nCombined flags (NfcA + NfcB + NoPlatformSounds) = 0x%X\n", flags)

	// --- TNF (Type Name Format) constants for NDEF records ---
	fmt.Fprintln(&output, "\nTNF types:")
	fmt.Fprintf(&output, "  TnfEmpty        = %d\n", nfc.TnfEmpty)
	fmt.Fprintf(&output, "  TnfWellKnown    = %d\n", nfc.TnfWellKnown)
	fmt.Fprintf(&output, "  TnfMimeMedia    = %d\n", nfc.TnfMimeMedia)
	fmt.Fprintf(&output, "  TnfAbsoluteUri  = %d\n", nfc.TnfAbsoluteUri)
	fmt.Fprintf(&output, "  TnfExternalType = %d\n", nfc.TnfExternalType)
	fmt.Fprintf(&output, "  TnfUnknown      = %d\n", nfc.TnfUnknown)
	fmt.Fprintf(&output, "  TnfUnchanged    = %d\n", nfc.TnfUnchanged)

	// --- Exported types ---
	// The nfc package exposes three main wrapper types with exported
	// struct fields (VM and Obj). The Adapter has an exported Close():
	//
	//   nfc.Adapter   - wraps android.nfc.NfcAdapter
	//     .Close()    - releases the global JNI reference
	//
	// The following types have exported struct fields but their methods
	// are package-internal (unexported), designed to be called from
	// within the nfc package or via higher-level wrappers:
	//
	//   nfc.NdefTag   - wraps android.nfc.tech.Ndef
	//     .connect(), .getNdefMessageRaw(), .writeNdefMessageRaw()
	//     .makeReadOnly(), .isWritable(), .getMaxSize(), .close()
	//
	//   nfc.IsoDepTag - wraps android.nfc.tech.IsoDep
	//     .connect(), .transceive(), .setTimeoutMs()
	//     .getMaxTransceiveLength(), .close()
	//
	// Unexported data classes:
	//   tag        - fields: ID []byte, TechList []string
	//   ndefRecord - fields: TNF int, Type/ID/Payload []byte
	//
	// Callback:
	//   readerCallback{OnTag func(*jni.Object)} is registered via
	//   registerreaderCallback to implement NfcAdapter.ReaderCallback.

	// --- Adapter lifecycle ---
	var adapter nfc.Adapter
	// adapter.VM and adapter.Obj would be populated by the runtime.
	// Always close when done to release the JNI global reference.
	defer adapter.Close()

	fmt.Fprintln(&output, "\nNFC example complete.")
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
