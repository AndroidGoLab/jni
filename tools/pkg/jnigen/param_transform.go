package jnigen

// ParamTransform defines how to convert between Go and C types.
type ParamTransform struct {
	Description string `yaml:"description"`
	GoType      string `yaml:"go_type"`
	CType       string `yaml:"c_type"`
	ToC         string `yaml:"to_c,omitempty"`
	CArg        string `yaml:"c_arg,omitempty"`
	FromC       string `yaml:"from_c,omitempty"`
}
