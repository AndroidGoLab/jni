package javagen

import (
	"strings"
	"testing"
)

func TestMerge_BasicClass(t *testing.T) {
	spec := &Spec{
		Package:  "test",
		GoImport: "github.com/example/test",
		Classes: []Class{
			{
				JavaClass:   "com.example.Foo",
				GoType:      "Foo",
				Obtain:      "system_service",
				ServiceName: "foo",
				Close:       true,
				Methods: []Method{
					{
						JavaMethod: "doSomething",
						GoName:     "DoSomething",
						Params: []Param{
							{JavaType: "String", GoName: "name"},
						},
						Returns: "int",
						Error:   true,
					},
				},
			},
		},
	}

	merged, err := Merge(spec, &Overlay{})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if merged.Package != "test" {
		t.Errorf("package = %q, want test", merged.Package)
	}
	if len(merged.Classes) != 1 {
		t.Fatalf("expected 1 class, got %d", len(merged.Classes))
	}

	cls := merged.Classes[0]
	if cls.JavaClassSlash != "com/example/Foo" {
		t.Errorf("JavaClassSlash = %q", cls.JavaClassSlash)
	}
	if cls.GoType != "Foo" {
		t.Errorf("GoType = %q", cls.GoType)
	}
	if !cls.Close {
		t.Error("expected Close=true")
	}
	if len(cls.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(cls.Methods))
	}

	m := cls.Methods[0]
	if m.GoName != "DoSomething" {
		t.Errorf("GoName = %q", m.GoName)
	}
	if m.JNISig == "" {
		t.Error("JNISig is empty")
	}
	if m.CallSuffix != "Int" {
		t.Errorf("CallSuffix = %q, want Int", m.CallSuffix)
	}
	if !m.HasError {
		t.Error("expected HasError=true")
	}
}

func TestMerge_DataClass(t *testing.T) {
	spec := &Spec{
		Package:  "test",
		GoImport: "github.com/example/test",
		Classes: []Class{
			{
				JavaClass: "com.example.Point",
				GoType:    "Point",
				Kind:      "data_class",
				Fields: []Field{
					{JavaMethod: "getX", Returns: "double", GoName: "X", GoType: "float64"},
					{JavaMethod: "getY", Returns: "double", GoName: "Y", GoType: "float64"},
				},
			},
		},
	}

	merged, err := Merge(spec, &Overlay{})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if len(merged.DataClasses) != 1 {
		t.Fatalf("expected 1 data class, got %d", len(merged.DataClasses))
	}
	if len(merged.Classes) != 0 {
		t.Errorf("expected 0 regular classes, got %d", len(merged.Classes))
	}

	dc := merged.DataClasses[0]
	if dc.GoType != "Point" {
		t.Errorf("GoType = %q", dc.GoType)
	}
	if len(dc.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(dc.Fields))
	}
	if dc.Fields[0].GoName != "X" {
		t.Errorf("field[0] GoName = %q", dc.Fields[0].GoName)
	}
	if dc.Fields[0].CallSuffix != "Double" {
		t.Errorf("field[0] CallSuffix = %q", dc.Fields[0].CallSuffix)
	}
}

func TestMerge_Callbacks(t *testing.T) {
	spec := &Spec{
		Package:  "test",
		GoImport: "github.com/example/test",
		Callbacks: []Callback{
			{
				JavaInterface: "com.example.Listener",
				GoType:        "Listener",
				Methods: []CallbackMethod{
					{
						JavaMethod: "onEvent",
						Params:     []string{"String", "int"},
						GoField:    "OnEvent",
					},
					{
						JavaMethod: "onDone",
						Params:     []string{},
						GoField:    "OnDone",
					},
				},
			},
		},
	}

	merged, err := Merge(spec, &Overlay{})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if len(merged.Callbacks) != 1 {
		t.Fatalf("expected 1 callback, got %d", len(merged.Callbacks))
	}
	cb := merged.Callbacks[0]
	if cb.GoType != "Listener" {
		t.Errorf("GoType = %q", cb.GoType)
	}
	if len(cb.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(cb.Methods))
	}
	if cb.Methods[0].GoField != "OnEvent" {
		t.Errorf("method[0] GoField = %q", cb.Methods[0].GoField)
	}
	if len(cb.Methods[0].Params) != 2 {
		t.Errorf("method[0] params: expected 2, got %d", len(cb.Methods[0].Params))
	}
	// First param is String -> should have IsString=true
	if !cb.Methods[0].Params[0].IsString {
		t.Error("expected String param to be IsString")
	}
	if cb.Methods[1].GoParams != "" {
		t.Errorf("method[1] GoParams = %q, want empty", cb.Methods[1].GoParams)
	}
}

