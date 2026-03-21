package jnigen

// OpaqueType represents an opaque JNI type (method/field IDs).
type OpaqueType struct {
	CType  string `yaml:"c_type"`
	GoType string `yaml:"go_type,omitempty"`
}
