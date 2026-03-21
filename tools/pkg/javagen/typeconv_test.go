package javagen

import (
	"testing"
)

func TestResolveType_Primitives(t *testing.T) {
	tests := []struct {
		javaType       string
		wantGoType     string
		wantJNISig     string
		wantCallSuffix string
		wantIsObject   bool
	}{
		{"boolean", "bool", "Z", "Boolean", false},
		{"byte", "int8", "B", "Byte", false},
		{"char", "uint16", "C", "Char", false},
		{"short", "int16", "S", "Short", false},
		{"int", "int32", "I", "Int", false},
		{"long", "int64", "J", "Long", false},
		{"float", "float32", "F", "Float", false},
		{"double", "float64", "D", "Double", false},
		{"void", "", "V", "Void", false},
	}
	for _, tt := range tests {
		t.Run(tt.javaType, func(t *testing.T) {
			tc := ResolveType(tt.javaType)
			if tc.GoType != tt.wantGoType {
				t.Errorf("GoType = %q, want %q", tc.GoType, tt.wantGoType)
			}
			if tc.JNISig != tt.wantJNISig {
				t.Errorf("JNISig = %q, want %q", tc.JNISig, tt.wantJNISig)
			}
			if tc.CallSuffix != tt.wantCallSuffix {
				t.Errorf("CallSuffix = %q, want %q", tc.CallSuffix, tt.wantCallSuffix)
			}
			if tc.IsObject != tt.wantIsObject {
				t.Errorf("IsObject = %v, want %v", tc.IsObject, tt.wantIsObject)
			}
		})
	}
}

func TestResolveType_String(t *testing.T) {
	tc := ResolveType("String")
	if tc.GoType != "string" {
		t.Errorf("GoType = %q, want string", tc.GoType)
	}
	if tc.JNISig != "Ljava/lang/String;" {
		t.Errorf("JNISig = %q", tc.JNISig)
	}
	if !tc.IsObject {
		t.Error("String should be IsObject")
	}

	tc2 := ResolveType("java.lang.String")
	if tc2.GoType != "string" {
		t.Errorf("java.lang.String GoType = %q", tc2.GoType)
	}
}

func TestResolveType_FullyQualified(t *testing.T) {
	tc := ResolveType("android.location.LocationManager")
	if tc.GoType != "*jni.Object" {
		t.Errorf("GoType = %q", tc.GoType)
	}
	if tc.JNISig != "Landroid/location/LocationManager;" {
		t.Errorf("JNISig = %q", tc.JNISig)
	}
	if !tc.IsObject {
		t.Error("expected IsObject")
	}
	if tc.CallSuffix != "Object" {
		t.Errorf("CallSuffix = %q", tc.CallSuffix)
	}
}

func TestResolveType_JavaLangShortNames(t *testing.T) {
	tests := []string{"Integer", "Long", "Boolean", "Float", "Double", "Byte", "Short", "Class"}
	for _, name := range tests {
		tc := ResolveType(name)
		if tc.GoType != "*jni.Object" {
			t.Errorf("ResolveType(%q).GoType = %q, want *jni.Object", name, tc.GoType)
		}
		if !tc.IsObject {
			t.Errorf("ResolveType(%q) should be IsObject", name)
		}
	}
}

func TestResolveType_UnknownShortName(t *testing.T) {
	tc := ResolveType("Xyz")
	if tc.GoType != "*jni.Object" {
		t.Errorf("GoType = %q", tc.GoType)
	}
	if tc.JNISig != "Ljava/lang/Xyz;" {
		t.Errorf("JNISig = %q", tc.JNISig)
	}
}

func TestResolveType_Generics(t *testing.T) {
	// Generic types must be erased: List<String> → Ljava/util/List;
	tc := ResolveType("java.util.List<java.lang.String>")
	if tc.JNISig != "Ljava/util/List;" {
		t.Errorf("JNISig = %q, want Ljava/util/List;", tc.JNISig)
	}
	if tc.GoType != "*jni.Object" {
		t.Errorf("GoType = %q, want *jni.Object", tc.GoType)
	}
	if !tc.IsObject {
		t.Error("expected IsObject")
	}

	// Nested generics.
	tc2 := ResolveType("java.util.Map<java.lang.String, java.util.List<java.lang.Integer>>")
	if tc2.JNISig != "Ljava/util/Map;" {
		t.Errorf("nested JNISig = %q, want Ljava/util/Map;", tc2.JNISig)
	}
}

func TestResolveType_Array(t *testing.T) {
	tc := ResolveType("int[]")
	if tc.GoType != "*jni.Object" {
		t.Errorf("GoType = %q, want *jni.Object", tc.GoType)
	}
	if tc.JNISig != "[I" {
		t.Errorf("JNISig = %q, want [I", tc.JNISig)
	}
	if !tc.IsObject {
		t.Error("arrays should be IsObject")
	}

	tc2 := ResolveType("String[]")
	if tc2.GoType != "*jni.Object" {
		t.Errorf("String[] GoType = %q, want *jni.Object", tc2.GoType)
	}

	tc3 := ResolveType("android.bluetooth.BluetoothDevice[]")
	if tc3.GoType != "*jni.Object" {
		t.Errorf("object array GoType = %q, want *jni.Object", tc3.GoType)
	}
}

