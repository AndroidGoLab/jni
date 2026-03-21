//go:build android

// Command health demonstrates the Health Connect API provided by the
// generated health/connect package. It is built as a c-shared library
// and packaged into an APK.
//
// The connect.Manager wraps the HealthConnectClient and provides raw
// methods for inserting, reading, aggregating, and deleting records.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	_ "github.com/AndroidGoLab/jni/health/connect"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// The health/connect package provides a Manager type wrapping
	// HealthConnectClient with raw methods:
	//   getOrCreateRaw(ctx) - obtain a HealthConnectClient
	//   insertRecordsRaw(records) - insert health records
	//   readRecordsRaw(request) - read health records
	//   aggregateRaw(request) - aggregate health data
	//   deleteRecordsRaw(recordType, timeRange) - delete records
	//
	// These are unexported and intended for use by higher-level wrappers.
	fmt.Fprintln(&output, "Health Connect Manager type available")
	fmt.Fprintln(&output, "Raw methods: getOrCreateRaw, insertRecordsRaw, readRecordsRaw, aggregateRaw, deleteRecordsRaw")
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
