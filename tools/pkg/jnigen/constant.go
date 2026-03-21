package jnigen

// Constant represents a JNI constant.
type Constant struct {
	Name   string `yaml:"name"`
	Value  string `yaml:"value"`
	GoType string `yaml:"go_type"`
}