func TestMerge_Constants(t *testing.T) {
	spec := &Spec{
		Package:  "test",
		GoImport: "github.com/example/test",
		Constants: []Constant{
			{GoName: "A", Value: "1", GoType: "int"},
			{GoName: "B", Value: "2", GoType: "int"},
			{GoName: "X", Value: `"hello"`, GoType: "string"},
		},
	}

	merged, err := Merge(spec, &Overlay{})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if len(merged.ConstantGroups) != 2 {
		t.Fatalf("expected 2 constant groups, got %d", len(merged.ConstantGroups))
	}

	// First group: int is a builtin, so GoType is empty (no named type).
	g := merged.ConstantGroups[0]
	if g.GoType != "" {
		t.Errorf("group[0] GoType = %q, want empty (builtin)", g.GoType)
	}
	if g.BaseType != "int" {
		t.Errorf("group[0] BaseType = %q, want int", g.BaseType)
	}
	if len(g.Values) != 2 {
		t.Errorf("group[0] values: expected 2, got %d", len(g.Values))
	}

	// Second group: string is a builtin, so GoType is empty (no named type).
	g2 := merged.ConstantGroups[1]
	if g2.GoType != "" {
		t.Errorf("group[1] GoType = %q, want empty (builtin)", g2.GoType)
	}
	if g2.BaseType != "string" {
		t.Errorf("group[1] BaseType = %q, want string", g2.BaseType)
	}
	if len(g2.Values) != 1 {
		t.Errorf("group[1] values: expected 1, got %d", len(g2.Values))
	}
}

func TestMerge_UntypedConstants(t *testing.T) {
	spec := &Spec{
		Package:  "test",
		GoImport: "github.com/example/test",
		Constants: []Constant{
			{GoName: "Foo", Value: "42"},
		},
	}

	merged, err := Merge(spec, &Overlay{})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if len(merged.ConstantGroups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(merged.ConstantGroups))
	}
	if merged.ConstantGroups[0].GoType != "" {
		t.Errorf("expected empty GoType for untyped, got %q", merged.ConstantGroups[0].GoType)
	}
}

func TestMerge_NamedTypeConstants(t *testing.T) {
	spec := &Spec{
		Package:  "test",
		GoImport: "github.com/example/test",
		Constants: []Constant{
			{GoName: "WiFi", Value: "1", GoType: "Transport"},
			{GoName: "Cell", Value: "2", GoType: "Transport"},
		},
	}

	merged, err := Merge(spec, &Overlay{})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if len(merged.ConstantGroups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(merged.ConstantGroups))
	}
	g := merged.ConstantGroups[0]
	if g.GoType != "Transport" {
		t.Errorf("GoType = %q, want Transport", g.GoType)
	}
	// BaseType is inferred from the value "1" → "int" since GoType == BaseType (self-referential).
	if g.BaseType != "int" {
		t.Errorf("BaseType = %q, want int", g.BaseType)
	}
}

func TestIsBuiltinType(t *testing.T) {
	builtins := []string{"string", "int", "int32", "int64", "float32", "float64", "bool", "byte", "rune", "uint32"}
	for _, b := range builtins {
		if !isBuiltinType(b) {
			t.Errorf("isBuiltinType(%q) = false, want true", b)
		}
	}
	nonBuiltins := []string{"Transport", "Status", "Provider", "RecordType", ""}
	for _, nb := range nonBuiltins {
		if isBuiltinType(nb) {
			t.Errorf("isBuiltinType(%q) = true, want false", nb)
		}
	}
}

func TestMerge_WithOverlay(t *testing.T) {
	spec := &Spec{
		Package:  "test",
		GoImport: "github.com/example/test",
		Classes: []Class{
			{
				JavaClass:   "com.example.Svc",
				GoType:      "Svc",
				Obtain:      "system_service",
				ServiceName: "svc",
				Methods: []Method{
					{
						JavaMethod: "doIt",
						GoName:     "DoIt",
						Params:     []Param{{JavaType: "String", GoName: "name"}},
						Returns:    "void",
						Error:      true,
					},
				},
			},
		},
	}

	overlay := &Overlay{
		GoNameOverrides: map[string]string{
			"doIt": "Execute",
		},
		TypeOverrides: map[string]string{},
	}

	merged, err := Merge(spec, overlay)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if merged.Classes[0].Methods[0].GoName != "Execute" {
		t.Errorf("expected overridden GoName Execute, got %q", merged.Classes[0].Methods[0].GoName)
	}
}

