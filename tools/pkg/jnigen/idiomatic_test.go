package jnigen

import "testing"

func TestErrorDescription(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"JNI_OK", "success"},
		{"JNI_ERR", "general error"},
		{"JNI_EDETACHED", "thread detached"},
		{"JNI_EVERSION", "version error"},
		{"JNI_ENOMEM", "out of memory"},
		{"JNI_EEXIST", "VM already exists"},
		{"JNI_EINVAL", "invalid argument"},
		{"JNI_UNKNOWN", "unknown error"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := errorDescription(tt.input); got != tt.want {
				t.Errorf("errorDescription(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCapiCast(t *testing.T) {
	// Known types produce a cast.
	if got := capiCast("int32", "x"); got != "capi.Jint(x)" {
		t.Errorf("capiCast(int32, x) = %q, want capi.Jint(x)", got)
	}
	if got := capiCast("float64", "y"); got != "capi.Jdouble(y)" {
		t.Errorf("capiCast(float64, y) = %q, want capi.Jdouble(y)", got)
	}
	// Unknown type returns the name unchanged.
	if got := capiCast("*Object", "obj"); got != "obj" {
		t.Errorf("capiCast(*Object, obj) = %q, want obj", got)
	}
}

func TestGoTypeToCapi(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"int32", "capi.Jint"},
		{"int64", "capi.Jlong"},
		{"float32", "capi.Jfloat"},
		{"float64", "capi.Jdouble"},
		{"uint8", "capi.Jboolean"},
		{"int8", "capi.Jbyte"},
		{"uint16", "capi.Jchar"},
		{"int16", "capi.Jshort"},
		{"string", ""},
		{"*Object", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := goTypeToCapi(tt.input); got != tt.want {
				t.Errorf("goTypeToCapi(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildReturnConversion(t *testing.T) {
	tests := []struct {
		goType    string
		transform string
		want      string
	}{
		{"*Object", "", "&Object{ref: _ret}"},
		{"*Class", "", "&Class{Object{ref: capi.Object(_ret)}}"},
		{"*String", "", "&String{Object{ref: capi.Object(_ret)}}"},
		{"*Throwable", "", "&Throwable{Object{ref: capi.Object(_ret)}}"},
		{"*Array", "", "&Array{Object{ref: capi.Object(_ret)}}"},
		{"*ObjectArray", "", "&ObjectArray{Array{Object{ref: capi.Object(_ret)}}}"},
		{"*WeakRef", "", "&WeakRef{Object{ref: capi.Object(_ret)}}"},
		{"bool", "", "_ret != 0"},
		{"int32", "", "int32(_ret)"},
		{"int64", "", "int64(_ret)"},
		{"float32", "", "float32(_ret)"},
		{"float64", "", "float64(_ret)"},
		{"uint8", "", "uint8(_ret)"},
		{"int8", "", "int8(_ret)"},
		{"uint16", "", "uint16(_ret)"},
		{"int16", "", "int16(_ret)"},
		{"MethodID", "", "_ret"},
		{"FieldID", "", "_ret"},
		{"unsafe.Pointer", "", "unsafe.Pointer(_ret)"},
		{"error", "jni_error", ""},
		{"ObjectRefType", "", "ObjectRefType(_ret)"},
		{"*IntArray", "", "&IntArray{Array{Object{ref: capi.Object(_ret)}}}"},
		{"SomeUnknown", "", "_ret"},
	}
	for _, tt := range tests {
		t.Run(tt.goType, func(t *testing.T) {
			if got := buildReturnConversion(tt.goType, tt.transform); got != tt.want {
				t.Errorf("buildReturnConversion(%q, %q) = %q, want %q", tt.goType, tt.transform, got, tt.want)
			}
		})
	}
}

func TestZeroValue(t *testing.T) {
	tests := []struct {
		goType string
		want   string
	}{
		{"bool", "false"},
		{"int32", "0"},
		{"int64", "0"},
		{"float32", "0"},
		{"float64", "0"},
		{"uint8", "0"},
		{"int8", "0"},
		{"uint16", "0"},
		{"int16", "0"},
		{"MethodID", "nil"},
		{"FieldID", "nil"},
		{"unsafe.Pointer", "nil"},
		{"ObjectRefType", "0"},
		{"*Object", "nil"},
		{"*Class", "nil"},
		{"SomeType", "0"},
	}
	for _, tt := range tests {
		t.Run(tt.goType, func(t *testing.T) {
			if got := zeroValue(tt.goType); got != tt.want {
				t.Errorf("zeroValue(%q) = %q, want %q", tt.goType, got, tt.want)
			}
		})
	}
}

func TestBuildCapiArgs_ImplicitWithCast(t *testing.T) {
	args := []CapiArgInfo{
		{IsImplicit: true, GoType: "int32", Implicit: "env.ptr"},
		{IsImplicit: true, GoType: "*Env", Implicit: "e.ptr"},
		{Name: "name", GoType: "string"},
	}
	got := buildCapiArgs(args)
	if got == "" {
		t.Error("expected non-empty result")
	}
	// First arg should be cast: capi.Jint(env.ptr)
	if !contains(got, "capi.Jint(env.ptr)") {
		t.Errorf("expected capi.Jint cast, got %q", got)
	}
	// Second arg has unknown type for goTypeToCapi, so passes through.
	if !contains(got, "e.ptr") {
		t.Errorf("expected e.ptr passthrough, got %q", got)
	}
}

func TestBuildCapiArgs_Empty(t *testing.T) {
	got := buildCapiArgs(nil)
	if got != "" {
		t.Errorf("expected empty string for nil params, got %q", got)
	}
}

func TestCapiArgExpr_AllTypes(t *testing.T) {
	tests := []struct {
		name   string
		arg    CapiArgInfo
		expect string
	}{
		{
			"string",
			CapiArgInfo{Name: "name", GoType: "string", CType: "const char*"},
			"(*capi.Cchar)(unsafe.Pointer(&cName[0]))",
		},
		{
			"bool",
			CapiArgInfo{Name: "flag", GoType: "bool"},
			"cFlag",
		},
		{
			"variadic value",
			CapiArgInfo{Name: "args", GoType: "...Value", IsVariadic: true},
			"cArgs",
		},
		{
			"*Object",
			CapiArgInfo{Name: "obj", GoType: "*Object"},
			"obj.Ref()",
		},
		{
			"*Class",
			CapiArgInfo{Name: "cls", GoType: "*Class"},
			"capi.Class(cls.Ref())",
		},
		{
			"*String",
			CapiArgInfo{Name: "s", GoType: "*String"},
			"capi.String(s.Ref())",
		},
		{
			"*Throwable",
			CapiArgInfo{Name: "t", GoType: "*Throwable"},
			"capi.Throwable(t.Ref())",
		},
		{
			"*Array",
			CapiArgInfo{Name: "a", GoType: "*Array"},
			"capi.Array(a.Ref())",
		},
		{
			"*ObjectArray",
			CapiArgInfo{Name: "a", GoType: "*ObjectArray"},
			"capi.ObjectArray(a.Ref())",
		},
		{
			"*WeakRef",
			CapiArgInfo{Name: "w", GoType: "*WeakRef"},
			"capi.WeakRef(w.Ref())",
		},
		{
			"typed array",
			CapiArgInfo{Name: "a", GoType: "*IntArray"},
			"capi.IntArray(a.Ref())",
		},
		{
			"MethodID",
			CapiArgInfo{Name: "mid", GoType: "MethodID"},
			"mid",
		},
		{
			"FieldID",
			CapiArgInfo{Name: "fid", GoType: "FieldID"},
			"fid",
		},
		{
			"int32",
			CapiArgInfo{Name: "x", GoType: "int32"},
			"capi.Jint(x)",
		},
		{
			"uint16",
			CapiArgInfo{Name: "c", GoType: "uint16"},
			"capi.Jchar(c)",
		},
		{
			"unsafe.Pointer void*",
			CapiArgInfo{Name: "p", GoType: "unsafe.Pointer", CType: "void*"},
			"p",
		},
		{
			"unsafe.Pointer jboolean*",
			CapiArgInfo{Name: "buf", GoType: "unsafe.Pointer", CType: "jboolean*"},
			"(*capi.Jboolean)(buf)",
		},
		{
			"unsafe.Pointer const jchar*",
			CapiArgInfo{Name: "chars", GoType: "unsafe.Pointer", CType: "const jchar*"},
			"(*capi.Jchar)(chars)",
		},
		{
			"unsafe.Pointer const char*",
			CapiArgInfo{Name: "utf", GoType: "unsafe.Pointer", CType: "const char*"},
			"(*capi.Cchar)(utf)",
		},
		{
			"unsafe.Pointer JNINativeMethod*",
			CapiArgInfo{Name: "m", GoType: "unsafe.Pointer", CType: "const JNINativeMethod*"},
			"(*capi.JNINativeMethod)(m)",
		},
		{
			"*uint8",
			CapiArgInfo{Name: "b", GoType: "*uint8"},
			"(*capi.Jboolean)(b)",
		},
		{
			"[]byte jbyte*",
			CapiArgInfo{Name: "buf", GoType: "[]byte", CType: "const jbyte*"},
			"(*capi.Jbyte)(unsafe.Pointer(&buf[0]))",
		},
		{
			"[]byte char*",
			CapiArgInfo{Name: "buf", GoType: "[]byte", CType: "char*"},
			"(*capi.Cchar)(unsafe.Pointer(&buf[0]))",
		},
		{
			"[]uint16",
			CapiArgInfo{Name: "chars", GoType: "[]uint16", CType: "jchar*"},
			"(*capi.Jchar)(unsafe.Pointer(&chars[0]))",
		},
		{
			"ObjectRefType",
			CapiArgInfo{Name: "rt", GoType: "ObjectRefType"},
			"rt",
		},
		{
			"unknown default",
			CapiArgInfo{Name: "x", GoType: "SomeUnknown"},
			"x",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := capiArgExpr(tt.arg)
			if got != tt.expect {
				t.Errorf("capiArgExpr(%+v) = %q, want %q", tt.arg, got, tt.expect)
			}
		})
	}
}

func TestBuildTransforms(t *testing.T) {
	params := []GoParam{
		{Name: "name", GoType: "string"},
		{Name: "flag", GoType: "bool"},
		{Name: "args", GoType: "...Value", IsVariadic: true},
		{Name: "count", GoType: "int32"},
	}
	transforms := buildTransforms(params)
	// Should have transforms for string, bool, and ...Value (3 total).
	if len(transforms) != 3 {
		t.Errorf("expected 3 transforms, got %d: %v", len(transforms), transforms)
	}
}

func TestBuildPostTransforms(t *testing.T) {
	got := buildPostTransforms([]GoParam{{Name: "x", GoType: "int32"}})
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestBuildRenderMethod(t *testing.T) {
	m := MergedMethod{
		GoName:         "FindClass",
		CapiCall:       "capi.FindClass",
		CheckException: true,
		Params: []GoParam{
			{Name: "name", GoType: "string"},
		},
		Returns: []GoReturn{
			{GoType: "*Class"},
			{IsError: true},
		},
		AllParams: []CapiArgInfo{
			{IsImplicit: true, GoType: "*Env", Implicit: "e.ptr"},
			{Name: "name", GoType: "string"},
		},
	}
	rm := buildRenderMethod(m)
	if rm.GoName != "FindClass" {
		t.Errorf("GoName = %q", rm.GoName)
	}
	if !rm.HasReturn {
		t.Error("expected HasReturn true")
	}
	if !rm.ReturnsError {
		t.Error("expected ReturnsError true")
	}
	if !rm.CheckException {
		t.Error("expected CheckException true")
	}
	if rm.JNIErrorTransform {
		t.Error("expected JNIErrorTransform false")
	}
}

func TestBuildRenderMethod_JNIErrorTransform(t *testing.T) {
	m := MergedMethod{
		GoName:   "Destroy",
		CapiCall: "capi.DestroyJavaVM",
		Returns: []GoReturn{
			{IsError: true, Transform: "jni_error"},
		},
		AllParams: []CapiArgInfo{
			{IsImplicit: true, GoType: "*VM", Implicit: "vm.ptr"},
		},
	}
	rm := buildRenderMethod(m)
	if !rm.JNIErrorTransform {
		t.Error("expected JNIErrorTransform true for error-only with jni_error transform")
	}
	if rm.HasReturn {
		t.Error("expected HasReturn false for error-only return")
	}
}

func TestBuildTypesData(t *testing.T) {
	merged := &MergedSpec{
		Constants: []MergedConstant{
			{CName: "JNI_VERSION_1_1", Value: "0x00010001"},
			{CName: "JNI_VERSION_1_6", Value: "0x00010006"},
			{CName: "JNI_OK", Value: "0"},
		},
	}
	data := BuildTypesData(merged)
	if len(data.VersionConstants) != 2 {
		t.Errorf("expected 2 version constants, got %d", len(data.VersionConstants))
	}
}

func TestBuildErrorData(t *testing.T) {
	merged := &MergedSpec{
		Constants: []MergedConstant{
			{CName: "JNI_OK", Value: "0"},
			{CName: "JNI_ERR", Value: "-1"},
			{CName: "JNI_EDETACHED", Value: "-2"},
			{CName: "JNI_VERSION_1_6", Value: "0x00010006"},
			{CName: "JNI_EVERSION", Value: "-3"},
		},
	}
	data := BuildErrorData(merged)
	// Should include JNI_OK, JNI_ERR, JNI_EDETACHED, JNI_EVERSION but not JNI_VERSION_1_6.
	if len(data.ErrorCodes) != 4 {
		t.Errorf("expected 4 error codes, got %d", len(data.ErrorCodes))
	}
}

func TestBuildEnvData(t *testing.T) {
	merged := &MergedSpec{
		EnvMethods: []MergedMethod{
			{
				GoName:   "GetVersion",
				CapiCall: "capi.GetVersion",
				Returns:  []GoReturn{{GoType: "int32"}},
				AllParams: []CapiArgInfo{
					{IsImplicit: true, GoType: "*Env", Implicit: "e.ptr"},
				},
			},
		},
	}
	data := BuildEnvData(merged)
	if len(data.Methods) != 1 {
		t.Errorf("expected 1 env method, got %d", len(data.Methods))
	}
}

func TestBuildVMData(t *testing.T) {
	merged := &MergedSpec{
		VMMethods: []MergedMethod{
			{
				GoName:   "Destroy",
				CapiCall: "capi.DestroyJavaVM",
				Returns:  []GoReturn{{IsError: true, Transform: "jni_error"}},
				AllParams: []CapiArgInfo{
					{IsImplicit: true, GoType: "*VM", Implicit: "vm.ptr"},
				},
			},
		},
	}
	data := BuildVMData(merged)
	if len(data.Methods) != 1 {
		t.Errorf("expected 1 vm method, got %d", len(data.Methods))
	}
}

func TestBuildValueData(t *testing.T) {
	merged := &MergedSpec{
		Primitives: []Primitive{
			{CType: "jint", GoType: "int32", Suffix: "Int"},
			{CType: "jboolean", GoType: "uint8", Suffix: "Boolean"},
		},
	}
	data := BuildValueData(merged)
	if len(data.Primitives) != 2 {
		t.Errorf("expected 2 primitives, got %d", len(data.Primitives))
	}
	if data.Primitives[0].CGoType != "capi.Jint" {
		t.Errorf("CGoType = %q, want capi.Jint", data.Primitives[0].CGoType)
	}
}

func TestBuildTypeData(t *testing.T) {
	rt := MergedRefType{
		CType:  "jobject",
		GoType: "Object",
	}
	data := BuildTypeData(rt)
	if data.GoName != "Object" {
		t.Errorf("GoName = %q", data.GoName)
	}
	if data.Parent != nil {
		t.Error("expected nil parent for root type")
	}

	// With parent.
	rtChild := MergedRefType{
		CType:  "jclass",
		GoType: "Class",
		Parent: "Object",
	}
	dataChild := BuildTypeData(rtChild)
	if dataChild.Parent == nil || dataChild.Parent.GoName != "Object" {
		t.Error("expected parent Object")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