func TestResolveType_Varargs(t *testing.T) {
	tc := ResolveType("int...")
	if tc.GoType != "*jni.Object" {
		t.Errorf("GoType = %q, want *jni.Object", tc.GoType)
	}
	if tc.JNISig != "[I" {
		t.Errorf("JNISig = %q, want [I", tc.JNISig)
	}
	if !tc.IsObject {
		t.Error("varargs should be IsObject (arrays)")
	}

	tc2 := ResolveType("long...")
	if tc2.JNISig != "[J" {
		t.Errorf("long... JNISig = %q, want [J", tc2.JNISig)
	}
}

func TestJNITypeSignature(t *testing.T) {
	tests := []struct {
		javaType string
		want     string
	}{
		{"int", "I"},
		{"boolean", "Z"},
		{"void", "V"},
		{"String", "Ljava/lang/String;"},
		{"android.location.Location", "Landroid/location/Location;"},
		{"int[]", "[I"},
		{"Integer", "Ljava/lang/Integer;"},
		{"Object", "Ljava/lang/Object;"},
		// Generic types must be erased for JNI signatures.
		{"java.util.Set<java.lang.String>", "Ljava/util/Set;"},
		{"java.util.List<android.telephony.CellInfo>", "Ljava/util/List;"},
		{"java.util.Map<java.lang.String, java.util.List<java.lang.Integer>>", "Ljava/util/Map;"},
		{"java.util.function.Consumer<android.location.Location>", "Ljava/util/function/Consumer;"},
		{"java.util.Set<java.util.Set<java.lang.String>>", "Ljava/util/Set;"},
		// Java varargs become arrays in JNI.
		{"int...", "[I"},
		{"long...", "[J"},
		// Incomplete generics from YAML splitting on commas inside generic types.
		// First half: "Map<String" (has '<' without '>').
		{"java.util.Map<java.lang.String", "Ljava/util/Map;"},
		// Second half: "SessionConfiguration>" (has '>' without '<').
		{"android.hardware.camera2.params.SessionConfiguration>", "Landroid/hardware/camera2/params/SessionConfiguration;"},
	}
	for _, tt := range tests {
		t.Run(tt.javaType, func(t *testing.T) {
			got := JNITypeSignature(tt.javaType)
			if got != tt.want {
				t.Errorf("JNITypeSignature(%q) = %q, want %q", tt.javaType, got, tt.want)
			}
		})
	}
}

func TestStripGenerics(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"java.util.Set<java.lang.String>", "java.util.Set"},
		{"java.util.Map<java.lang.String, java.lang.Integer>", "java.util.Map"},
		{"java.util.Set<java.util.Set<java.lang.String>>", "java.util.Set"},
		{"java.lang.String", "java.lang.String"},
		{"int", "int"},
		{"", ""},
		// Incomplete generics from YAML splitting on commas.
		{"java.util.Map<java.lang.String", "java.util.Map"},
		{"android.hardware.camera2.params.SessionConfiguration>", "android.hardware.camera2.params.SessionConfiguration"},
		{"android.telephony.TelephonyManager$NetworkSlicingException>", "android.telephony.TelephonyManager$NetworkSlicingException"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripGenerics(tt.input)
			if got != tt.want {
				t.Errorf("stripGenerics(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJNISignature(t *testing.T) {
	// No params, returns void.
	sig := JNISignature(nil, "void")
	if sig != "()V" {
		t.Errorf("JNISignature void = %q", sig)
	}

	// String param, returns int.
	params := []Param{
		{JavaType: "String"},
	}
	sig = JNISignature(params, "int")
	if sig != "(Ljava/lang/String;)I" {
		t.Errorf("JNISignature = %q", sig)
	}

	// Multiple params.
	params = []Param{
		{JavaType: "String"},
		{JavaType: "int"},
		{JavaType: "android.location.Location"},
	}
	sig = JNISignature(params, "boolean")
	if sig != "(Ljava/lang/String;ILandroid/location/Location;)Z" {
		t.Errorf("JNISignature = %q", sig)
	}
}

func TestJavaClassToSlash(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"android.location.LocationManager", "android/location/LocationManager"},
		{"java.lang.String", "java/lang/String"},
		{"Simple", "Simple"},
	}
	for _, tt := range tests {
		got := JavaClassToSlash(tt.input)
		if got != tt.want {
			t.Errorf("JavaClassToSlash(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParamConversionCode(t *testing.T) {
	// String param should produce conversion code.
	p := MergedParam{GoName: "name", GoType: "string", IsString: true}
	code := ParamConversionCode(p)
	if code == "" {
		t.Error("expected non-empty conversion code for string param")
	}

	// Non-string param should produce no code.
	p2 := MergedParam{GoName: "count", GoType: "int32"}
	code2 := ParamConversionCode(p2)
	if code2 != "" {
		t.Errorf("expected empty conversion code, got %q", code2)
	}
}

func TestJniValueFunc(t *testing.T) {
	tests := []struct {
		suffix string
		want   string
	}{
		{"Boolean", "BooleanValue"},
		{"Byte", "ByteValue"},
		{"Char", "CharValue"},
		{"Short", "ShortValue"},
		{"Int", "IntValue"},
		{"Long", "LongValue"},
		{"Float", "FloatValue"},
		{"Double", "DoubleValue"},
		{"Object", "ObjectValue"},
		{"Unknown", "IntValue"}, // default
	}
	for _, tt := range tests {
		got := jniValueFunc(tt.suffix)
		if got != tt.want {
			t.Errorf("jniValueFunc(%q) = %q, want %q", tt.suffix, got, tt.want)
		}
	}
}
