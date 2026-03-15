//go:build android

// Command health demonstrates the Health Connect API record type constants.
// It is built as a c-shared library and packaged into an APK.
//
// This example prints all available Health Connect record types defined
// in the health package. These constants identify different categories
// of health data (steps, heart rate, blood pressure, etc.) that can
// be read and written through the Health Connect API.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"bytes"
	"fmt"

	"github.com/AndroidGoLab/jni/health/connect"
)

func main() {}

var output bytes.Buffer

//export goRun
func goRun(cvm *C.JavaVM) {
	// Health Connect record type constants identify the kind of
	// health data to read or write through the Health Connect client.
	types := []struct {
		Name  string
		Value connect.RecordType
	}{
		{"Steps", connect.Steps},
		{"HeartRate", connect.HeartRate},
		{"Distance", connect.Distance},
		{"ActiveCaloriesBurned", connect.ActiveCaloriesBurned},
		{"TotalCaloriesBurned", connect.TotalCaloriesBurned},
		{"Weight", connect.Weight},
		{"Height", connect.Height},
		{"BloodPressure", connect.BloodPressure},
		{"BloodGlucose", connect.BloodGlucose},
		{"OxygenSaturation", connect.OxygenSaturation},
		{"SleepSession", connect.SleepSession},
	}

	fmt.Fprintln(&output, "Health Connect record types:")
	for _, t := range types {
		fmt.Fprintf(&output, "  %-25s = %q\n", t.Name, t.Value)
	}
}

//export goGetOutput
func goGetOutput() *C.char {
	return C.CString(output.String())
}
