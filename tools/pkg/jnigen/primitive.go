package jnigen

// Primitive represents a JNI primitive type.
type Primitive struct {
	CType  string `yaml:"c_type"`
	GoType string `yaml:"go_type"`
	Suffix string `yaml:"suffix"`
}
