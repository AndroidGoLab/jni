package jnigen

// MergedRefType is a reference type with both C and Go names.
type MergedRefType struct {
	CType    string
	GoType   string
	Parent   string
	IsArray  bool
	ElemType string
}
