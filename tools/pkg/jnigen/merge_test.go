package jnigen

import (
	"testing"
)

func TestMerge_FamilyExpansion(t *testing.T) {
	spec := &Spec{
		Primitives: []Primitive{
			{CType: "jint", GoType: "int32", Suffix: "Int"},
			{CType: "jboolean", GoType: "uint8", Suffix: "Boolean"},
		},
		FunctionFamilies: []FunctionFamily{
			{
				Pattern: "Call{Type}MethodA",
				Vtable:  "JNIEnv",
				Params:  []string{"jobject", "jmethodID", "const jvalue*"},
				Returns: "{type}",
				Expand:  []string{"Int", "Void"},
			},
		},
	}

	overlay := &Overlay{
		TypeRenames: map[string]string{
			"jobject": "Object",
			"jint":    "int32",
		},
		Receivers: map[string]string{
			"env_functions":     "*Env",
			"function_families": "*Env",
			"vm_functions":      "*VM",
		},
		Functions: map[string]FuncOverlay{},
		FamilyOverlays: map[string]FamilyOverlay{
			"Call{Type}MethodA": {
				GoPattern:      "Call{Type}Method",
				CheckException: true,
				Params: []ParamOverlay{
					{Name: "obj", CType: "jobject", GoType: "*Object"},
					{Name: "method", CType: "jmethodID", GoType: "MethodID"},
					{Name: "args", CType: "const jvalue*", GoType: "...Value"},
				},
				Returns: map[string]interface{}{
					"primitive": "{go_type}",
					"Object":    "*Object",
					"Void":      nil,
				},
			},
		},
		ParamTransforms: map[string]ParamTransform{},
	}

	merged, err := Merge(spec, overlay)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should have 2 capi functions from the family.
	if len(merged.CapiFunctions) != 2 {
		t.Fatalf("expected 2 capi functions, got %d", len(merged.CapiFunctions))
	}

	// Check that we have CallIntMethodA and CallVoidMethodA.
	names := make(map[string]bool)
	for _, f := range merged.CapiFunctions {
		names[f.CName] = true
	}
	if !names["CallIntMethodA"] {
		t.Error("missing CallIntMethodA")
	}
	if !names["CallVoidMethodA"] {
		t.Error("missing CallVoidMethodA")
	}

	// Check idiomatic methods.
	if len(merged.EnvMethods) != 2 {
		t.Fatalf("expected 2 env methods, got %d", len(merged.EnvMethods))
	}

	methodNames := make(map[string]bool)
	for _, m := range merged.EnvMethods {
		methodNames[m.GoName] = true
	}
	if !methodNames["CallIntMethod"] {
		t.Error("missing CallIntMethod")
	}
	if !methodNames["CallVoidMethod"] {
		t.Error("missing CallVoidMethod")
	}
}

func TestMerge_TypeExpansionPlaceholders(t *testing.T) {
	spec := &Spec{
		Primitives: []Primitive{
			{CType: "jint", GoType: "int32", Suffix: "Int"},
		},
		FunctionFamilies: []FunctionFamily{
			{
				Pattern: "Get{Type}Field",
				Vtable:  "JNIEnv",
				Params:  []string{"jobject", "jfieldID"},
				Returns: "{type}",
				Expand:  []string{"Int", "Object"},
			},
		},
	}

	overlay := &Overlay{
		TypeRenames: map[string]string{},
		Receivers: map[string]string{
			"env_functions":     "*Env",
			"function_families": "*Env",
			"vm_functions":      "*VM",
		},
		Functions: map[string]FuncOverlay{},
		FamilyOverlays: map[string]FamilyOverlay{
			"Get{Type}Field": {
				GoPattern: "Get{Type}Field",
				Params: []ParamOverlay{
					{Name: "obj", CType: "jobject", GoType: "*Object"},
					{Name: "field", CType: "jfieldID", GoType: "FieldID"},
				},
				Returns: map[string]interface{}{
					"primitive": "{go_type}",
					"Object":    "*Object",
				},
			},
		},
		ParamTransforms: map[string]ParamTransform{},
	}

	merged, err := Merge(spec, overlay)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Verify C names have {Type} replaced.
	for _, f := range merged.CapiFunctions {
		if f.CName == "GetIntField" {
			if f.Returns != "jint" {
				t.Errorf("GetIntField returns: expected jint, got %s", f.Returns)
			}
		}
		if f.CName == "GetObjectField" {
			if f.Returns != "jobject" {
				t.Errorf("GetObjectField returns: expected jobject, got %s", f.Returns)
			}
		}
	}
}

func TestMerge_MissingOverlay(t *testing.T) {
	spec := &Spec{
		Primitives: []Primitive{
			{CType: "jint", GoType: "int32", Suffix: "Int"},
		},
		EnvFunctions: []Function{
			{Name: "GetVersion", Returns: "jint"},
		},
	}

	overlay := &Overlay{
		TypeRenames:     map[string]string{},
		Receivers:       map[string]string{"env_functions": "*Env", "vm_functions": "*VM"},
		Functions:       map[string]FuncOverlay{},
		FamilyOverlays:  map[string]FamilyOverlay{},
		ParamTransforms: map[string]ParamTransform{},
	}

	_, err := Merge(spec, overlay)
	if err == nil {
		t.Fatal("expected error for missing overlay")
	}
}

func TestMerge_ExceptionAnnotation(t *testing.T) {
	spec := &Spec{
		Primitives: []Primitive{
			{CType: "jint", GoType: "int32", Suffix: "Int"},
		},
		EnvFunctions: []Function{
			{Name: "FindClass", Params: []string{"const char*"}, Returns: "jclass", Exception: true},
		},
	}

	overlay := &Overlay{
		TypeRenames: map[string]string{},
		Receivers:   map[string]string{"env_functions": "*Env", "vm_functions": "*VM"},
		Functions: map[string]FuncOverlay{
			"FindClass": {
				GoName:         "FindClass",
				CheckException: true,
				Params:         []ParamOverlay{{Name: "name", CType: "const char*", GoType: "string"}},
				Returns:        &ReturnOverlay{GoType: "*Class"},
			},
		},
		FamilyOverlays:  map[string]FamilyOverlay{},
		ParamTransforms: map[string]ParamTransform{},
	}

	merged, err := Merge(spec, overlay)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(merged.EnvMethods) != 1 {
		t.Fatalf("expected 1 env method, got %d", len(merged.EnvMethods))
	}

	m := merged.EnvMethods[0]
	if !m.CheckException {
		t.Error("expected check_exception=true")
	}

	// Should have error in returns.
	hasError := false
	for _, r := range m.Returns {
		if r.IsError {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected error return for exception-checking method")
	}
}

func TestMerge_VoidReturn(t *testing.T) {
	spec := &Spec{
		Primitives: []Primitive{},
		EnvFunctions: []Function{
			{Name: "ExceptionClear", Returns: "void"},
		},
	}

	overlay := &Overlay{
		TypeRenames: map[string]string{},
		Receivers:   map[string]string{"env_functions": "*Env", "vm_functions": "*VM"},
		Functions: map[string]FuncOverlay{
			"ExceptionClear": {GoName: "ExceptionClear"},
		},
		FamilyOverlays:  map[string]FamilyOverlay{},
		ParamTransforms: map[string]ParamTransform{},
	}

	merged, err := Merge(spec, overlay)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Find the capi function and check it's void.
	for _, f := range merged.CapiFunctions {
		if f.CName == "ExceptionClear" {
			if !f.IsVoid {
				t.Error("ExceptionClear should be void")
			}
			return
		}
	}
	t.Error("ExceptionClear not found in capi functions")
}
