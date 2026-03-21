package jnigen

// MergedSpec is the fully resolved specification ready for template rendering.
type MergedSpec struct {
	Primitives     []Primitive
	ReferenceTypes []MergedRefType
	OpaqueTypes    []MergedOpaqueType
	Constants      []MergedConstant
	CapiFunctions  []MergedCapiFunc
	EnvMethods     []MergedMethod
	VMMethods      []MergedMethod
	TypeMethods    map[string][]MergedMethod
}