func TestMerge_ExtraMethodsFromOverlay(t *testing.T) {
	spec := &Spec{
		Package:  "test",
		GoImport: "github.com/example/test",
		Classes: []Class{
			{
				JavaClass:   "com.example.Svc",
				GoType:      "Svc",
				Obtain:      "system_service",
				ServiceName: "svc",
			},
		},
	}

	overlay := &Overlay{
		ExtraMethods: []Method{
			{
				JavaMethod: "extraMethod",
				GoName:     "ExtraMethod",
				Returns:    "void",
			},
		},
	}

	merged, err := Merge(spec, overlay)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if len(merged.Classes[0].Methods) != 1 {
		t.Fatalf("expected 1 method (from overlay), got %d", len(merged.Classes[0].Methods))
	}
	if merged.Classes[0].Methods[0].GoName != "ExtraMethod" {
		t.Errorf("expected ExtraMethod, got %q", merged.Classes[0].Methods[0].GoName)
	}
}

func TestBuildGoReturnList(t *testing.T) {
	tests := []struct {
		goReturn string
		isVoid   bool
		hasError bool
		want     string
	}{
		{"", true, true, "error"},
		{"", true, false, ""},
		{"int32", false, true, "(int32, error)"},
		{"int32", false, false, "int32"},
	}
	for _, tt := range tests {
		got := buildGoReturnList(tt.goReturn, tt.isVoid, tt.hasError)
		if got != tt.want {
			t.Errorf("buildGoReturnList(%q, %v, %v) = %q, want %q",
				tt.goReturn, tt.isVoid, tt.hasError, got, tt.want)
		}
	}
}

func TestBuildGoReturnVars(t *testing.T) {
	if got := buildGoReturnVars("int32", false); got != "var result int32" {
		t.Errorf("buildGoReturnVars = %q", got)
	}
	if got := buildGoReturnVars("", true); got != "" {
		t.Errorf("buildGoReturnVars void = %q", got)
	}
}

func TestBuildGoReturnValues(t *testing.T) {
	tests := []struct {
		goReturn string
		isVoid   bool
		hasError bool
		want     string
	}{
		{"", true, true, "callErr"},
		{"", true, false, ""},
		{"int32", false, true, "result, callErr"},
		{"int32", false, false, "result"},
	}
	for _, tt := range tests {
		got := buildGoReturnValues(tt.goReturn, tt.isVoid, tt.hasError)
		if got != tt.want {
			t.Errorf("buildGoReturnValues(%q, %v, %v) = %q, want %q",
				tt.goReturn, tt.isVoid, tt.hasError, got, tt.want)
		}
	}
}

func TestSanitizeGoName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"type", "type_"},
		{"var", "var_"},
		{"range", "range_"},
		{"name", "name"},
		{"provider", "provider"},
	}
	for _, tt := range tests {
		got := sanitizeGoName(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeGoName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBuildGoParamList(t *testing.T) {
	params := []MergedParam{
		{GoName: "name", GoType: "string"},
		{GoName: "count", GoType: "int32"},
	}
	got := buildGoParamList(params)
	if got != "name string, count int32" {
		t.Errorf("buildGoParamList = %q", got)
	}

	got = buildGoParamList(nil)
	if got != "" {
		t.Errorf("buildGoParamList nil = %q", got)
	}
}

func TestBuildJNIArgs(t *testing.T) {
	// Empty params should return empty.
	if got := buildJNIArgs(nil); got != "" {
		t.Errorf("buildJNIArgs nil = %q", got)
	}

	params := []MergedParam{
		{JavaType: "String", GoName: "name", GoType: "string", IsString: true},
		{JavaType: "int", GoName: "count", GoType: "int32"},
	}
	got := buildJNIArgs(params)
	if got == "" {
		t.Error("buildJNIArgs should not be empty")
	}
	if !strings.Contains(got, "jni.ObjectValue(&jName.Object)") {
		t.Errorf("expected ObjectValue for String param, got %q", got)
	}
	if !strings.Contains(got, "jni.IntValue(count)") {
		t.Errorf("expected IntValue for int param, got %q", got)
	}
}
