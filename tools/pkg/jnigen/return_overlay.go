package jnigen

// ReturnOverlay defines a Go-friendly return type override.
type ReturnOverlay struct {
	GoType    string `yaml:"go_type"`
	Transform string `yaml:"transform,omitempty"`
}
