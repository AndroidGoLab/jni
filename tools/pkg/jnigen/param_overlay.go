package jnigen

// ParamOverlay defines a Go-friendly parameter override.
type ParamOverlay struct {
	Name      string `yaml:"name"`
	CType     string `yaml:"c_type"`
	GoType    string `yaml:"go_type"`
	Transform string `yaml:"transform,omitempty"`
	Implicit  string `yaml:"implicit,omitempty"`
}
