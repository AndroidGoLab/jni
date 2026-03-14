package jnigen

import (
	"testing"
)

func TestBuildCapiData(t *testing.T) {
	merged := &MergedSpec{
		Primitives: []Primitive{
			{CType: "jint", GoType: "int32", Suffix: "Int"},
		},
		ReferenceTypes: []MergedRefType{
			{CType: "jobject", GoType: "Object"},
		},
		OpaqueTypes: []MergedOpaqueType{
			{CType: "jmethodID", GoType: "MethodID"},
		},
		Constants: []MergedConstant{
			{CName: "JNI_OK", GoName: "JNI_OK", Value: "0", GoType: "int32"},
		},
		CapiFunctions: []MergedCapiFunc{
			{CName: "FindClass", Vtable: "JNIEnv"},
			{CName: "DestroyJavaVM", Vtable: "JavaVM"},
		},
		TypeMethods: make(map[string][]MergedMethod),
	}

	data := BuildCapiData(merged)

	if len(data.EnvFunctions) != 1 {
		t.Errorf("expected 1 env function, got %d", len(data.EnvFunctions))
	}
	if len(data.VMFunctions) != 1 {
		t.Errorf("expected 1 vm function, got %d", len(data.VMFunctions))
	}
	if len(data.AllFunctions) != 2 {
		t.Errorf("expected 2 total functions, got %d", len(data.AllFunctions))
	}
	if len(data.Types) != 1 {
		t.Errorf("expected 1 type, got %d", len(data.Types))
	}
	if len(data.OpaqueTypes) != 1 {
		t.Errorf("expected 1 opaque type, got %d", len(data.OpaqueTypes))
	}
	if len(data.Constants) != 1 {
		t.Errorf("expected 1 constant, got %d", len(data.Constants))
	}
	if len(data.Primitives) != 1 {
		t.Errorf("expected 1 primitive, got %d", len(data.Primitives))
	}
}
