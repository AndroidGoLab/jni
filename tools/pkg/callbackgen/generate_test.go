package callbackgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	specContent := `
callbacks:
  - class: com.example.Foo.Callback
    adapter: FooCallbackAdapter
    methods:
      - name: onEvent
        params:
          - type: com.example.Foo
            name: foo
          - type: int
            name: code
          - type: boolean
            name: ok

  - class: com.example.Bar.Listener
    adapter: BarListenerAdapter
    interface: true
    methods:
      - name: onBar
        params:
          - type: com.example.Bar
            name: bar
`
	specDir := t.TempDir()
	specPath := filepath.Join(specDir, "callbacks.yaml")
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatal(err)
	}

	outputDir := t.TempDir()
	if err := Generate(specPath, outputDir); err != nil {
		t.Fatal(err)
	}

	t.Run("extends_adapter", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(outputDir, "FooCallbackAdapter.java"))
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)

		for _, want := range []string{
			"DO NOT EDIT",
			"extends com.example.Foo.Callback",
			"public FooCallbackAdapter(long handlerID)",
			`GoAbstractDispatch.invoke(handlerID, "onEvent", new Object[]{foo, Integer.valueOf(code), Boolean.valueOf(ok)})`,
		} {
			if !strings.Contains(content, want) {
				t.Errorf("missing %q in output", want)
			}
		}
		if strings.Contains(content, "implements com.example.Foo.Callback") {
			t.Error("should not use 'implements' for class-based callback")
		}
	})

	t.Run("implements_adapter", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(outputDir, "BarListenerAdapter.java"))
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)

		for _, want := range []string{
			"implements com.example.Bar.Listener",
			`GoAbstractDispatch.invoke(handlerID, "onBar", new Object[]{bar})`,
		} {
			if !strings.Contains(content, want) {
				t.Errorf("missing %q in output", want)
			}
		}
		if strings.Contains(content, "extends com.example.Bar.Listener") {
			t.Error("should not use 'extends' for interface-based callback")
		}
	})
}

func TestBoxExpression(t *testing.T) {
	tests := []struct {
		javaType string
		param    string
		want     string
	}{
		{"int", "x", "Integer.valueOf(x)"},
		{"long", "val", "Long.valueOf(val)"},
		{"boolean", "flag", "Boolean.valueOf(flag)"},
		{"float", "f", "Float.valueOf(f)"},
		{"double", "d", "Double.valueOf(d)"},
		{"byte", "b", "Byte.valueOf(b)"},
		{"char", "c", "Character.valueOf(c)"},
		{"short", "s", "Short.valueOf(s)"},
		{"com.example.Foo", "foo", "foo"},
		{"android.hardware.camera2.CameraDevice", "cam", "cam"},
	}
	for _, tt := range tests {
		t.Run(tt.javaType+"_"+tt.param, func(t *testing.T) {
			got := BoxExpression(tt.javaType, tt.param)
			if got != tt.want {
				t.Errorf("BoxExpression(%q, %q) = %q, want %q", tt.javaType, tt.param, got, tt.want)
			}
		})
	}
}

func TestInheritanceKeyword(t *testing.T) {
	if got := inheritanceKeyword(CallbackEntry{Interface: false}); got != "extends" {
		t.Errorf("class: got %q, want extends", got)
	}
	if got := inheritanceKeyword(CallbackEntry{Interface: true}); got != "implements" {
		t.Errorf("interface: got %q, want implements", got)
	}
}

func TestRenderParamDecl(t *testing.T) {
	params := []ParamEntry{
		{Type: "android.hardware.camera2.CameraDevice", Name: "camera"},
		{Type: "int", Name: "error"},
	}
	got := renderParamDecl(params)
	want := "android.hardware.camera2.CameraDevice camera, int error"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderArgsList(t *testing.T) {
	params := []ParamEntry{
		{Type: "android.hardware.camera2.CameraDevice", Name: "camera"},
		{Type: "int", Name: "error"},
	}
	got := renderArgsList(params)
	want := "camera, Integer.valueOf(error)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGenerateAllPrimitiveBoxing(t *testing.T) {
	specContent := `
callbacks:
  - class: test.AllPrimitives
    adapter: AllPrimitivesAdapter
    methods:
      - name: onAll
        params:
          - type: int
            name: i
          - type: long
            name: l
          - type: boolean
            name: b
          - type: float
            name: f
          - type: double
            name: d
          - type: byte
            name: by
          - type: char
            name: c
          - type: short
            name: s
`
	specDir := t.TempDir()
	specPath := filepath.Join(specDir, "callbacks.yaml")
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatal(err)
	}

	outputDir := t.TempDir()
	if err := Generate(specPath, outputDir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "AllPrimitivesAdapter.java"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	expectedBoxed := strings.Join([]string{
		"Integer.valueOf(i)",
		"Long.valueOf(l)",
		"Boolean.valueOf(b)",
		"Float.valueOf(f)",
		"Double.valueOf(d)",
		"Byte.valueOf(by)",
		"Character.valueOf(c)",
		"Short.valueOf(s)",
	}, ", ")
	if !strings.Contains(content, expectedBoxed) {
		t.Errorf("missing boxed args %q in output", expectedBoxed)
	}
}
