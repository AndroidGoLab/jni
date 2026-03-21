package jnigen

// ReferenceType represents a JNI reference type in the hierarchy.
type ReferenceType struct {
	CType  string  `yaml:"c_type"`
	Parent *string `yaml:"parent"`
}
