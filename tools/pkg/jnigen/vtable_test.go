package jnigen

import (
	"strings"
	"testing"
)

func TestGenerateVtableHelpers_NonVoid(t *testing.T) {
	funcs := []MergedCapiFunc{
		{
			CName:      "FindClass",
			HelperName: "jni_FindClass",
			Vtable:     "JNIEnv",
			Params: []CapiParam{
				{CType: "JNIEnv*", CName: "env"},
				{CType: "const char*", CName: "p0"},
			},
			Returns: "jclass",
			IsVoid:  false,
		},
	}

	output := GenerateVtableHelpers(funcs)

	if !strings.Contains(output, "static inline jclass jni_FindClass") {
		t.Error("missing function signature")
	}
	if !strings.Contains(output, "return (*env)->FindClass(env, p0)") {
		t.Error("missing return statement with vtable call")
	}
}

func TestGenerateVtableHelpers_Void(t *testing.T) {
	funcs := []MergedCapiFunc{
		{
			CName:      "ExceptionClear",
			HelperName: "jni_ExceptionClear",
			Vtable:     "JNIEnv",
			Params: []CapiParam{
				{CType: "JNIEnv*", CName: "env"},
			},
			Returns: "void",
			IsVoid:  true,
		},
	}

	output := GenerateVtableHelpers(funcs)

	if !strings.Contains(output, "static inline void jni_ExceptionClear") {
		t.Error("missing function signature")
	}
	if strings.Contains(output, "return") && !strings.Contains(output, "return (*") {
		t.Error("void function should not have bare 'return' statements")
	}
	if !strings.Contains(output, "(*env)->ExceptionClear(env)") {
		t.Error("missing vtable call")
	}
}

func TestGenerateVtableHelpers_JavaVM(t *testing.T) {
	funcs := []MergedCapiFunc{
		{
			CName:      "DestroyJavaVM",
			HelperName: "jni_DestroyJavaVM",
			Vtable:     "JavaVM",
			Params: []CapiParam{
				{CType: "JavaVM*", CName: "vm"},
			},
			Returns: "jint",
			IsVoid:  false,
		},
	}

	output := GenerateVtableHelpers(funcs)

	if !strings.Contains(output, "static inline jint jni_DestroyJavaVM") {
		t.Error("missing function signature")
	}
	if !strings.Contains(output, "return (*vm)->DestroyJavaVM(vm)") {
		t.Error("missing JavaVM vtable call pattern")
	}
}

func TestGenerateVtableHelpers_IncludeGuard(t *testing.T) {
	output := GenerateVtableHelpers(nil)

	if !strings.Contains(output, "#ifndef JNI_CGO_VTABLE_DISPATCH_H") {
		t.Error("missing include guard")
	}
	if !strings.Contains(output, "#define JNI_CGO_VTABLE_DISPATCH_H") {
		t.Error("missing include guard define")
	}
	if !strings.Contains(output, "#endif") {
		t.Error("missing include guard endif")
	}
	if !strings.Contains(output, "#include <jni.h>") {
		t.Error("missing jni.h include")
	}
}
