package callbackgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, os.WriteFile(specPath, []byte(specContent), 0o644))

	outputDir := t.TempDir()
	require.NoError(t, Generate(specPath, outputDir))

	t.Run("extends_adapter", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(outputDir, "FooCallbackAdapter.java"))
		require.NoError(t, err)
		content := string(data)

		assert.Contains(t, content, "DO NOT EDIT")
		assert.Contains(t, content, "extends com.example.Foo.Callback")
		assert.Contains(t, content, "public FooCallbackAdapter(long handlerID)")
		assert.Contains(t, content, `GoAbstractDispatch.invoke(handlerID, "onEvent", new Object[]{foo, Integer.valueOf(code), Boolean.valueOf(ok)})`)

		// Verify it does NOT use "implements" for a class-based callback.
		assert.NotContains(t, content, "implements com.example.Foo.Callback")
	})

	t.Run("implements_adapter", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(outputDir, "BarListenerAdapter.java"))
		require.NoError(t, err)
		content := string(data)

		assert.Contains(t, content, "implements com.example.Bar.Listener")
		assert.Contains(t, content, `GoAbstractDispatch.invoke(handlerID, "onBar", new Object[]{bar})`)

		// Verify it does NOT use "extends" for an interface-based callback.
		assert.NotContains(t, content, "extends com.example.Bar.Listener")
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
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInheritanceKeyword(t *testing.T) {
	t.Run("class_uses_extends", func(t *testing.T) {
		entry := CallbackEntry{Interface: false}
		assert.Equal(t, "extends", inheritanceKeyword(entry))
	})

	t.Run("interface_uses_implements", func(t *testing.T) {
		entry := CallbackEntry{Interface: true}
		assert.Equal(t, "implements", inheritanceKeyword(entry))
	})
}

func TestRenderParamDecl(t *testing.T) {
	params := []ParamEntry{
		{Type: "android.hardware.camera2.CameraDevice", Name: "camera"},
		{Type: "int", Name: "error"},
	}
	got := renderParamDecl(params)
	assert.Equal(t, "android.hardware.camera2.CameraDevice camera, int error", got)
}

func TestRenderArgsList(t *testing.T) {
	params := []ParamEntry{
		{Type: "android.hardware.camera2.CameraDevice", Name: "camera"},
		{Type: "int", Name: "error"},
	}
	got := renderArgsList(params)
	assert.Equal(t, "camera, Integer.valueOf(error)", got)
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
	require.NoError(t, os.WriteFile(specPath, []byte(specContent), 0o644))

	outputDir := t.TempDir()
	require.NoError(t, Generate(specPath, outputDir))

	data, err := os.ReadFile(filepath.Join(outputDir, "AllPrimitivesAdapter.java"))
	require.NoError(t, err)
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
	assert.Contains(t, content, expectedBoxed)
}
