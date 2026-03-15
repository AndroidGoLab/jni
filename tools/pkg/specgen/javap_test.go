package specgen

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitRespectingGenerics(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "single_primitive",
			input:    "int",
			expected: []string{"int"},
		},
		{
			name:     "two_primitives",
			input:    "int, boolean",
			expected: []string{"int", " boolean"},
		},
		{
			name:     "generic_with_comma",
			input:    "java.util.Map<java.lang.String, android.hardware.camera2.params.SessionConfiguration>",
			expected: []string{"java.util.Map<java.lang.String, android.hardware.camera2.params.SessionConfiguration>"},
		},
		{
			name:     "generic_followed_by_plain",
			input:    "java.util.Map<java.lang.String, java.lang.String>, int",
			expected: []string{"java.util.Map<java.lang.String, java.lang.String>", " int"},
		},
		{
			name:     "nested_generics",
			input:    "java.util.Map<java.lang.String, java.util.List<java.util.Set<java.lang.Integer>>>",
			expected: []string{"java.util.Map<java.lang.String, java.util.List<java.util.Set<java.lang.Integer>>>"},
		},
		{
			name:     "multiple_generics_with_commas",
			input:    "java.util.Map<A, B>, java.util.List<C, D>",
			expected: []string{"java.util.Map<A, B>", " java.util.List<C, D>"},
		},
		{
			name:     "mixed_params_with_generic_in_middle",
			input:    "int, java.util.Map<K, V>, boolean",
			expected: []string{"int", " java.util.Map<K, V>", " boolean"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := splitRespectingGenerics(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestParseParams(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    string
		expected []JavapParam
	}{
		{
			name:     "empty",
			input:    "",
			expected: nil,
		},
		{
			name:  "single_primitive",
			input: "int",
			expected: []JavapParam{
				{JavaType: "int"},
			},
		},
		{
			name:  "generic_map_not_split",
			input: "java.util.Map<java.lang.String, android.hardware.camera2.params.SessionConfiguration>",
			expected: []JavapParam{
				{JavaType: "java.util.Map<java.lang.String, android.hardware.camera2.params.SessionConfiguration>"},
			},
		},
		{
			name:  "generic_followed_by_plain_params",
			input: "java.util.concurrent.Executor, android.os.OutcomeReceiver<android.telecom.CallControl, android.telecom.CallException>, android.telecom.CallControlCallback",
			expected: []JavapParam{
				{JavaType: "java.util.concurrent.Executor"},
				{JavaType: "android.os.OutcomeReceiver<android.telecom.CallControl, android.telecom.CallException>"},
				{JavaType: "android.telecom.CallControlCallback"},
			},
		},
		{
			name:  "nested_generics",
			input: "java.util.Map<java.lang.String, java.util.List<java.util.Set<java.lang.Integer>>>",
			expected: []JavapParam{
				{JavaType: "java.util.Map<java.lang.String, java.util.List<java.util.Set<java.lang.Integer>>>"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := parseParams(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestParseJavap_ConstantValues(t *testing.T) {
	// Simulates javap -public -verbose output with ConstantValue attributes.
	verboseOutput := strings.Join([]string{
		`Classfile jar:file:///android.jar!/android/location/LocationManager.class`,
		`  Compiled from "LocationManager.java"`,
		`public class android.location.LocationManager`,
		`  minor version: 0`,
		`  major version: 52`,
		`  flags: (0x0021) ACC_PUBLIC, ACC_SUPER`,
		`Constant pool:`,
		`    #1 = Utf8               GPS_PROVIDER`,
		`{`,
		`  public static final java.lang.String GPS_PROVIDER;`,
		`    descriptor: Ljava/lang/String;`,
		`    flags: (0x0019) ACC_PUBLIC, ACC_STATIC, ACC_FINAL`,
		`    ConstantValue: String gps`,
		``,
		`  public static final java.lang.String NETWORK_PROVIDER;`,
		`    descriptor: Ljava/lang/String;`,
		`    flags: (0x0019) ACC_PUBLIC, ACC_STATIC, ACC_FINAL`,
		`    ConstantValue: String network`,
		``,
		`  public static final int SOME_INT;`,
		`    descriptor: I`,
		`    flags: (0x0019) ACC_PUBLIC, ACC_STATIC, ACC_FINAL`,
		`    ConstantValue: int 42`,
		``,
		`  public boolean isLocationEnabled();`,
		`    descriptor: ()Z`,
		`    flags: (0x0001) ACC_PUBLIC`,
		`}`,
	}, "\n")

	jc, err := parseJavap(verboseOutput)
	if err != nil {
		t.Fatalf("parseJavap: %v", err)
	}

	if jc.FullName != "android.location.LocationManager" {
		t.Errorf("FullName = %q, want %q", jc.FullName, "android.location.LocationManager")
	}

	// Verify constants were parsed with values.
	wantConstants := []JavapConstant{
		{Name: "GPS_PROVIDER", JavaType: "java.lang.String", Value: "gps"},
		{Name: "NETWORK_PROVIDER", JavaType: "java.lang.String", Value: "network"},
		{Name: "SOME_INT", JavaType: "int", Value: "42"},
	}
	if !reflect.DeepEqual(jc.Constants, wantConstants) {
		t.Errorf("Constants:\n  got  %+v\n  want %+v", jc.Constants, wantConstants)
	}

	// Verify methods still parse.
	if len(jc.Methods) != 1 {
		t.Fatalf("len(Methods) = %d, want 1", len(jc.Methods))
	}
	if jc.Methods[0].Name != "isLocationEnabled" {
		t.Errorf("Methods[0].Name = %q, want %q", jc.Methods[0].Name, "isLocationEnabled")
	}
}

func TestParseJavap_ImplementsWithoutBrace(t *testing.T) {
	// In javap -verbose output, the class line does not end with {.
	verboseOutput := strings.Join([]string{
		`Classfile jar:file:///android.jar!/android/location/GnssStatus.class`,
		`  Compiled from "GnssStatus.java"`,
		`public final class android.location.GnssStatus implements android.os.Parcelable`,
		`  minor version: 0`,
		`{`,
		`  public static final int CONSTELLATION_GPS;`,
		`    descriptor: I`,
		`    flags: (0x0019) ACC_PUBLIC, ACC_STATIC, ACC_FINAL`,
		`    ConstantValue: int 1`,
		`}`,
	}, "\n")

	jc, err := parseJavap(verboseOutput)
	if err != nil {
		t.Fatalf("parseJavap: %v", err)
	}

	if jc.FullName != "android.location.GnssStatus" {
		t.Errorf("FullName = %q, want %q", jc.FullName, "android.location.GnssStatus")
	}
	if !jc.IsFinal {
		t.Error("IsFinal = false, want true")
	}
	if len(jc.Implements) != 1 || jc.Implements[0] != "android.os.Parcelable" {
		t.Errorf("Implements = %v, want [android.os.Parcelable]", jc.Implements)
	}
	if len(jc.Constants) != 1 {
		t.Fatalf("len(Constants) = %d, want 1", len(jc.Constants))
	}
	if jc.Constants[0].Value != "1" {
		t.Errorf("Constants[0].Value = %q, want %q", jc.Constants[0].Value, "1")
	}
}

func TestParseJavap_NonVerboseStillWorks(t *testing.T) {
	// Non-verbose javap output (no ConstantValue lines).
	output := strings.Join([]string{
		`Compiled from "LocationManager.java"`,
		`public class android.location.LocationManager {`,
		`  public static final java.lang.String GPS_PROVIDER;`,
		`  public static final int SOME_INT;`,
		`  public boolean isLocationEnabled();`,
		`}`,
	}, "\n")

	jc, err := parseJavap(output)
	if err != nil {
		t.Fatalf("parseJavap: %v", err)
	}

	// Constants should be parsed but with empty values.
	if len(jc.Constants) != 2 {
		t.Fatalf("len(Constants) = %d, want 2", len(jc.Constants))
	}
	if jc.Constants[0].Value != "" {
		t.Errorf("Constants[0].Value = %q, want empty", jc.Constants[0].Value)
	}
	if jc.Constants[1].Value != "" {
		t.Errorf("Constants[1].Value = %q, want empty", jc.Constants[1].Value)
	}
}
