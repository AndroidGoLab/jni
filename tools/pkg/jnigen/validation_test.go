package jnigen

import (
	"testing"
)

func TestValidateSpec_NoPrimitives(t *testing.T) {
	path := writeTemp(t, `
primitives: []
reference_types: []
opaque_types: []
function_families: []
env_functions: []
vm_functions: []
constants: []
`)
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for empty primitives")
	}
}

func TestValidateSpec_MissingCType(t *testing.T) {
	path := writeTemp(t, `
primitives:
  - { c_type: "", go_type: uint8, suffix: Boolean }
reference_types: []
opaque_types: []
function_families: []
env_functions: []
vm_functions: []
constants: []
`)
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for missing c_type")
	}
}

func TestValidateSpec_MissingGoType(t *testing.T) {
	path := writeTemp(t, `
primitives:
  - { c_type: jboolean, go_type: "", suffix: Boolean }
reference_types: []
opaque_types: []
function_families: []
env_functions: []
vm_functions: []
constants: []
`)
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for missing go_type")
	}
}

func TestValidateSpec_MissingRefCType(t *testing.T) {
	path := writeTemp(t, `
primitives:
  - { c_type: jint, go_type: int32, suffix: Int }
reference_types:
  - { c_type: "", parent: ~ }
opaque_types: []
function_families: []
env_functions: []
vm_functions: []
constants: []
`)
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for missing reference type c_type")
	}
}

func TestValidateSpec_MissingFamilyPattern(t *testing.T) {
	path := writeTemp(t, `
primitives:
  - { c_type: jint, go_type: int32, suffix: Int }
reference_types: []
opaque_types: []
function_families:
  - { pattern: "", vtable: JNIEnv, expand: [Int] }
env_functions: []
vm_functions: []
constants: []
`)
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for missing family pattern")
	}
}

func TestValidateSpec_MissingFamilyVtable(t *testing.T) {
	path := writeTemp(t, `
primitives:
  - { c_type: jint, go_type: int32, suffix: Int }
reference_types: []
opaque_types: []
function_families:
  - { pattern: "Call{Type}MethodA", vtable: "", expand: [Int] }
env_functions: []
vm_functions: []
constants: []
`)
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for missing family vtable")
	}
}

func TestValidateSpec_MissingFamilyExpand(t *testing.T) {
	path := writeTemp(t, `
primitives:
  - { c_type: jint, go_type: int32, suffix: Int }
reference_types: []
opaque_types: []
function_families:
  - { pattern: "Call{Type}MethodA", vtable: JNIEnv, expand: [] }
env_functions: []
vm_functions: []
constants: []
`)
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for empty expand list")
	}
}

func TestValidateSpec_MissingEnvFuncName(t *testing.T) {
	path := writeTemp(t, `
primitives:
  - { c_type: jint, go_type: int32, suffix: Int }
reference_types: []
opaque_types: []
function_families: []
env_functions:
  - { name: "" }
vm_functions: []
constants: []
`)
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for missing env function name")
	}
}

func TestValidateSpec_MissingVMFuncName(t *testing.T) {
	path := writeTemp(t, `
primitives:
  - { c_type: jint, go_type: int32, suffix: Int }
reference_types: []
opaque_types: []
function_families: []
env_functions: []
vm_functions:
  - { name: "" }
constants: []
`)
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for missing vm function name")
	}
}

func TestValidateSpec_MissingConstantName(t *testing.T) {
	path := writeTemp(t, `
primitives:
  - { c_type: jint, go_type: int32, suffix: Int }
reference_types: []
opaque_types: []
function_families: []
env_functions: []
vm_functions: []
constants:
  - { name: "", value: 0, go_type: int32 }
`)
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for missing constant name")
	}
}

func TestFamilyReturnMap_StringReturn(t *testing.T) {
	fo := &FamilyOverlay{
		Returns: "*Object",
	}
	retMap := fo.FamilyReturnMap()
	if retMap == nil {
		t.Fatal("expected non-nil return map for string return")
	}
	if retMap["_all"] != "*Object" {
		t.Errorf("expected _all = *Object, got %q", retMap["_all"])
	}
}

func TestFamilyReturnMap_NilReturn(t *testing.T) {
	fo := &FamilyOverlay{
		Returns: nil,
	}
	retMap := fo.FamilyReturnMap()
	if retMap != nil {
		t.Error("expected nil return map for nil returns")
	}
}

func TestCamelCase_Empty(t *testing.T) {
	if got := camelCase(""); got != "" {
		t.Errorf("camelCase('') = %q", got)
	}
}

func TestLowerCamel_Empty(t *testing.T) {
	if got := lowerCamel(""); got != "" {
		t.Errorf("lowerCamel('') = %q", got)
	}
}

func TestCapitalizeFirst_Empty(t *testing.T) {
	if got := capitalizeFirst(""); got != "" {
		t.Errorf("capitalizeFirst('') = %q", got)
	}
}

func TestRender_NoTemplates(t *testing.T) {
	merged := &MergedSpec{TypeMethods: make(map[string][]MergedMethod)}
	dir := t.TempDir()
	tmplDir := dir + "/nonexistent"
	err := Render(merged, tmplDir, dir)
	if err == nil {
		t.Fatal("expected error for missing template directory")
	}
}

func TestMerge_VMFunctionMissingOverlay(t *testing.T) {
	spec := &Spec{
		Primitives: []Primitive{{CType: "jint", GoType: "int32", Suffix: "Int"}},
		VMFunctions: []Function{
			{Name: "SomeVMFunc", Returns: "jint"},
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
		t.Fatal("expected error for missing VM function overlay")
	}
}
