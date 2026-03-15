package specgen

import (
	"reflect"
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
